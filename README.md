# Agent Demo

`agent-demo` 是一个 Go 实现的 RAG Agent Demo。服务启动时加载默认知识库，运行过程中可以上传文件扩充知识库，并提供聊天问答、全量知识库查看和按问题召回知识库内容的 HTTP API。

## 功能概览

- 启动时加载 `knowledge_attachment/default/` 下的 Markdown 文件作为默认知识库，知识库 ID 为 `default`。
- 支持上传 `.md`、`.txt`、`.docx`、`.doc` 文件，上传内容会切分为 chunks 并写入共享召回器。
- Chat API 每次请求都会从共享召回器读取最新知识库内容。
- Knowledge API 支持查看全部 chunks，也支持只根据问题做知识库召回。
- LLM 支持 `mock`、`openai`、`gemini` 三种模式。
- Chat API 支持自动调用 `file_reader` 读取安全目录内的 `.md`/`.txt` 文件，也支持调用 `log_analyzer` 分析常见服务日志。
- 运行参数从 `configs/local.yaml` 读取，也可以通过 `CONFIG_PATH` 指定其他配置文件。

## 技术栈

- Go 1.25+
- `gopkg.in/yaml.v3`：YAML 配置解析
- `github.com/nguyenthenguyen/docx`：Word 文档内容读取

## 项目结构

```text
agent-demo/
├── cmd/server/                 # HTTP 服务入口
├── configs/                    # 本地配置目录，默认被 .gitignore 忽略
├── internal/agent/             # Agent 编排：意图识别、召回、Prompt、LLM、会话
├── internal/config/            # 配置结构、默认值和加载逻辑
├── internal/converter/         # 上传文件转换器
├── internal/document/          # 文档加载和 chunk 切分
├── internal/handler/           # HTTP handler
├── internal/knowledge/         # 默认知识库注册表
├── internal/llm/               # Mock/OpenAI/Gemini LLM client
├── internal/retriever/         # 统一召回器和关键词召回
├── internal/session/           # 内存会话存储
├── internal/tool/              # Agent 工具：注册、路由、文件读取、日志分析
├── internal/upload/            # 上传文件保存和上传文件知识库
├── knowledge_attachment/
│   ├── default/                # 启动加载的默认知识库
│   └── days/                   # 上传文件保存目录，默认被 .gitignore 忽略
├── scripts/                    # 本地脚本
├── go.mod
└── README.md
```

## 配置

服务默认读取：

```bash
configs/local.yaml
```

也可以通过环境变量覆盖配置路径：

```bash
CONFIG_PATH=/path/to/local.yaml go run ./cmd/server
```

当前 `configs/local.yaml` 是本地配置文件，仓库 `.gitignore` 默认忽略 `/configs`。默认配置结构如下：

```yaml
server:
  addr: ":8080"

llm:
  mode: mock
  api_key: ""
  model: ""
  base_url: ""
  timeout_seconds: 60

rag:
  docs_dir: ""
  top_k: 3

upload:
  dir: knowledge_attachment/days
  max_size_mb: 20

knowledge:
  root_dir: knowledge_attachment/default/

session:
  max_messages: 30
  recent_limit: 8

tool:
  enabled: true
  root_dir: knowledge_attachment/default/

intent:
  mode: rule
```

关键字段说明：

- `server.addr`：HTTP 服务监听地址。
- `llm.mode`：`mock`、`openai` 或 `gemini`。
- `llm.api_key`：LLM API key。为空时 OpenAI/Gemini 会 fallback 到 `OPENAI_API_KEY` / `GEMINI_API_KEY`。
- `llm.model`：LLM 模型名。为空时会 fallback 到 `LLM_MODEL` 或 client 默认模型。
- `llm.base_url`：LLM base URL。为空时使用 client 默认地址。
- `llm.timeout_seconds`：LLM 请求超时时间。
- `knowledge.root_dir`：启动时加载的默认知识库目录。
- `upload.dir`：上传文件保存根目录。
- `upload.max_size_mb`：上传文件大小限制。
- `rag.top_k`：聊天和知识库召回默认返回 chunk 数。
- `session.max_messages`：每个 session 最多保存的消息数。
- `session.recent_limit`：构建 Prompt 时读取的最近消息数。
- `tool.enabled`：是否启用工具自动调用。默认启用。
- `tool.root_dir`：`file_reader` 允许读取的安全根目录。为空时使用 `knowledge.root_dir`。

## 快速开始

安装依赖：

```bash
go mod download
```

使用 mock LLM 启动服务：

```bash
CONFIG_PATH=configs/local.yaml go run ./cmd/server
```

构建：

```bash
go build ./cmd/...
```

测试：

```bash
GOCACHE=/tmp/agent-demo-go-build go test ./... -count=1
```

## LLM 模式

Mock 模式适合本地开发，不需要 API key：

```yaml
llm:
  mode: mock
```

OpenAI 模式：

```yaml
llm:
  mode: openai
  api_key: "你的 OPENAI API Key"
  model: "gpt-5.5"
```

Gemini 模式：

```yaml
llm:
  mode: gemini
  api_key: "你的 GEMINI API Key"
  model: "gemini-3.5-flash"
```

如果 `api_key` 或 `model` 没有配置，服务会继续读取环境变量：

- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `LLM_MODEL`

## API

默认服务地址为 `http://localhost:8080`。

### 聊天问答

```http
POST /api/v1/chat
Content-Type: application/json
```

请求示例：

```json
{
  "question": "什么是 RAG？",
  "sessionID": "s1",
  "knowledge_base_ids": ["default"],
  "file_ids": []
}
```

字段说明：

- `question`：必填，用户问题。
- `sessionID`：可选，会话 ID。为空时服务自动生成。
- `type`：可选，指定 Prompt 类型；为空时使用规则分类。
- `knowledge_base_ids`：可选，指定默认知识库 ID，例如 `default`。
- `file_ids`：可选，指定上传文件 ID。

响应示例：

```json
{
  "sessionID": "s1",
  "answer": "这是 Mock LLM 返回的回答...",
  "type": "chat",
  "sources": [
    {
      "file": "knowledge_attachment/default/faq.md",
      "chunk_id": "knowledge_attachment/default/faq.md-1",
      "content": "...",
      "position": 1
    }
  ]
}
```

### 工具调用

`/api/v1/chat` 会根据问题自动路由工具。当前内置工具：

- `file_reader`：读取 `tool.root_dir` 下的 `.md`/`.txt` 文件内容，并把结果作为工具上下文交给 LLM。
- `log_analyzer`：分析 APISIX、Nginx、Kubernetes、Go 服务等日志，提取错误类型、request_id、trace_id、状态码、可能根因和排查建议。

请求示例：

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"请读取 faq.md 并总结重点"}'
```

说明：

- `faq.md` 会按相对路径解析到 `tool.root_dir/faq.md`。
- 绝对路径、`../` 越权路径和目录路径会被拒绝。
- 日志分析会把整段问题文本作为工具输入，因此可以直接粘贴日志和排查问题。
- 工具结果只作为回答上下文，不会写入响应里的 `sources`；`sources` 仍表示 RAG 召回来源。

日志分析示例：

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"question":"帮我分析日志 request_id=abc trace_id=t1 status=502 upstream timeout"}'
```

### 上传文件扩充知识库

```http
POST /api/v1/files/upload
Content-Type: multipart/form-data
```

请求示例：

```bash
curl -F "file=@notes.txt" http://localhost:8080/api/v1/files/upload
```

响应示例：

```json
{
  "file_id": "f-1783607989609598778",
  "file_name": "notes.txt",
  "size": 128
}
```

说明：

- 支持 `.md`、`.txt`、`.docx`、`.doc`。
- 文件保存到 `upload.dir/YYYY-MM-DD/`。
- 转换成功后 chunks 写入共享召回器，后续 chat 和 knowledge API 可以立即召回。

### 获取全量知识库信息

```http
GET /api/v1/knowledge
```

响应示例：

```json
{
  "chunks": [
    {
      "type": "knowledge_base",
      "knowledge_base_id": "default",
      "file": "knowledge_attachment/default/faq.md",
      "chunk_id": "knowledge_attachment/default/faq.md-1",
      "content": "...",
      "position": 1
    },
    {
      "type": "file",
      "file_id": "f-1783607989609598778",
      "file": "knowledge_attachment/days/2026-07-10/f-1783607989609598778.txt",
      "chunk_id": "...",
      "content": "...",
      "position": 1
    }
  ]
}
```

### 根据问题召回知识库信息

```http
POST /api/v1/knowledge/retrieve
Content-Type: application/json
```

请求示例：

```json
{
  "question": "RAG 是什么？",
  "knowledge_base_ids": ["default"],
  "file_ids": [],
  "top_k": 3
}
```

响应示例：

```json
{
  "question": "RAG 是什么？",
  "chunks": [
    {
      "type": "knowledge_base",
      "knowledge_base_id": "default",
      "file": "knowledge_attachment/default/faq.md",
      "chunk_id": "knowledge_attachment/default/faq.md-1",
      "content": "...",
      "position": 1
    }
  ]
}
```

说明：

- 该接口只做召回，不调用 LLM。
- `top_k` 小于等于 0 时使用 `rag.top_k`。
- 不传 `knowledge_base_ids` 和 `file_ids` 时，默认从所有默认知识库和上传文件中召回。

## 知识库流程

1. 服务启动时读取 `knowledge.root_dir`。
2. `document.LoadFromDir` 加载目录下的 `.md` 文件。
3. `document.SplitByParagraph` 按段落切分为 chunks。
4. 默认 chunks 注册到 `UnifiedRetriever`，知识库 ID 为 `default`。
5. 上传文件转换成功后，文件 chunks 通过同一个 `UnifiedRetriever` 存储。
6. Chat API 和 Knowledge API 每次请求都从共享 `UnifiedRetriever` 读取最新内容。

## 开发说明

- 默认知识库目前只加载 Markdown 文件。
- 上传文件知识库保存在内存中，服务重启后不会自动恢复历史上传 chunks。
- `knowledge_attachment/days/` 是上传文件目录，默认被 `.gitignore` 忽略。
- `configs/local.yaml` 是本地运行配置，默认被 `.gitignore` 忽略。
- 详细开发规范见 [AGENTS.md](./AGENTS.md)。

## 常用命令

```bash
go mod download
GOCACHE=/tmp/agent-demo-go-build go test ./... -count=1
go build ./cmd/...
CONFIG_PATH=configs/local.yaml go run ./cmd/server
```
