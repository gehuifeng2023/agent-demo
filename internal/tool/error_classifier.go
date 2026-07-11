package tool

import "strings"

func classifyError(log string) string {
	s := strings.ToLower(log)
	switch {
	case strings.Contains(s, "502"):
		return "gateway_502"
	case strings.Contains(s, "401"):
		return "auth_401"
	case strings.Contains(s, "403"):
		return "auth_403"
	case strings.Contains(s, "405"):
		return "method_405"
	case strings.Contains(s, "timeout"):
		return "timeout"
	case strings.Contains(s, "connection refused") || strings.Contains(s, "connection reset"):
		return "connection"
	case strings.Contains(s, "panic"):
		return "panic"
	default:
		return "unknown"
	}
}
