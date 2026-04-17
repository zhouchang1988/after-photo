package pkg

import "io"

// 全局输出变量，用于同时输出到屏幕和日志文件
var out io.Writer

// SetOutput 设置全局输出
func SetOutput(w io.Writer) {
	out = w
}

// ANSI颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
	ColorWhite  = "\033[37m"
)

// 步骤1：按文件类型拆分目录
func Step1(photoDir string) {
	step1(photoDir)
}

// 步骤2：重复照片归类
func Step2(photoDir string) {
	step2(photoDir)
}

// 步骤3：在重复照片中选择最佳
func Step3(photoDir string) {
	step3(photoDir)
}

// 步骤4：删除多余的RAW文件
func Step4(photoDir string) {
	step4(photoDir)
}

// 常见的RAW文件扩展名
var rawExtensions = map[string]bool{
	".RAF":  true, // Fuji
	".CR2":  true, // Canon
	".CR3":  true, // Canon
	".NEF":  true, // Nikon
	".NRW":  true, // Nikon
	".ARW":  true, // Sony
	".DNG":  true, // Adobe Digital Negative
	".ORF":  true, // Olympus
	".RW2":  true, // Panasonic
	".PEF":  true, // Pentax
	".SRW":  true, // Samsung
	".MRW":  true, // Minolta
	".3FR":  true, // Hasselblad
	".FFF":  true, // Hasselblad
	".IIQ":  true, // Phase One
	".KDC":  true, // Kodak
	".MDC":  true, // Minolta
	".MOS":  true, // Leaf
	".MEF":  true, // Mamiya
	".X3F":  true, // Sigma
}

// 常见的视频文件扩展名
var videoExtensions = map[string]bool{
	".MP4":  true,
	".MOV":  true,
	".AVI":  true,
	".MKV":  true,
	".MTS":  true,
	".M2TS": true,
}

// 图片相似度阈值，用于判断两张图片是否相似
const SimilarityThreshold = 17