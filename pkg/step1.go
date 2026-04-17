package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 按文件类型拆分目录（内部函数）
func step1(photoDir string) {
	jpgDir := filepath.Join(photoDir, "jpg")
	rawDir := filepath.Join(photoDir, "raw")
	videoDir := filepath.Join(photoDir, "video")

	dirsToCreate := []string{}
	dirsToCreate = append(dirsToCreate, jpgDir)
	dirsToCreate = append(dirsToCreate, rawDir)
	dirsToCreate = append(dirsToCreate, videoDir)

	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(out, "创建目录失败 %s: %v\n", dir, err)
			return
		}
	}

	entries, err := os.ReadDir(photoDir)
	if err != nil {
		fmt.Fprintf(out, "读取目录失败: %v\n", err)
		return
	}

	movedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(photoDir, entry.Name())
		ext := strings.ToUpper(filepath.Ext(filePath))

		var targetDir string
		if ext == ".JPG" || ext == ".JPEG" {
			targetDir = jpgDir
		} else if rawExtensions[ext] {
			targetDir = rawDir
		} else if videoExtensions[ext] {
			targetDir = videoDir
		} else {
			continue
		}

		destPath := filepath.Join(targetDir, entry.Name())
		if err := os.Rename(filePath, destPath); err != nil {
			fmt.Fprintf(out, "移动文件失败 %s: %v\n", entry.Name(), err)
		} else {
			fmt.Fprintf(out, "  移动: %s -> %s/\n", entry.Name(), filepath.Base(targetDir))
			movedCount++
		}
	}

	// 删除空的video目录（如果没有视频文件）
	if entries, err := os.ReadDir(videoDir); err == nil && len(entries) == 0 {
		os.Remove(videoDir)
	}

	fmt.Fprintf(out, "\n%s步骤1完成！%s 共移动 %d 个文件\n", ColorGreen, ColorReset, movedCount)
}