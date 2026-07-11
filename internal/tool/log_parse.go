package tool

import "regexp"

type LogAnalysisResult struct {
	ErrorType   string   `json:"error_type"`
	Level       string   `json:"level"`
	RequestID   string   `json:"request_id"`
	TraceID     string   `json:"trace_id"`
	StatusCode  string   `json:"status_code"`
	RootCause   string   `json:"root_cause"`
	Suggestions []string `json:"suggestions"`
}

func extractValue(log, key string) string {
	re := regexp.MustCompile(key + `[:=]([A-Za-z0-9_\-\.]+)`)
	m := re.FindStringSubmatch(log)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
