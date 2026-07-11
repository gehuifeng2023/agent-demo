package tool

import (
	"path/filepath"
	"strings"
	"unicode"
)

func RouteTool(question string) string {
	s := strings.ToLower(question)
	if strings.Contains(s, "读取") ||
		strings.Contains(s, "查看文件") ||
		strings.Contains(s, "打开文件") ||
		strings.Contains(s, "读一下") ||
		strings.Contains(s, "看一下") ||
		strings.Contains(s, ".md") ||
		strings.Contains(s, ".txt") {
		return "file_reader"
	}
	if looksLikeLogAnalysisQuestion(s) {
		return "log_analyzer"
	}
	return ""
}

func ExtractToolInput(toolName string, question string) string {
	switch toolName {
	case "file_reader":
		return ExtractFilePath(question)
	case "log_analyzer":
		return strings.TrimSpace(question)
	default:
		return strings.TrimSpace(question)
	}
}

func ExtractFilePath(question string) string {
	fields := strings.FieldsFunc(question, func(r rune) bool {
		return unicode.IsSpace(r) ||
			r == '"' ||
			r == '\'' ||
			r == '`' ||
			r == '“' ||
			r == '”' ||
			r == '‘' ||
			r == '’' ||
			r == '，' ||
			r == '。' ||
			r == '；' ||
			r == '：' ||
			r == ',' ||
			r == ';' ||
			r == ':'
	})

	for _, field := range fields {
		candidate := strings.Trim(field, "[](){}<>")
		if path := extractPathCandidate(candidate); path != "" {
			return path
		}
	}

	return ""
}

func extractPathCandidate(text string) string {
	lower := strings.ToLower(text)
	for _, ext := range []string{".md", ".txt"} {
		idx := strings.Index(lower, ext)
		if idx < 0 {
			continue
		}

		end := idx + len(ext)
		start := idx - 1
		for start >= 0 && isPathChar(text[start]) {
			start--
		}
		candidate := text[start+1 : end]
		if strings.ToLower(filepath.Ext(candidate)) == ext {
			return candidate
		}
	}
	return ""
}

func isPathChar(b byte) bool {
	return b == '/' ||
		b == '\\' ||
		b == '.' ||
		b == '_' ||
		b == '-' ||
		(b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z')
}

func looksLikeLogAnalysisQuestion(text string) bool {
	keywords := []string{
		"分析日志",
		"日志分析",
		"排查日志",
		"error",
		"warn",
		"panic",
		"exception",
		"timeout",
		"failed",
		"502",
		"500",
		"401",
		"403",
		"405",
		"request_id",
		"trace_id",
		"apisix",
		"nginx",
		"kubernetes",
		"connection refused",
		"connection reset",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
