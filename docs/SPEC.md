export OPENAI_API_KEY="你的 API Key"
export LLM_MODEL="gpt-5.5"

$env:OPENAI_API_KEY="你的 API Key"
$env:LLM_MODEL="gpt-5.5"

$env:LLM_MODE="mock"
go run ./cmd/server