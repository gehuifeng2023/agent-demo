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
	return ""
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
