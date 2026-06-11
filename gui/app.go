package main

import (
	"after_photo/pkg"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	outputBuf      *bytes.Buffer
	outputMu       sync.Mutex
	logFile        *os.File
	logWriter      *logWriter
	startTime      time.Time
	confirmCh      chan *confirmRequest
	processing     bool
	processingMu   sync.Mutex
	cancelFunc     context.CancelFunc
	wg             sync.WaitGroup
}

type confirmRequest struct {
	Message string
	Result  chan bool
}

func NewApp() *App {
	return &App{
		outputBuf: &bytes.Buffer{},
		confirmCh: make(chan *confirmRequest, 1),
		startTime: time.Now(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
	a.wg.Wait()
	if a.logFile != nil {
		a.logFile.Close()
	}
}

func (a *App) GetDefaultDir() string {
	if exePath, err := os.Executable(); err == nil {
		return filepath.Dir(exePath)
	}
	cwd, _ := os.Getwd()
	return cwd
}

func (a *App) PickDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择照片目录",
	})
	if err != nil {
		return ""
	}
	return dir
}

func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

func (a *App) ValidateDirectory(path string) (string, error) {
	if path == "" {
		path = a.GetDefaultDir()
	}
	path = strings.TrimSpace(path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("目录 '%s' 不存在", path)
	}
	if err != nil {
		return "", fmt.Errorf("无法访问目录: %v", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("'%s' 不是一个有效的目录", path)
	}
	return path, nil
}

func (a *App) StartProcessing(dir string, steps [4]bool) error {
	a.processingMu.Lock()
	if a.processing {
		a.processingMu.Unlock()
		return fmt.Errorf("处理正在进行中，请等待完成")
	}
	a.processing = true
	a.processingMu.Unlock()

	validDir, err := a.ValidateDirectory(dir)
	if err != nil {
		a.processingMu.Lock()
		a.processing = false
		a.processingMu.Unlock()
		return err
	}

	a.outputMu.Lock()
	a.outputBuf.Reset()
	a.outputMu.Unlock()

	a.startTime = time.Now()
	logFileName := filepath.Join(validDir, fmt.Sprintf("after_photo_%s.txt", a.startTime.Format("20060102150405")))
	logFile, err := os.Create(logFileName)
	if err == nil {
		a.logFile = logFile
		a.logWriter = &logWriter{file: logFile}
		fmt.Fprintf(logFile, "=== 照片整理工具 - 开始运行 ===\n")
		fmt.Fprintf(logFile, "时间: %s\n", a.startTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(logFile, "工作目录: %s\n", validDir)
		fmt.Fprintf(logFile, "日志文件: %s\n\n", logFileName)
	}

	var writers []io.Writer
	writers = append(writers, a)
	if a.logWriter != nil {
		writers = append(writers, a.logWriter)
	}
	pkg.SetOutput(io.MultiWriter(writers...))

	pkg.SetConfirmFunc(func(message string) bool {
		result := make(chan bool, 1)
		a.confirmCh <- &confirmRequest{Message: message, Result: result}
		runtime.EventsEmit(a.ctx, "confirm-request", message)
		return <-result
	})

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelFunc = cancel
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		defer func() {
			a.processingMu.Lock()
			a.processing = false
			a.processingMu.Unlock()
		}()

		stepFuncs := []func(string){
			pkg.Step1,
			pkg.Step2,
			pkg.Step3,
			pkg.Step4,
		}
		stepNames := []string{
			"步骤1: 按文件类型拆分目录",
			"步骤2: 检测并归类重复照片",
			"步骤3: 在重复照片中选择最佳",
			"步骤4: 删除多余的RAW文件",
		}

		for i, selected := range steps {
			if ctx.Err() != nil {
				a.appendOutput("\n⚠ 已取消\n")
				return
			}
			if selected {
				a.appendOutput(fmt.Sprintf("\n▶ %s\n", stepNames[i]))
				stepFuncs[i](validDir)
			}
		}

		duration := time.Since(a.startTime)
		a.appendOutput(fmt.Sprintf("\n✓ 执行完成！总耗时: %v\n", duration.Round(time.Millisecond)))
		runtime.EventsEmit(a.ctx, "processing-complete", duration.String())
	}()

	return nil
}

func (a *App) ConfirmStep4(confirmed bool) error {
	select {
	case req := <-a.confirmCh:
		req.Result <- confirmed
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("确认请求超时")
	}
}

func (a *App) GetOutput() string {
	a.outputMu.Lock()
	defer a.outputMu.Unlock()
	return a.outputBuf.String()
}

func (a *App) Write(p []byte) (n int, err error) {
	a.outputMu.Lock()
	n, err = a.outputBuf.Write(p)
	a.outputMu.Unlock()

	runtime.EventsEmit(a.ctx, "output", string(p))
	return n, err
}

func (a *App) appendOutput(text string) {
	a.outputMu.Lock()
	a.outputBuf.WriteString(text)
	a.outputMu.Unlock()

	runtime.EventsEmit(a.ctx, "output", text)
}

type logWriter struct {
	file *os.File
	mu   sync.Mutex
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.file == nil {
		return len(p), nil
	}

	cleanText := removeANSICodes(string(p))
	lines := strings.Split(cleanText, "\n")
	for i, line := range lines {
		if line != "" {
			timestamp := time.Now().Format("15:04:05.000")
			if _, err := fmt.Fprintf(lw.file, "[%s] %s", timestamp, line); err != nil {
				return 0, err
			}
		}
		if i < len(lines)-1 {
			if _, err := fmt.Fprintln(lw.file); err != nil {
				return 0, err
			}
		}
	}
	return len(p), nil
}

func removeANSICodes(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) {
				c := s[j]
				if (c >= 0x40 && c <= 0x7E) || c == 'm' || c == 'K' || c == 'H' || c == 'J' || c == 'A' || c == 'B' || c == 'C' || c == 'D' {
					i = j + 1
					break
				}
				j++
			}
			continue
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
