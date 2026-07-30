package main

import (
	"after_photo/pkg"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 推送到前端的 Wails 事件名，前端通过 window.runtime.EventsOn 监听
const (
	eventLog      = "log"      // string：pkg 的实时输出（含 ANSI 颜色码，由前端渲染）
	eventProgress = "progress" // progressPayload：步骤内的处理进度
	eventStep     = "step"     // stepPayload：某个步骤开始执行
	eventConfirm  = "confirm"  // string：step4 删除确认请求
	eventDone     = "done"     // donePayload：全部步骤结束
)

type stepPayload struct {
	Index int    `json:"index"` // 0-3
	Total int    `json:"total"` // 本次选中的步骤总数
	Name  string `json:"name"`
}

type progressPayload struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Message string `json:"message"`
}

type donePayload struct {
	Duration  string `json:"duration"`
	Cancelled bool   `json:"cancelled"`
}

// DirStats 是 ScanDirectory 的返回结果，用于前端展示目录概况
type DirStats struct {
	JPG       int   `json:"jpg"`
	RAW       int   `json:"raw"`
	Video     int   `json:"video"`
	Other     int   `json:"other"`
	TotalSize int64 `json:"totalSize"` // 字节
}

// 扫描时使用的扩展名集合，与 pkg 中未导出的 rawExtensions / videoExtensions 保持一致
var (
	jpegExts = map[string]bool{".JPG": true, ".JPEG": true}
	rawExts  = map[string]bool{
		".RAF": true, ".CR2": true, ".CR3": true, ".NEF": true, ".NRW": true,
		".ARW": true, ".DNG": true, ".ORF": true, ".RW2": true, ".PEF": true,
		".SRW": true, ".MRW": true, ".3FR": true, ".FFF": true, ".IIQ": true,
		".KDC": true, ".MDC": true, ".MOS": true, ".MEF": true, ".X3F": true,
	}
	videoExts = map[string]bool{
		".MP4": true, ".MOV": true, ".AVI": true, ".MKV": true, ".MTS": true, ".M2TS": true,
	}
)

var stepDefs = []struct {
	Name string
	Func func(string)
}{
	{"步骤 1 · 按类型拆分目录", pkg.Step1},
	{"步骤 2 · 检测并归类重复照片", pkg.Step2},
	{"步骤 3 · 在重复照片中选择最佳", pkg.Step3},
	{"步骤 4 · 删除多余的 RAW 文件", pkg.Step4},
}

// App 是 Wails 绑定到前端的后端对象，所有导出方法可在 JS 中通过 window.go.main.App 调用
type App struct {
	ctx context.Context

	mu         sync.Mutex
	processing bool
	cancel     context.CancelFunc
	confirmRes chan bool // 非 nil 表示有待回应的确认请求
	wg         sync.WaitGroup

	logFile *os.File
	logw    *logWriter
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(_ context.Context) {
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	res := a.confirmRes
	a.mu.Unlock()
	if res != nil {
		res <- false // 唤醒可能阻塞在确认上的 goroutine
	}
	// pkg 步骤本身不可中断，长时间步骤（如大批量哈希计算）不能卡住退出流程，
	// 限时等待后放弃：进程随即退出，goroutine 由系统回收
	done := make(chan struct{})
	go func() { a.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	if a.logFile != nil {
		a.logFile.Close()
	}
}

// GetDefaultDir 返回可执行文件所在目录，作为默认工作目录
func (a *App) GetDefaultDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	cwd, _ := os.Getwd()
	return cwd
}

// PickDirectory 弹出系统目录选择框，取消时返回空字符串
func (a *App) PickDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择照片目录"})
	if err != nil {
		return ""
	}
	return dir
}

// ScanDirectory 校验目录并统计其中各类文件的数量与体积
func (a *App) ScanDirectory(dir string) (*DirStats, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = a.GetDefaultDir()
	}
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在：%s", dir)
	}
	if err != nil {
		return nil, fmt.Errorf("无法访问目录：%v", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("不是一个目录：%s", dir)
	}

	stats := &DirStats{}
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过无法读取的条目
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") { // 跳过隐藏文件与目录
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToUpper(filepath.Ext(name))
		switch {
		case jpegExts[ext]:
			stats.JPG++
		case rawExts[ext]:
			stats.RAW++
		case videoExts[ext]:
			stats.Video++
		default:
			stats.Other++
		}
		if fi, err := d.Info(); err == nil {
			stats.TotalSize += fi.Size()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("扫描目录失败：%v", err)
	}
	return stats, nil
}

// StartProcessing 在后台 goroutine 中依次执行选中的步骤，期间通过事件推送日志与进度
func (a *App) StartProcessing(dir string, steps [4]bool) error {
	a.mu.Lock()
	if a.processing {
		a.mu.Unlock()
		return fmt.Errorf("已有任务正在执行")
	}
	a.mu.Unlock()

	stats, err := a.ScanDirectory(dir)
	if err != nil {
		return err
	}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = a.GetDefaultDir()
	}

	selected := 0
	for _, s := range steps {
		if s {
			selected++
		}
	}
	if selected == 0 {
		return fmt.Errorf("请至少选择一个步骤")
	}

	// 日志文件创建在工作目录下，与 TUI 版本一致；先关闭上一轮遗留的句柄，避免泄漏
	startTime := time.Now()
	a.mu.Lock()
	if a.logFile != nil {
		a.logFile.Close()
		a.logFile = nil
		a.logw = nil
	}
	a.mu.Unlock()
	logPath := filepath.Join(dir, fmt.Sprintf("after_photo_%s.txt", startTime.Format("20060102150405")))
	if f, err := os.Create(logPath); err == nil {
		a.mu.Lock()
		a.logFile = f
		a.logw = &logWriter{file: f}
		a.mu.Unlock()
		fmt.Fprintf(f, "=== After Photo 运行日志 ===\n时间: %s\n目录: %s\n\n", startTime.Format("2006-01-02 15:04:05"), dir)
	} else {
		a.mu.Lock()
		a.logw = nil
		a.mu.Unlock()
		logPath = ""
	}

	pkg.SetOutput(a) // App 实现 io.Writer，把输出转发为 log 事件
	pkg.SetProgressFunc(func(current, total int, message string) {
		a.emit(eventProgress, progressPayload{Current: current, Total: total, Message: message})
	})
	pkg.SetConfirmFunc(func(message string) bool {
		res := make(chan bool, 1)
		a.mu.Lock()
		a.confirmRes = res
		a.mu.Unlock()
		a.emit(eventConfirm, message)
		select {
		case ok := <-res:
			return ok
		case <-a.ctx.Done(): // 应用退出时不再阻塞
			return false
		}
	})

	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.processing = true
	a.cancel = cancel
	a.mu.Unlock()
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		cancelled := false
		defer func() {
			if r := recover(); r != nil {
				a.appendLog(fmt.Sprintf("\n✗ 发生内部错误：%v\n", r))
			}
			a.mu.Lock()
			a.processing = false
			a.cancel = nil
			a.mu.Unlock()
			a.emit(eventDone, donePayload{
				Duration:  time.Since(startTime).Round(time.Millisecond).String(),
				Cancelled: cancelled,
			})
		}()

		a.appendLog(fmt.Sprintf("工作目录：%s\n", dir))
		a.appendLog(fmt.Sprintf("文件概况：JPG %d · RAW %d · 视频 %d\n", stats.JPG, stats.RAW, stats.Video))
		if logPath != "" {
			a.appendLog(fmt.Sprintf("日志文件：%s\n", logPath))
		}

		order := 0
		for i, sel := range steps {
			if !sel {
				continue
			}
			if ctx.Err() != nil {
				cancelled = true
				a.appendLog("\n⚠ 已取消\n")
				return
			}
			order++
			a.emit(eventStep, stepPayload{Index: i, Total: selected, Name: stepDefs[i].Name})
			a.appendLog(fmt.Sprintf("\n▶ %s\n", stepDefs[i].Name))
			stepDefs[i].Func(dir)
		}
	}()

	return nil
}

// CancelProcessing 请求取消：当前步骤结束后不再执行后续步骤
func (a *App) CancelProcessing() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// ConfirmStep4 回应前端的确认对话框（step4 删除 RAW 前触发）
func (a *App) ConfirmStep4(confirmed bool) {
	a.mu.Lock()
	res := a.confirmRes
	a.confirmRes = nil
	a.mu.Unlock()
	if res != nil {
		res <- confirmed
	}
}

// Quit 退出应用
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// Write 实现 io.Writer：pkg 的每一行输出都实时推送到前端，同时写入日志文件
func (a *App) Write(p []byte) (int, error) {
	a.emit(eventLog, string(p))
	if a.logw != nil {
		a.logw.Write(p)
	}
	return len(p), nil
}

func (a *App) appendLog(text string) {
	a.Write([]byte(text))
}

func (a *App) emit(event string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data)
	}
}

// logWriter 将输出写入日志文件：去除 ANSI 颜色码并添加时间戳前缀
type logWriter struct {
	file *os.File
	mu   sync.Mutex
}

func (w *logWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return len(p), nil
	}
	clean := stripANSI(string(p))
	lines := strings.Split(clean, "\n")
	for i, line := range lines {
		if line != "" {
			ts := time.Now().Format("15:04:05.000")
			if _, err := fmt.Fprintf(w.file, "[%s] %s", ts, line); err != nil {
				return 0, err
			}
		}
		if i < len(lines)-1 {
			if _, err := fmt.Fprintln(w.file); err != nil {
				return 0, err
			}
		}
	}
	return len(p), nil
}

// stripANSI 去除 ANSI 转义序列
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !(s[j] >= 0x40 && s[j] <= 0x7e) {
				j++
			}
			if j < len(s) {
				i = j + 1
			} else {
				i = j
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
