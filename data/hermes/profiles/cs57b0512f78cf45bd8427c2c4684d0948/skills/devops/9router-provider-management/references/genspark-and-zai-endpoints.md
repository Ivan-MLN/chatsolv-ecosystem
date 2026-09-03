# Genspark & Z.AI Provider Integration in 9router

## Genspark LLM Proxy (`genspark`)

- **Base URL**: `https://www.genspark.ai/api/llm_proxy/v1`
- **CLI Config Location**: `~/.genspark-tool-cli/config.json` (key `api_key`, format `gsk-eyJ...`)
- **Node Type**: `openai-compatible`
- **Prefix**: `genspark`
- **Key Models**:
  - `genspark/claude-sonnet-4-6`, `genspark/claude-sonnet-5`
  - `genspark/claude-opus-5` (includes thinking/reasoning tokens)
  - `genspark/claude-fable-5`
  - `genspark/gpt-5.5`, `genspark/gpt-5.6-luna`
  - `genspark/deep-seek-v4-pro`

### Multi-Account Registration
Multiple `gsk-*` keys can be attached to the same `providerNodes` entry (`openai-compatible-chat-<id>`) by inserting multiple rows into `providerConnections` pointing to the same `node_id`. 9router will automatically round-robin and failover across accounts.

---

## Z.AI / ZCode Endpoints (`z.ai`)

Z.AI segregates pay-as-you-go API balances from **GLM Coding Plan** subscription quotas. Using the wrong Base URL results in `HTTP 429 (code 1113: Insufficient balance or no resource package)`.

### Protocol & Endpoint Matrix

| Protocol | Base URL | Quota Scope |
| :--- | :--- | :--- |
| **OpenAI Chat Completions (Coding Plan)** | `https://api.z.ai/api/coding/paas/v4` | GLM Coding Plan subscriptions (ZCode Start/Weekend) |
| **Anthropic Messages** | `https://api.z.ai/api/anthropic` | Anthropic-compatible clients (Claude Code, etc.) |
| **OpenAI Responses** | `https://api.z.ai/api/v1` | OpenAI Response format |
| **Standard Pay-As-You-Go** | `https://api.z.ai/api/paas/v4` | Resource bundles & account cash balance only |

### Supported Models
- `glm-5.3-flash`, `glm-5.3`, `glm-5.2`, `glm-5.1`, `glm-5`, `glm-4.7`, `glm-4.6`, `glm-4.5`
