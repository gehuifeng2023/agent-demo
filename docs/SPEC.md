# Agent Demo 规格说明

## 概述

Agent Demo 是一个最小化的 Go HTTP 服务，对外提供一个聊天接口：

`POST /api/v1/chat`

服务接收用户问题，并返回由 LLM 客户端生成的回答。项目刻意保持小而清晰，适合演示 `handler`、`service`、`prompt` 和 `llm` 的分层设计。

## API

### POST /api/v1/chat

请求体：

```json
{
  "question": "什么是 APISIX？"
}
```

成功响应：

状态码：`200 OK`

```json
{
  "answer": "APISIX 是一个云原生 API 网关……"
}
```

校验错误：

- JSON 格式非法时，返回 `400 Bad Request` 和 `{"error":"invalid request body"}`。
- `question` 缺失或为空白字符串时，返回 `400 Bad Request` 和 `{"error":"question is required"}`。
- 使用非 `POST` 方法访问时，返回 `405 Method Not Allowed` 和 `{"error":"method not allowed"}`。

## 架构

- `cmd/server`：进程入口，负责配置加载、依赖装配和 HTTP 服务启动。
- `internal/handler`：处理 HTTP 传输层逻辑，包括请求校验和 JSON 响应。
- `internal/service`：编排聊天业务用例。
- `internal/config`：读取运行配置，并处理默认值。
- `internal/llm`：定义 LLM 客户端契约，并提供本地模拟、百炼模型、DeepSeek 模型和 Gemini 模型实现。
- `internal/prompt`：构造提示词。
- `configs`：运行配置。

## LLM 模式

服务支持四种 LLM 模式：

- `mock`：本地模拟回答，默认模式，适合本地调试。
- `model`：调用阿里百炼 DashScope OpenAI-compatible Chat Completions API。
- `deepseek`：调用 DeepSeek OpenAI-compatible Chat Completions API。
- `gemini`：调用 Google Gemini OpenAI-compatible Chat Completions API。

配置文件位于 `configs/config.yaml`：

```yaml
llm:
  mode: mock
  api_key: "your-dashscope-api-key"
  model: qwen-plus
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
  timeout_seconds: 30
```

`LLM_MODE` 环境变量优先于配置文件中的 `llm.mode`，可用于临时切换：

```bash
LLM_MODE=model go run ./cmd/server
```

当 `LLM_MODE` 覆盖配置文件中的 `llm.mode` 时，服务会按最终模式同步使用对应 provider 的默认 `model` 和 `base_url`，避免例如 `LLM_MODE=gemini` 仍沿用 DeepSeek 地址。

DeepSeek 配置示例：

```yaml
llm:
  mode: deepseek
  api_key: "your-deepseek-api-key"
  model: deepseek-v4-flash
  base_url: "https://api.deepseek.com"
  timeout_seconds: 30
```

```bash
LLM_MODE=deepseek go run ./cmd/server
```

Gemini 配置示例：

```yaml
llm:
  mode: gemini
  api_key: "your-gemini-api-key"
  model: gemini-3.5-flash
  base_url: "https://generativelanguage.googleapis.com/v1beta/openai"
  timeout_seconds: 30
```

```bash
LLM_MODE=gemini go run ./cmd/server
```

## 运行方式

服务默认监听 `:8080`。

启动：

```bash
go run ./cmd/server
```

手动验证：

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"question":"什么是 APISIX？"}'
```

## 非目标

- 不实现认证鉴权。
- 不实现流式响应。
- 不保存会话历史。
