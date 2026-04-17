package pkg

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// 在重复照片中挑选最佳（内部函数）
func step3(photoDir string) {
	jpgDir := filepath.Join(photoDir, "jpg")
	rawDir := filepath.Join(photoDir, "raw")

	if _, err := os.Stat(jpgDir); os.IsNotExist(err) {
		fmt.Fprintf(out, "错误: jpg目录不存在，请先执行步骤1\n")
		return
	}

	fmt.Fprintf(out, "\n处理JPG目录，挑选最佳照片...\n")
	selectBestInGroups(jpgDir, "JPG")

	// 检查并删除空的 JPG 目录
	if removeEmptyDir(jpgDir, "JPG") {
		fmt.Fprintf(out, "✓ JPG目录已删除（无图片文件）\n")
	}

	if _, err := os.Stat(rawDir); err == nil {
		fmt.Fprintf(out, "\n处理RAW目录，同步标记最佳照片...\n")
		keepRawByJpgSelection(jpgDir, rawDir)

		// 检查并删除空的 RAW 目录
		if removeEmptyDir(rawDir, "RAW") {
			fmt.Fprintf(out, "✓ RAW目录已删除（无图片文件）\n")
		}
	}
}

// 在分组中选择最佳照片
func selectBestInGroups(dirPath string, fileType string) {
	groups := make(map[string][]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if path == dirPath {
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}

		groupName := filepath.Base(path)

		var files []string
		targetExts := map[string]bool{".JPG": true, ".JPEG": true}
		if fileType == "RAW" {
			targetExts = rawExtensions
		}

		err = filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if subInfo.IsDir() {
				return nil
			}

			ext := strings.ToUpper(filepath.Ext(subPath))
			if targetExts[ext] {
				files = append(files, subPath)
			}

			return nil
		})

		if err != nil {
			return err
		}

		if len(files) > 0 {
			groups[groupName] = files
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "扫描分组失败: %v\n", err)
		return
	}

	if len(groups) == 0 {
		fmt.Fprintf(out, "没有找到分组\n")
		return
	}

	markedCount := 0

	for _, files := range groups {
		if len(files) <= 1 {
			continue
		}

		bestFile := files[0]
		bestScore := 0.0

		for _, file := range files {
			score := calculateImageQualityScore(file)
			if score > bestScore {
				bestScore = score
				bestFile = file
			}
		}

		bestFilename := filepath.Base(bestFile)
		ext := filepath.Ext(bestFile)
		baseName := strings.TrimSuffix(bestFilename, ext)
		newFilename := baseName + "-" + ext

		newPath := filepath.Join(filepath.Dir(bestFile), newFilename)

		if err := os.Rename(bestFile, newPath); err != nil {
			fmt.Fprintf(out, "  重命名失败 %s: %v\n", bestFilename, err)
		} else {
			fmt.Fprintf(out, "  [最佳] %s -> %s (得分: %.2f)\n", bestFilename, newFilename, bestScore)
			markedCount++
		}
	}

	fmt.Fprintf(out, "\n%s完成！共标记 %d 个最佳文件%s\n", ColorGreen, markedCount, ColorReset)
}

// 同步RAW文件的标记
func keepRawByJpgSelection(jpgDir, rawDir string) {
	rawFiles := make(map[string]string) // baseName -> fullPath

	// 扫描RAW目录，包括子目录
	err := filepath.Walk(rawDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身，只处理文件
		if info.IsDir() {
			return nil
		}

		ext := strings.ToUpper(filepath.Ext(path))
		if rawExtensions[ext] {
			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			rawFiles[baseName] = path
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "扫描RAW目录失败: %v\n", err)
		return
	}

	if len(rawFiles) == 0 {
		fmt.Fprintf(out, "未找到RAW文件\n")
		return
	}

	fmt.Fprintf(out, "找到 %d 个RAW文件\n", len(rawFiles))

	updatedCount := 0

	// 扫描JPG目录，查找标记的文件
	err = filepath.Walk(jpgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)
		ext := filepath.Ext(filename)
		baseName := strings.TrimSuffix(filename, ext)

		// 检查是否是标记文件（有-后缀）
		if strings.HasSuffix(baseName, "-") {
			originalBaseName := strings.TrimSuffix(baseName, "-")

			// 在RAW文件中查找对应的文件
			if rawPath, exists := rawFiles[originalBaseName]; exists {
				rawExt := filepath.Ext(rawPath)
				newRawFilename := originalBaseName + "-" + rawExt
				newRawPath := filepath.Join(filepath.Dir(rawPath), newRawFilename)

				if err := os.Rename(rawPath, newRawPath); err != nil {
					fmt.Fprintf(out, "  重命名RAW失败 %s: %v\n", filepath.Base(rawPath), err)
				} else {
					fmt.Fprintf(out, "  [同步] %s -> %s\n", filepath.Base(rawPath), newRawFilename)
					updatedCount++
				}
				delete(rawFiles, originalBaseName)
			} else {
				fmt.Fprintf(out, "  [警告] 未找到对应的RAW文件: %s\n", originalBaseName)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "同步RAW标记失败: %v\n", err)
		return
	}

	fmt.Fprintf(out, "\nRAW目录同步完成！共更新 %d 个文件\n", updatedCount)
}

// 计算图片质量得分
func calculateImageQualityScore(imagePath string) float64 {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 缩小图片以提高性能
	scale := 0.25
	scaledWidth := int(float64(width) * scale)
	scaledHeight := int(float64(height) * scale)

	grays := make([]float64, scaledHeight*scaledWidth)
	stepX := width / scaledWidth
	stepY := height / scaledHeight

	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			px := x * stepX
			py := y * stepY
			if px >= width {
				px = width - 1
			}
			if py >= height {
				py = height - 1
			}
			r, g, b, _ := img.At(px, py).RGBA()
			gray := float64(r*299+g*587+b*114) / 1000 / 256 * 255
			grays[y*scaledWidth+x] = gray
		}
	}

	// 1. 清晰度检测（拉普拉斯算子）
	sharpness := 0.0
	for y := 1; y < scaledHeight-1; y++ {
		for x := 1; x < scaledWidth-1; x++ {
			center := grays[y*scaledWidth+x]
			left := grays[y*scaledWidth+(x-1)]
			right := grays[y*scaledWidth+(x+1)]
			up := grays[(y-1)*scaledWidth+x]
			down := grays[(y+1)*scaledWidth+x]

			laplacian := math.Abs(4*center - left - right - up - down)
			sharpness += laplacian
		}
	}
	sharpness /= float64((scaledWidth-2) * (scaledHeight-2))

	// 2. 亮度分布
	totalBrightness := 0.0
	for _, v := range grays {
		totalBrightness += v
	}
	avgBrightness := totalBrightness / float64(len(grays))

	centerX := scaledWidth / 2
	centerY := scaledHeight / 2
	radius := min(scaledWidth, scaledHeight) / 4

	centerBrightness := 0.0
	centerCount := 0
	for y := centerY - radius; y < centerY+radius; y++ {
		for x := centerX - radius; x < centerX+radius; x++ {
			if x >= 0 && x < scaledWidth && y >= 0 && y < scaledHeight {
				centerBrightness += grays[y*scaledWidth+x]
				centerCount++
			}
		}
	}
	if centerCount > 0 {
		centerBrightness /= float64(centerCount)
	}

	centerRatio := centerBrightness / avgBrightness
	if centerRatio > 1.2 {
		centerRatio = 1.0
	}

	// 综合评分
	score := sharpness*0.6 + centerRatio*40

	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 检查目录是否为空（没有图片文件）
func isDirEmpty(dirPath string, dirType string) bool {
	hasFiles := false

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		ext := strings.ToUpper(filepath.Ext(path))

		// 根据目录类型检查相应的文件扩展名
		if dirType == "JPG" {
			if ext == ".JPG" || ext == ".JPEG" {
				hasFiles = true
				return filepath.SkipAll // 找到文件后停止遍历
			}
		} else if dirType == "RAW" {
			if rawExtensions[ext] {
				hasFiles = true
				return filepath.SkipAll // 找到文件后停止遍历
			}
		}

		return nil
	})

	if err != nil {
		return false // 出错时不删除
	}

	return !hasFiles // 没有文件才返回 true
}

// 删除空目录
func removeEmptyDir(dirPath, dirType string) bool {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return false // 目录不存在
	}

	if isDirEmpty(dirPath, dirType) {
		if err := os.RemoveAll(dirPath); err != nil {
			fmt.Fprintf(out, "  警告：删除目录失败 %s: %v\n", dirPath, err)
			return false
		}
		return true
	}

	return false
}
