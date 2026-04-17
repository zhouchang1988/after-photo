package main

import (
	"after_photo/pkg"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// removeANSICodes 去除字符串中的ANSI颜色代码
func removeANSICodes(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		// 检测 ANSI 转义序列的开始
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// 跳过整个转义序列，直到找到结束字符 (m, K, H 等)
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

// logWriter 同时写入屏幕和日志文件
type logWriter struct {
	file      *os.File
	startTime time.Time
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	// 写入屏幕
	n, err = os.Stdout.Write(p)
	if err != nil {
		return
	}
	// 写入日志文件（去除ANSI颜色代码，并为每行添加时间戳）
	if lw.file != nil {
		cleanText := removeANSICodes(string(p))
		lines := strings.Split(cleanText, "\n")
		for i, line := range lines {
			if line != "" {
				// 获取当前时间
				now := time.Now()
				timestamp := now.Format("15:04:05.000")
				lw.file.WriteString(fmt.Sprintf("[%s] %s", timestamp, line))
			}
			if i < len(lines)-1 {
				lw.file.WriteString("\n")
			}
		}
	}
	return
}

// 全局变量，用于输出
var out io.Writer

func main() {
	startTime := time.Now()

	// 创建自定义writer，但先不设置日志文件
	lw := &logWriter{
		file:      nil,
		startTime: startTime,
	}

	// 设置全局输出（此时只输出到屏幕）
	out = lw
	pkg.SetOutput(lw)

	// 先输出到屏幕（清屏）
	fmt.Print("\033[H\033[2J")

	// 使用 logWriter 输出标题
	fmt.Fprintf(out, "\n%s%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", pkg.ColorCyan, pkg.ColorBold, pkg.ColorReset)
	fmt.Fprintf(out, "%s%s  📸 照片整理工具 - After Photo v1.0%s\n", pkg.ColorCyan, pkg.ColorBold, pkg.ColorReset)
	fmt.Fprintf(out, "%s%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", pkg.ColorCyan, pkg.ColorBold, pkg.ColorReset)

	scanner := bufio.NewScanner(os.Stdin)
	currentDir := ""

	for {
		if currentDir == "" {
			fmt.Fprintf(out, "%s请输入照片所在目录路径 %s(直接回车使用程序所在目录): %s", pkg.ColorCyan, pkg.ColorYellow, pkg.ColorReset)
			if !scanner.Scan() {
				fmt.Fprintf(out, "读取输入失败\n")
				return
			}
			photoDir := strings.TrimSpace(scanner.Text())

			// 如果输入为空，使用程序所在目录
			if photoDir == "" {
				if exePath, err := os.Executable(); err == nil {
					photoDir = filepath.Dir(exePath)
					fmt.Fprintf(out, "%s使用程序所在目录: %s%s\n", pkg.ColorGreen, photoDir, pkg.ColorReset)
				} else {
					fmt.Fprintf(out, "%s错误: 无法获取程序路径: %v%s\n", pkg.ColorRed, err, pkg.ColorReset)
					continue
				}
			}

			if _, err := os.Stat(photoDir); os.IsNotExist(err) {
				fmt.Fprintf(out, "%s错误: 目录 '%s' 不存在%s\n", pkg.ColorRed, photoDir, pkg.ColorReset)
				continue
			}

			info, err := os.Stat(photoDir)
			if err != nil {
				fmt.Fprintf(out, "%s错误: 无法访问目录 '%s': %v%s\n", pkg.ColorRed, photoDir, err, pkg.ColorReset)
				continue
			}
			if !info.IsDir() {
				fmt.Fprintf(out, "%s错误: '%s' 不是一个目录%s\n", pkg.ColorRed, photoDir, pkg.ColorReset)
				continue
			}

			currentDir = photoDir

			// 确认工作目录后，创建日志文件并保存在工作目录中
			logFileName := filepath.Join(currentDir, fmt.Sprintf("after_photo_%s.txt", startTime.Format("20060102150405")))
			logFile, err := os.Create(logFileName)
			if err != nil {
				fmt.Printf("无法创建日志文件: %v\n", err)
				return
			}
			defer logFile.Close()

			// 设置日志文件
			lw.file = logFile

			// 写入日志文件头部
			logFile.WriteString(fmt.Sprintf("=== 照片整理工具 - 开始运行 ===\n"))
			logFile.WriteString(fmt.Sprintf("时间: %s\n", startTime.Format("2006-01-02 15:04:05")))
			logFile.WriteString(fmt.Sprintf("工作目录: %s\n", currentDir))
			logFile.WriteString(fmt.Sprintf("日志文件: %s\n\n", logFileName))

			fmt.Fprintf(out, "%s日志文件: %s%s\n\n", pkg.ColorGreen, logFileName, pkg.ColorReset)
		}

		fmt.Fprintf(out, "\n%s当前目录: %s%s\n\n", pkg.ColorYellow, currentDir, pkg.ColorReset)
		fmt.Fprintf(out, "%s%s=== 可执行的步骤 ===%s\n", pkg.ColorCyan, pkg.ColorBold, pkg.ColorReset)
		fmt.Fprintf(out, "%s[1]%s 按文件类型拆分目录 (JPG -> jpg/, RAW -> raw/, 视频 -> video/)%s\n", pkg.ColorGreen, pkg.ColorReset, pkg.ColorReset)
		fmt.Fprintf(out, "%s[2]%s 检测并归类重复照片%s\n", pkg.ColorGreen, pkg.ColorReset, pkg.ColorReset)
		fmt.Fprintf(out, "%s[3]%s 在重复照片中选择最佳%s\n", pkg.ColorGreen, pkg.ColorReset, pkg.ColorReset)
		fmt.Fprintf(out, "%s[4]%s 删除多余的RAW文件（无对应JPG的RAW文件）%s\n", pkg.ColorRed, pkg.ColorReset, pkg.ColorReset)
		fmt.Fprintf(out, "\n%s请输入要执行的步骤 %s(不输入则执行1-3，可组合如'123'): %s", pkg.ColorCyan, pkg.ColorWhite, pkg.ColorReset)

		var input string
		if !scanner.Scan() {
			break
		}
		input = strings.TrimSpace(scanner.Text())

		if input == "" {
			fmt.Fprintf(out, "\n%s%s[执行中]%s 正在执行步骤1-3...\n", pkg.ColorYellow, pkg.ColorBold, pkg.ColorReset)
			pkg.Step1(currentDir)
			pkg.Step2(currentDir)
			pkg.Step3(currentDir)
		} else {
			for _, step := range input {
				switch step {
				case '1':
					fmt.Fprintf(out, "\n%s%s[执行中]%s 正在执行步骤1...\n", pkg.ColorYellow, pkg.ColorBold, pkg.ColorReset)
					pkg.Step1(currentDir)
				case '2':
					fmt.Fprintf(out, "\n%s%s[执行中]%s 正在执行步骤2...\n", pkg.ColorYellow, pkg.ColorBold, pkg.ColorReset)
					pkg.Step2(currentDir)
				case '3':
					fmt.Fprintf(out, "\n%s%s[执行中]%s 正在执行步骤3...\n", pkg.ColorYellow, pkg.ColorBold, pkg.ColorReset)
					pkg.Step3(currentDir)
				case '4':
					fmt.Fprintf(out, "\n%s%s[执行中]%s 正在执行步骤4...\n", pkg.ColorYellow, pkg.ColorBold, pkg.ColorReset)
					pkg.Step4(currentDir)
				default:
					fmt.Fprintf(out, "%s无效的步骤: %c%s\n", pkg.ColorRed, step, pkg.ColorReset)
				}
			}
		}

		// 执行完成后显示菜单
		fmt.Fprintf(os.Stdout, "\n%s%s✓ 执行完成！总耗时: %v%s\n\n", pkg.ColorGreen, pkg.ColorBold, time.Since(startTime).Round(time.Millisecond), pkg.ColorReset)
		fmt.Fprintf(os.Stdout, "%s请选择下一步操作:%s\n", pkg.ColorCyan, pkg.ColorReset)
		fmt.Fprintf(os.Stdout, "  %s[1]%s 继续其他步骤\n", pkg.ColorGreen, pkg.ColorReset)
		fmt.Fprintf(os.Stdout, "  %s[2]%s 整理其他文件夹\n", pkg.ColorGreen, pkg.ColorReset)
		fmt.Fprintf(os.Stdout, "  %s[3]%s 退出程序 (直接回车也可以退出)\n", pkg.ColorGreen, pkg.ColorReset)
		fmt.Fprintf(os.Stdout, "\n%s请输入选项 (1/2/3, 或直接回车退出): %s", pkg.ColorCyan, pkg.ColorReset)

		var choice string
		if !scanner.Scan() {
			break
		}
		choice = strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			continue
		case "2":
			currentDir = ""
		case "3", "":
			fmt.Fprintf(out, "\n%s%s感谢使用照片整理工具！%s\n", pkg.ColorGreen, pkg.ColorBold, pkg.ColorReset)
			fmt.Fprintf(out, "%s再见！👋%s\n", pkg.ColorCyan, pkg.ColorReset)
			return
		default:
			fmt.Fprintf(out, "%s无效的选项: %s，继续执行其他步骤%s\n", pkg.ColorYellow, choice, pkg.ColorReset)
		}
	}
}