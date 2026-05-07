package main

import (
	"after_photo/pkg"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

// ── Styles ───────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DC4E4")).
			MarginBottom(1)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0AF68"))

	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A"))

	stepActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A")).Bold(true)

	stepDangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A")).Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E"))

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#F7768E")).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DC4E4")).Bold(true)
)

// ── Gradient Banner ──────────────────────────────────────────────────

// gradientStop defines a color at a position (0.0-1.0) along the gradient.
type gradientStop struct {
	pos    float64
	r, g, b uint8
}

// bannerGradient matches the Tokyo Night palette used throughout the app:
//   steel blue → lavender → sakura pink
var bannerGradient = []gradientStop{
	{0.0, 0x7D, 0xC4, 0xE4}, // #7DC4E4 blue
	{0.5, 0xBB, 0x9A, 0xF7}, // #BB9AF7 purple
	{1.0, 0xF7, 0x76, 0x8E}, // #F7768E pink
}

// sampleGradient returns the RGB color at position t (0.0-1.0) along the gradient.
func sampleGradient(t float64) (uint8, uint8, uint8) {
	stops := bannerGradient
	if t <= stops[0].pos {
		return stops[0].r, stops[0].g, stops[0].b
	}
	if t >= stops[len(stops)-1].pos {
		s := stops[len(stops)-1]
		return s.r, s.g, s.b
	}
	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].pos && t <= stops[i+1].pos {
			local := (t - stops[i].pos) / (stops[i+1].pos - stops[i].pos)
			return lerp8(stops[i].r, stops[i+1].r, local),
				lerp8(stops[i].g, stops[i+1].g, local),
				lerp8(stops[i].b, stops[i+1].b, local)
		}
	}
	return stops[0].r, stops[0].g, stops[0].b
}

func lerp8(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + t*float64(b-a))
}

// renderBanner renders text using a figlet font with a vertical gradient.
func renderBanner(text string) string {
	myFigure := figure.NewFigure(text, "slant", true)
	lines := strings.Split(myFigure.String(), "\n")

	// Count non-empty lines for gradient mapping
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		return ""
	}

	var b strings.Builder
	idx := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		t := float64(idx) / float64(nonEmpty-1)
		r, g, bl := sampleGradient(t)
		color := fmt.Sprintf("#%02X%02X%02X", r, g, bl)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		b.WriteString("  ")
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		idx++
	}
	return b.String()
}

// ── App States ───────────────────────────────────────────────────

type appState int

const (
	stateInputDir appState = iota
	stateSelectSteps
	stateRunning
	stateConfirm
	stateDone
)

// ── Messages ─────────────────────────────────────────────────────

type outputMsg string
type allDoneMsg struct{ duration time.Duration }
type confirmRequestMsg struct {
	req *pkg.ConfirmRequest
}

// ── Channel Writer ───────────────────────────────────────────────

// channelWriter captures io.Writer output and sends lines to a channel
type channelWriter struct {
	ch  chan string
	buf bytes.Buffer
	mu  sync.Mutex
}

func newChannelWriter(ch chan string) *channelWriter {
	return &channelWriter{ch: ch}
}

func (cw *channelWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.buf.Write(p)
	for {
		line, err := cw.buf.ReadString('\n')
		if err != nil {
			cw.buf.WriteString(line)
			break
		}
		cw.ch <- strings.TrimRight(line, "\n\r")
	}
	return len(p), nil
}

func (cw *channelWriter) Flush() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.buf.Len() > 0 {
		cw.ch <- cw.buf.String()
		cw.buf.Reset()
	}
}

// ── Log Writer ───────────────────────────────────────────────────

type logWriter struct {
	file *os.File
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	if lw.file != nil {
		cleanText := removeANSICodes(string(p))
		lines := strings.Split(cleanText, "\n")
		for i, line := range lines {
			if line != "" {
				now := time.Now()
				timestamp := now.Format("15:04:05.000")
				lw.file.WriteString(fmt.Sprintf("[%s] %s", timestamp, line))
			}
			if i < len(lines)-1 {
				lw.file.WriteString("\n")
			}
		}
	}
	return len(p), nil
}

// ── Model ────────────────────────────────────────────────────────

type model struct {
	state        appState
	dirInput     textinput.Model
	viewport     viewport.Model
	currentDir   string
	steps        [4]bool
	cursor       int
	outputLines  []string
	outputCh     chan string
	logFile      *os.File
	logWriter    *logWriter
	channelW     *channelWriter
	startTime    time.Time
	ready        bool
	err          string
	windowWidth  int
	windowHeight int
	confirmReq   *pkg.ConfirmRequest
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "输入照片目录路径（留空使用程序所在目录）"
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	return model{
		state:     stateInputDir,
		dirInput:  ti,
		steps:     [4]bool{true, true, true, false},
		outputCh:  make(chan string, 200),
		startTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		headerH := 3
		footerH := 2
		vpHeight := msg.Height - headerH - footerH
		if vpHeight < 5 {
			vpHeight = 5
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, vpHeight)
			m.viewport.Style = lipgloss.NewStyle().PaddingLeft(1)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = vpHeight
		}
		return m, nil

	case outputMsg:
		m.outputLines = append(m.outputLines, string(msg))
		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.viewport.GotoBottom()
		return m, waitForOutput(m.outputCh)

	case confirmRequestMsg:
		m.confirmReq = msg.req
		m.state = stateConfirm
		return m, nil

	case allDoneMsg:
		m.state = stateDone
		doneLine := fmt.Sprintf("\n✓ 执行完成！总耗时: %v", msg.duration.Round(time.Millisecond))
		m.outputLines = append(m.outputLines, doneLine)
		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.viewport.GotoBottom()
		if m.channelW != nil {
			m.channelW.Flush()
		}
		return m, nil
	}

	// State-specific key handling
	var cmd tea.Cmd
	switch m.state {
	case stateInputDir:
		m, cmd = m.updateInputDir(msg)
	case stateSelectSteps:
		m, cmd = m.updateSelectSteps(msg)
	case stateRunning:
		m.viewport, cmd = m.viewport.Update(msg)
	case stateConfirm:
		m, cmd = m.updateConfirm(msg)
	case stateDone:
		m, cmd = m.updateDone(msg)
	default:
		cmd = nil
	}

	return m, cmd
}

func (m model) updateInputDir(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			photoDir := strings.TrimSpace(m.dirInput.Value())
			if photoDir == "" {
				if exePath, err := os.Executable(); err == nil {
					photoDir = filepath.Dir(exePath)
				} else {
					m.err = "无法获取程序路径"
					return m, nil
				}
			}
			if _, err := os.Stat(photoDir); os.IsNotExist(err) {
				m.err = fmt.Sprintf("目录 '%s' 不存在", photoDir)
				return m, nil
			}
			info, err := os.Stat(photoDir)
			if err != nil || !info.IsDir() {
				m.err = fmt.Sprintf("'%s' 不是一个有效的目录", photoDir)
				return m, nil
			}
			m.currentDir = photoDir
			m.err = ""

			// Create log file
			logFileName := filepath.Join(m.currentDir, fmt.Sprintf("after_photo_%s.txt", m.startTime.Format("20060102150405")))
			logFile, err := os.Create(logFileName)
			if err == nil {
				m.logFile = logFile
				m.logWriter = &logWriter{file: logFile}
				logFile.WriteString(fmt.Sprintf("=== 照片整理工具 - 开始运行 ===\n"))
				logFile.WriteString(fmt.Sprintf("时间: %s\n", m.startTime.Format("2006-01-02 15:04:05")))
				logFile.WriteString(fmt.Sprintf("工作目录: %s\n", m.currentDir))
				logFile.WriteString(fmt.Sprintf("日志文件: %s\n\n", logFileName))
			}

			m.state = stateSelectSteps
			m.cursor = 0
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.dirInput, cmd = m.dirInput.Update(msg)
	return m, cmd
}

func (m model) updateSelectSteps(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 3 {
				m.cursor++
			}
		case " ":
			m.steps[m.cursor] = !m.steps[m.cursor]
		case "a":
			allChecked := m.steps[0] && m.steps[1] && m.steps[2] && m.steps[3]
			for i := range m.steps {
				m.steps[i] = !allChecked
			}
		case "enter":
			hasStep := false
			for _, s := range m.steps {
				if s {
					hasStep = true
					break
				}
			}
			if !hasStep {
				m.err = "请至少选择一个步骤"
				return m, nil
			}
			m.err = ""
			m.state = stateRunning
			m.outputLines = nil
			cmd := m.buildExecutionCmd()
			return m, cmd
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) updateConfirm(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.confirmReq != nil {
				m.confirmReq.Result <- true
				m.confirmReq = nil
			}
			m.state = stateRunning
			return m, tea.Batch(waitForOutput(m.outputCh), waitForConfirmRequest(pkg.ConfirmCh))
		case "n", "N", "esc":
			if m.confirmReq != nil {
				m.confirmReq.Result <- false
				m.confirmReq = nil
			}
			m.state = stateRunning
			return m, tea.Batch(waitForOutput(m.outputCh), waitForConfirmRequest(pkg.ConfirmCh))
		}
	}
	return m, nil
}

func (m model) updateDone(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.state = stateSelectSteps
			m.outputLines = nil
			m.viewport.SetContent("")
			m.err = ""
			return m, nil
		case "2":
			m.state = stateInputDir
			m.currentDir = ""
			m.outputLines = nil
			m.viewport.SetContent("")
			m.err = ""
			m.dirInput.SetValue("")
			m.dirInput.Focus()
			if m.logFile != nil {
				m.logFile.Close()
				m.logFile = nil
			}
			return m, textinput.Blink
		case "3", "enter", "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ── Execution ────────────────────────────────────────────────────

func (m model) buildExecutionCmd() tea.Cmd {
	// Set up channel writer for TUI output
	m.channelW = newChannelWriter(m.outputCh)

	// Multi-writer: channel (for TUI) + log file
	var writers []io.Writer
	writers = append(writers, m.channelW)
	if m.logWriter != nil {
		writers = append(writers, m.logWriter)
	}
	pkg.SetOutput(io.MultiWriter(writers...))

	// Set up confirm channel for step4
	pkg.ConfirmCh = make(chan *pkg.ConfirmRequest, 1)

	// Capture execution params
	steps := m.steps
	currentDir := m.currentDir
	startTime := m.startTime
	outputCh := m.outputCh
	confirmCh := pkg.ConfirmCh
	channelW := m.channelW

	return tea.Batch(
		waitForOutput(outputCh),
		waitForConfirmRequest(confirmCh),
		executeSteps(steps, currentDir, startTime, outputCh, channelW),
	)
}

func executeSteps(steps [4]bool, currentDir string, startTime time.Time, outputCh chan string, channelW *channelWriter) tea.Cmd {
	return func() tea.Msg {
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
			if selected {
				outputCh <- fmt.Sprintf("\n▶ %s", stepNames[i])
				stepFuncs[i](currentDir)
			}
		}

		channelW.Flush()
		return allDoneMsg{duration: time.Since(startTime)}
	}
}

func waitForOutput(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return outputMsg(line)
	}
}

func waitForConfirmRequest(ch chan *pkg.ConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-ch
		if !ok {
			return nil
		}
		return confirmRequestMsg{req: req}
	}
}

// ── View ─────────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "\n  初始化中..."
	}

	switch m.state {
	case stateInputDir:
		return m.viewInputDir()
	case stateSelectSteps:
		return m.viewSelectSteps()
	case stateRunning, stateConfirm:
		return m.viewRunning()
	case stateDone:
		return m.viewDone()
	}
	return ""
}

func (m model) viewInputDir() string {
	var b strings.Builder

	b.WriteString(renderBanner("AFTER"))
	b.WriteString(renderBanner("PHOTO"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  v1.0 · 照片整理工具"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  按 JPEG/RAW/视频 分类照片，检测重复，保留最佳"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  支持 JPG·CR2·NEF·ARW·MP4·MOV 等常见格式"))
	b.WriteString("\n\n")

	b.WriteString("  照片目录:\n")
	b.WriteString("  ")
	b.WriteString(m.dirInput.View())
	b.WriteString("\n\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ %s", m.err)))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  Enter 确认 · Ctrl+C 退出"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewSelectSteps() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  After Photo v1.0"))
	b.WriteString("  ")
	b.WriteString(dirStyle.Render(m.currentDir))
	b.WriteString("\n\n")

	b.WriteString("  选择要执行的步骤:\n\n")

	stepNames := []string{
		"按文件类型拆分目录 (JPG→jpg/ RAW→raw/ 视频→video/)",
		"检测并归类重复照片",
		"在重复照片中选择最佳",
		"删除多余的RAW文件（无对应JPG的RAW文件）",
	}

	for i, name := range stepNames {
		cursor := " "
		if m.cursor == i {
			cursor = "▸"
		}

		checked := "○"
		if m.steps[i] {
			checked = "●"
		}

		style := stepStyle
		if m.cursor == i {
			style = stepActiveStyle
		}
		if i == 3 {
			style = stepDangerStyle
		}

		b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checked, style.Render(name)))
	}

	b.WriteString("\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ %s", m.err)))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  ↑↓ 选择 · 空格 切换 · a 全选/全不选 · Enter 执行 · q 退出"))
	b.WriteString("\n")

	return b.String()
}

func (m model) viewRunning() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  After Photo v1.0"))
	b.WriteString("  ")
	b.WriteString(dirStyle.Render(m.currentDir))

	if m.state == stateConfirm {
		status := lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")).Bold(true).Render("⚠ 等待确认")
		b.WriteString("  ")
		b.WriteString(status)
	} else {
		status := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AF68")).Bold(true).Render("⏳ 执行中")
		b.WriteString("  ")
		b.WriteString(status)
	}
	b.WriteString("\n")

	// Output viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Confirm dialog overlay
	if m.state == stateConfirm && m.confirmReq != nil {
		b.WriteString("\n")
		b.WriteString(confirmStyle.Render(fmt.Sprintf("  ⚠ 确认操作\n\n  %s\n\n  Y 确认 · N/ESC 取消", m.confirmReq.Message)))
		b.WriteString("\n")
	} else {
		b.WriteString(helpStyle.Render("  ↑↓ 滚动 · 等待执行完成..."))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) viewDone() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  After Photo v1.0"))
	b.WriteString("  ")
	b.WriteString(successStyle.Render("✓ 完成"))
	b.WriteString("\n")

	b.WriteString(m.viewport.View())
	b.WriteString("\n\n")

	options := []string{
		"继续其他步骤",
		"整理其他文件夹",
		"退出程序",
	}

	for i, opt := range options {
		b.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, opt))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  1/2/3 选择 · Enter/3 退出"))
	b.WriteString("\n")

	return b.String()
}

// ── Main ─────────────────────────────────────────────────────────

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("运行出错: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	if m.logFile != nil {
		m.logFile.Close()
	}
}

// removeANSICodes 去除字符串中的ANSI颜色代码
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
