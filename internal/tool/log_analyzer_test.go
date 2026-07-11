package tool

import (
	"context"
	"strings"
	"testing"
)

func TestLogAnalyzerToolAnalyzesGateway502(t *testing.T) {
	got, err := LogAnalyzerTool{}.Execute(context.Background(), "request_id=abc trace_id=t1 status=502 upstream timeout")
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}

	for _, want := range []string{
		`"error_type": "gateway_502"`,
		`"request_id": "abc"`,
		`"trace_id": "t1"`,
		`"status_code": "502"`,
		"网关或上游服务异常",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q, got %s", want, got)
		}
	}
}

func TestLogAnalyzerToolHasSuggestionsForKnownCategories(t *testing.T) {
	tests := []struct {
		name      string
		log       string
		errorType string
		rootCause string
	}{
		{name: "403", log: "status=403 forbidden", errorType: "auth_403", rootCause: "权限不足或访问被拒绝"},
		{name: "timeout", log: "upstream timeout", errorType: "timeout", rootCause: "请求或上游处理超时"},
		{name: "connection", log: "connection refused", errorType: "connection", rootCause: "连接被拒绝或连接被重置"},
		{name: "panic", log: "panic: runtime error", errorType: "panic", rootCause: "服务运行时 panic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LogAnalyzerTool{}.Execute(context.Background(), tt.log)
			if err != nil {
				t.Fatalf("execute tool: %v", err)
			}
			if !strings.Contains(got, `"error_type": "`+tt.errorType+`"`) {
				t.Fatalf("expected error type %q, got %s", tt.errorType, got)
			}
			if !strings.Contains(got, tt.rootCause) {
				t.Fatalf("expected root cause %q, got %s", tt.rootCause, got)
			}
		})
	}
}

func TestLogAnalyzerToolRejectsEmptyInput(t *testing.T) {
	_, err := LogAnalyzerTool{}.Execute(context.Background(), " ")
	if err == nil {
		t.Fatal("expected empty input to fail")
	}
	if !strings.Contains(err.Error(), "input is empty") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestLogAnalyzerToolReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := LogAnalyzerTool{}.Execute(ctx, "status=502")
	if err == nil {
		t.Fatal("expected context error")
	}
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
