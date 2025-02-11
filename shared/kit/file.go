package kit

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteContents 写入文件内容，如果文件不存在则创建文件
func WriteContents(filePath string, content any) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempFile := filePath + ".tmp"

	if err := os.WriteFile(tempFile, []byte(String(content)), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// PutContents 写入文件内容，如果文件不存在则创建文件
func PutContents(filePath string, content any) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, []byte(String(content)), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
