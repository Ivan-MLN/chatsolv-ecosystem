# OpenCode Free Models Reference & Benchmark

OpenCode (`https://opencode.ai/zen/v1`) offers several permanent free LLM models for text inference.

## Endpoint & Authentication
- Endpoint: `https://opencode.ai/zen/v1/chat/completions` (OpenAI compatible)
- Models endpoint: `https://opencode.ai/zen/v1/models`
- **Crucial Header**: Must send a valid `User-Agent` (e.g. `User-Agent: opencode/1.0.0`) to avoid HTTP 403 Forbidden errors.

## Active Free Models & Benchmark Results (2026)

| Model Name | TPS (Tokens/sec) | Avg Latency | Status / Kualitas |
|---|---|---|---|
| `ling-3.0-flash-free` | **~125–144 TPS** | ~2.0s | Paling kencang, lancar Indonesia & koding |
| `mimo-v2.5-free` | **~52–55 TPS** | ~2.8s | Sangat stabil & responsif |
| `deepseek-v4-flash-free` | **~33–39 TPS** | ~3.3s | Reasoning tajam khas DeepSeek |
| `laguna-s-2.1-free` | **~21–44 TPS** | ~6.8s | Stable, medium speed |
| `longcat-2.0-free` | **~20 TPS** | ~14.5s | Slower response |

## Known Errors / Offline Models
- `nemotron-3-ultra-free`: Timeout / offline
- `north-mini-code-free`: HTTP 401 Unauthorized
- `ling-3.0-tiny-free`: Null completion output
