package pkg

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 重复照片归类（内部函数）
func step2(photoDir string) {
	jpgDir := filepath.Join(photoDir, "jpg")
	rawDir := filepath.Join(photoDir, "raw")

	if _, err := os.Stat(jpgDir); os.IsNotExist(err) {
		fmt.Fprintf(out, "错误: jpg目录不存在，请先执行步骤1\n")
		return
	}

	fmt.Fprintf(out, "\n处理JPG目录...\n")
	processDirectory(jpgDir, "JPG")

	if _, err := os.Stat(rawDir); err == nil {
		fmt.Fprintf(out, "\n处理RAW目录...\n")
		processRawByJpgStructure(jpgDir, rawDir)
	}
}

// 处理单个目录，找出并归类重复图片
func processDirectory(dirPath string, fileType string) {
	type fileInfo struct {
		path      string
		birthTime int64
	}
	var files []fileInfo
	targetExts := map[string]bool{".JPG": true, ".JPEG": true}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToUpper(filepath.Ext(path))
		if targetExts[ext] {
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			if !strings.Contains(relPath, string(filepath.Separator)) {
				var birthTime int64
				if sys := info.Sys(); sys != nil {
					switch sys := sys.(type) {
					case interface{ Birthtime() time.Time }:
						birthTime = sys.Birthtime().Unix()
					default:
						birthTime = info.ModTime().Unix()
					}
				} else {
					birthTime = info.ModTime().Unix()
				}
				files = append(files, fileInfo{
					path:      path,
					birthTime: birthTime,
				})
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "扫描%s文件失败: %v\n", fileType, err)
		return
	}

	if len(files) == 0 {
		fmt.Fprintf(out, "未找到%s文件\n", fileType)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].birthTime < files[j].birthTime
	})

	fmt.Fprintf(out, "找到 %d 张%s图片\n", len(files), fileType)

	const timeThreshold = 30

	var groupsCreated int
	var currentGroup []string
	var currentHash string
	var lastBirthTime int64

	for _, file := range files {
		if len(currentGroup) == 0 {
			hash, err := calculateImageHash(file.path)
			if err == nil {
				currentGroup = []string{file.path}
				currentHash = hash
				lastBirthTime = file.birthTime
			}
			continue
		}

		timeDiff := file.birthTime - lastBirthTime
		if timeDiff > timeThreshold {
			if len(currentGroup) > 1 {
				moveGroup(dirPath, currentGroup)
				groupsCreated++
			}
			currentGroup = nil
			currentHash = ""
			lastBirthTime = 0
			hash, err := calculateImageHash(file.path)
			if err == nil {
				currentGroup = []string{file.path}
				currentHash = hash
				lastBirthTime = file.birthTime
			}
			continue
		}

		hash, err := calculateImageHash(file.path)
		if err != nil {
			continue
		}

		if hammingDistance(currentHash, hash) <= SimilarityThreshold {
			currentGroup = append(currentGroup, file.path)
			lastBirthTime = file.birthTime
		} else {
			if len(currentGroup) > 1 {
				moveGroup(dirPath, currentGroup)
				groupsCreated++
			}
			currentGroup = []string{file.path}
			currentHash = hash
			lastBirthTime = file.birthTime
		}
	}

	if len(currentGroup) > 1 {
		moveGroup(dirPath, currentGroup)
		groupsCreated++
	}

	fmt.Fprintf(out, "%s目录处理完成！共创建 %d 个分组\n", fileType, groupsCreated)
}

func moveGroup(dirPath string, files []string) {
	groupName := strings.TrimSuffix(filepath.Base(files[0]), filepath.Ext(files[0]))
	groupDir := filepath.Join(dirPath, groupName)

	if err := os.MkdirAll(groupDir, 0755); err != nil {
		fmt.Fprintf(out, "创建分组目录失败: %v\n", err)
		return
	}

	fmt.Fprintf(out, "创建分组: %s (%d 张图片)\n", groupName, len(files))
	for _, filePath := range files {
		destPath := filepath.Join(groupDir, filepath.Base(filePath))
		if err := os.Rename(filePath, destPath); err != nil {
			fmt.Fprintf(out, "移动文件失败 %s: %v\n", filepath.Base(filePath), err)
		} else {
			fmt.Fprintf(out, "  移动: %s\n", filepath.Base(filePath))
		}
	}
}

// 按照JPG目录的结构移动RAW文件
func processRawByJpgStructure(jpgDir, rawDir string) {
	rawFiles := make(map[string]string)

	err := filepath.Walk(rawDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToUpper(filepath.Ext(path))
		if rawExtensions[ext] {
			relPath, err := filepath.Rel(rawDir, path)
			if err != nil {
				return err
			}
			if !strings.Contains(relPath, string(filepath.Separator)) {
				baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				rawFiles[baseName] = path
			}
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

	jpgGroups := make(map[string][]string)
	err = filepath.Walk(jpgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if path == jpgDir {
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}

		groupName := filepath.Base(path)

		err = filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if subInfo.IsDir() {
				return nil
			}

			ext := strings.ToUpper(filepath.Ext(subPath))
			if ext == ".JPG" || ext == ".JPEG" {
				baseName := strings.TrimSuffix(filepath.Base(subPath), filepath.Ext(subPath))
				jpgGroups[groupName] = append(jpgGroups[groupName], baseName)
			}

			return nil
		})

		return err
	})

	if err != nil {
		fmt.Fprintf(out, "扫描JPG目录结构失败: %v\n", err)
		return
	}

	movedCount := 0

	for groupName, jpgFilenames := range jpgGroups {
		groupDir := filepath.Join(rawDir, groupName)

		if err := os.MkdirAll(groupDir, 0755); err != nil {
			fmt.Fprintf(out, "创建RAW分组目录失败 %s: %v\n", groupName, err)
			continue
		}

		for _, jpgFilename := range jpgFilenames {
			if rawPath, exists := rawFiles[jpgFilename]; exists {
				destPath := filepath.Join(groupDir, filepath.Base(rawPath))
				if err := os.Rename(rawPath, destPath); err != nil {
					fmt.Fprintf(out, "移动RAW文件失败 %s: %v\n", filepath.Base(rawPath), err)
				} else {
					fmt.Fprintf(out, "移动: %s -> %s/\n", filepath.Base(rawPath), groupName)
					movedCount++
					delete(rawFiles, jpgFilename)
				}
			}
		}
	}

	fmt.Fprintf(out, "\nRAW目录处理完成！共移动 %d 个文件\n", movedCount)
}

// 计算图片的感知哈希值
func calculateImageHash(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	resizedSize := 32
	grays := make([]float64, resizedSize*resizedSize)

	for y := 0; y < resizedSize; y++ {
		for x := 0; x < resizedSize; x++ {
			px := x * width / resizedSize
			py := y * height / resizedSize
			if px >= width {
				px = width - 1
			}
			if py >= height {
				py = height - 1
			}
			r, g, b, _ := img.At(px, py).RGBA()
			gray := float64(r*299+g*587+b*114) / 1000 / 256 * 255
			grays[y*resizedSize+x] = gray
		}
	}

	dctCoeffs := performDCT(grays, resizedSize)

	hashSize := 8
	lowFreq := make([]float64, hashSize*hashSize)
	for v := 0; v < hashSize; v++ {
		for u := 0; u < hashSize; u++ {
			lowFreq[v*hashSize+u] = dctCoeffs[v*resizedSize+u]
		}
	}

	median := calculateMedian(lowFreq)

	var hash strings.Builder
	for i := 0; i < len(lowFreq); i++ {
		if lowFreq[i] > median {
			hash.WriteString("1")
		} else {
			hash.WriteString("0")
		}
	}

	return hash.String(), nil
}

// 执行2D DCT变换
func performDCT(input []float64, size int) []float64 {
	output := make([]float64, size*size)

	colDCT := make([][]float64, size)
	for x := 0; x < size; x++ {
		col := make([]float64, size)
		for y := 0; y < size; y++ {
			col[y] = input[y*size+x]
		}
		dctCol := performDCT1D(col)
		for y := 0; y < size; y++ {
			colDCT[y] = append(colDCT[y], dctCol[y])
		}
	}

	for y := 0; y < size; y++ {
		dctRow := performDCT1D(colDCT[y])
		for x := 0; x < size; x++ {
			output[y*size+x] = dctRow[x]
		}
	}

	return output
}

// 执行1D DCT变换
func performDCT1D(input []float64) []float64 {
	N := len(input)
	output := make([]float64, N)

	for k := 0; k < N; k++ {
		sum := 0.0
		for n := 0; n < N; n++ {
			sum += input[n] * math.Cos(math.Pi*float64(k)*(2*float64(n)+1)/float64(2*N))
		}

		if k == 0 {
			output[k] = sum * math.Sqrt(1.0/float64(N))
		} else {
			output[k] = sum * math.Sqrt(2.0/float64(N))
		}
	}

	return output
}

// 计算中位数
func calculateMedian(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// 计算汉明距离
func hammingDistance(hash1, hash2 string) int {
	if len(hash1) != len(hash2) {
		return 64
	}

	distance := 0
	for i := 0; i < len(hash1); i++ {
		if hash1[i] != hash2[i] {
			distance++
		}
	}
	return distance
}