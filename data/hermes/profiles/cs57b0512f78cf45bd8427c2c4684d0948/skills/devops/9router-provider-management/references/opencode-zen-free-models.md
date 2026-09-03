# OpenCode Zen Free Models Benchmark & Context Windows

OpenCode Zen (`https://opencode.ai/zen/v1`) offers 8 free models.

## Free Models Summary & Context Windows

| Model ID | Provider/Developer | Context Window (Input / Output) | Average Speed (TPS) | Stability Notes |
| :--- | :--- | :--- | :--- | :--- |
| `deepseek-v4-flash-free` | DeepSeek | 1,048,576 (1M) / 128K | ~50–66 TPS | **High (100%)** — Recommended for 9router combo/free tier |
| `ling-3.0-flash-free` | InclusionAI | 256,000 (256K) / 16K | ~106–144 TPS | **Medium** — Fast burst response, but susceptible to HTTP 503 / rate limits |
| `mimo-v2.5-free` | Xiaomi LLM | 1,048,576 (1M) / 131K | ~52–55 TPS | **Medium** — Intermittent rate limits |
| `laguna-s-2.1-free` | Poolside AI | 1,048,576 (1M) / 64K | ~21–43 TPS | **Low** — Flaky / frequent timeouts on free tier |
| `longcat-2.0-free` | Meituan | 128,000 (128K) / 8K | ~20–28 TPS | **High** — Stable for long-text generation |

## API Testing Endpoint
Base URL: `https://opencode.ai/zen/v1/chat/completions`
User-Agent header required: `User-Agent: opencode/1.0.0` (missing UA or curl default can trigger HTTP 403 Forbidden).
