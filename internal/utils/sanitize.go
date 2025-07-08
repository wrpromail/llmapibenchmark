package utils

import "strings"

// SanitizeFilename 清理文件名，移除或替换不安全的字符
func SanitizeFilename(filename string) string {
	// 替换路径分隔符
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// 替换其他可能有问题的字符
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")

	return filename
}
