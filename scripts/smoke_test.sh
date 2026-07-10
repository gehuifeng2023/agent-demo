#!/usr/bin/env bash
set -euo pipefail

curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id":"s1","knowledge_base_ids":["sec-proxy"],"question":"sec-proxy 核心职责是什么？"}' | jq .

curl -s -F "file=@docs/project.md" http://localhost:8080/api/v1/files/upload | jq .