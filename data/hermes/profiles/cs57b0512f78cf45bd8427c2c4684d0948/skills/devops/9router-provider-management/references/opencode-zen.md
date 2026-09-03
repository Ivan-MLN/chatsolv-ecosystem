# OpenCode Zen Provider Reference

OpenCode Zen API (`https://opencode.ai/zen/v1`) offers OpenAI-compatible endpoints with free tiers.

## Endpoint & Headers
- Base URL: `https://opencode.ai/zen/v1`
- Models URL: `https://opencode.ai/zen/v1/models`
- Chat completions: `https://opencode.ai/zen/v1/chat/completions`
- Required Header: `User-Agent: opencode/1.0.0` (Without proper UA or headers, requests return HTTP 403 Forbidden).

## Verified Free Models & Performance Benchmark
| Model | Context Window | Max Output | Latency & Speed | Stability |
| :--- | :--- | :--- | :--- | :--- |
| `deepseek-v4-flash-free` | 1M (1,048,576) | 128K | ~3.2s latency, ~50-66 TPS | 🟩 100% Stable (Recommended) |
| `ling-3.0-flash-free` | 256K | 16K | ~1.8s latency, ~106-144 TPS | 🟧 Flaky (Intermittent HTTP 503) |
| `mimo-v2.5-free` | 1M | 131K | ~2.8s latency, ~52-55 TPS | 🟨 Moderate (Rate limits) |
| `laguna-s-2.1-free` | 1M | 64K | ~7.0s latency, ~21-43 TPS | 🟥 Flaky / Timeout |
| `longcat-2.0-free` | 128K | 8K | ~7.2s latency, ~20-28 TPS | 🟩 High stability |
