package kit

import "strings"

// LogSnippet 将日志内容裁剪到指定长度，避免慢日志输出过大消息体。
func LogSnippet(value any, limit int) string {
	if limit <= 0 {
		limit = 256
	}

	text := strings.TrimSpace(String(value))
	if text == "" {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

// JoinLogSnippets 将多条日志摘要拼接成单行文本。
func JoinLogSnippets(values []string) string {
	nonEmpty := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		nonEmpty = append(nonEmpty, value)
	}
	return strings.Join(nonEmpty, " | ")
}
