package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// step4 删除多余的RAW文件
func step4(photoDir string) {
	// 检查 jpg 和 raw 目录是否存在
	jpgDir := filepath.Join(photoDir, "jpg")
	rawDir := filepath.Join(photoDir, "raw")

	if _, err := os.Stat(jpgDir); os.IsNotExist(err) {
		fmt.Fprintf(out, "错误: 未找到 jpg 目录，请先执行步骤1\n")
		return
	}

	if _, err := os.Stat(rawDir); os.IsNotExist(err) {
		fmt.Fprintf(out, "错误: 未找到 raw 目录，请先执行步骤1\n")
		return
	}

	fmt.Fprintf(out, "\n警告: 此操作将删除在 raw 目录中但没有对应 JPG 文件的 RAW 文件\n")
	fmt.Fprintf(out, "建议: 请确保已备份重要的 RAW 文件\n")

	// 收集所有需要删除的文件
	var filesToDelete []string

	// 遍历 raw 目录（包含子目录）
	err := filepath.Walk(rawDir, func(rawPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 获取相对于 raw 目录的相对路径
		relPath, err := filepath.Rel(rawDir, rawPath)
		if err != nil {
			return err
		}

		// 构造对应的 JPG 文件路径
		jpgPath := filepath.Join(jpgDir, relPath)
		
		// 获取不带扩展名的文件名
		ext := filepath.Ext(rawPath)
		baseName := strings.TrimSuffix(filepath.Base(rawPath), ext)
		
		// 尝试查找对应的 JPG 文件（可能有 .JPG 或 .jpg 扩展名）
		parentDir := filepath.Dir(jpgPath)
		jpgPath1 := filepath.Join(parentDir, baseName+".JPG")
		jpgPath2 := filepath.Join(parentDir, baseName+".jpg")

		// 检查 JPG 文件是否存在
		jpgExists := false
		if _, err := os.Stat(jpgPath1); err == nil {
			jpgExists = true
		} else if _, err := os.Stat(jpgPath2); err == nil {
			jpgExists = true
		}

		// 如果 JPG 不存在，则标记为待删除
		if !jpgExists {
			filesToDelete = append(filesToDelete, rawPath)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "错误: 遍历 raw 目录失败: %v\n", err)
		return
	}

	// 显示统计信息
	if len(filesToDelete) == 0 {
		fmt.Fprintf(out, "\n✓ 没有发现多余的 RAW 文件\n")
		return
	}

	fmt.Fprintf(out, "\n发现 %d 个多余的 RAW 文件将被删除:\n", len(filesToDelete))
	for i, file := range filesToDelete {
		relPath, _ := filepath.Rel(photoDir, file)
		fmt.Fprintf(out, "  [%d] %s\n", i+1, relPath)
	}

	// 用户确认
	if !RequestConfirm("\n确认删除这些文件吗？(输入 'y' 确认，其他任何输入取消): ") {
		fmt.Fprintf(out, "操作已取消\n")
		return
	}

	// 执行删除
	deletedCount := 0
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			fmt.Fprintf(out, "错误: 删除文件失败 %s: %v\n", file, err)
		} else {
			deletedCount++
		}
	}

	fmt.Fprintf(out, "\n✓ 成功删除 %d 个多余的 RAW 文件\n", deletedCount)
}