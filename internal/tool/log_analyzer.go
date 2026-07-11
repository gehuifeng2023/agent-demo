package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type LogAnalyzerTool struct{}

func (t LogAnalyzerTool) Name() string        { return "log_analyzer" }
func (t LogAnalyzerTool) Description() string { return "分析 APISIX、Nginx、K8s、Go 服务日志" }

func (t LogAnalyzerTool) Execute(ctx context.Context, input string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("input is empty")
	}

	r := LogAnalysisResult{
		ErrorType:  classifyError(input),
		RequestID:  extractValue(input, "request_id"),
		TraceID:    extractValue(input, "trace_id"),
		StatusCode: extractValue(input, "status"),
	}
	r.RootCause, r.Suggestions = suggest(r.ErrorType)
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal log analysis result: %w", err)
	}
	return string(data), nil
}

func suggest(t string) (string, []string) {
	switch t {
	case "gateway_502":
		return "网关或上游服务异常", []string{"检查上游服务是否存活", "检查超时配置", "查看 APISIX/Nginx upstream 日志"}
	case "auth_401":
		return "认证失败", []string{"检查 token/session", "检查鉴权服务返回"}
	case "auth_403":
		return "权限不足或访问被拒绝", []string{"检查用户权限", "检查网关或应用鉴权策略", "核对资源访问范围"}
	case "method_405":
		return "请求方法不被允许", []string{"确认接口方法", "检查路由配置"}
	case "timeout":
		return "请求或上游处理超时", []string{"检查上游服务耗时", "检查网关和客户端超时配置", "查看同一 request_id 的慢调用链路"}
	case "connection":
		return "连接被拒绝或连接被重置", []string{"确认上游地址和端口可达", "检查服务实例是否存活", "排查负载均衡和连接池配置"}
	case "panic":
		return "服务运行时 panic", []string{"定位 panic 堆栈", "检查空指针或边界条件", "补充恢复和告警机制"}
	default:
		return "暂未识别明确根因", []string{"补充完整日志", "查看 request_id 对应链路"}
	}
}
