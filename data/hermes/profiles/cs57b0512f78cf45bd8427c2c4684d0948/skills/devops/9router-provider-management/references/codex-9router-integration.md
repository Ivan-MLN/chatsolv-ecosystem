# Codex CLI Integration with 9router

Guide for configuring the OpenAI Codex CLI (`@openai/codex`) to route LLM calls through a local or custom 9router instance.

## Config File Location

- Path: `~/.codex/config.toml`

## Required Configuration

Codex uses the `wire_api = "responses"` protocol for custom providers and expects credentials supplied via `experimental_bearer_token`.

```toml
# Model selection (e.g. cx/gpt-5.6-sol or ag/gemini-3.7-flash-high)
model = "cx/gpt-5.6-sol"
model_provider = "9router"
model_reasoning_effort = "high"

[model_providers.9router]
name = "9Router"
base_url = "http://127.0.0.1:20128/v1"
wire_api = "responses"
experimental_bearer_token = "sk-..."

[agents.subagent]
model = "cx/gpt-5.6-sol"
description = "Default subagent"

[projects."/home/nldt"]
trust_level = "trusted"
```

## Key Pitfalls & Details

1. **Reasoning Effort Configuration (`model_reasoning_effort`)**:
   - For `cx/gpt-5.6-sol`, `cx/gpt-5.6-terra`, and `cx/gpt-5.6-luna`, configure `model_reasoning_effort = "high"` (or `"xhigh"` / `"ultra"`) directly in `~/.codex/config.toml`.
2. **Model Aliases via Combos Table**:
   - If Codex sends bare model IDs (e.g. `gpt-5.6-sol` instead of `cx/gpt-5.6-sol`), 9router returns 404 `No active credentials for provider: openai`. Map the bare model alias to `cx/*` in the `combos` table:
     ```sql
     INSERT OR REPLACE INTO combos (id, name, kind, models, createdAt, updatedAt)
     VALUES ('<uuid>', 'gpt-5.6-sol', '', '["cx/gpt-5.6-sol"]', datetime('now'), datetime('now'));
     ```
     Then restart `naelrouter.service` (`systemctl restart naelrouter.service`).
3. **`wire_api` Requirement**:
   - `wire_api = "chat"` is no longer supported by modern Codex CLI versions (v0.140+). Always set `wire_api = "responses"`.
4. **Bearer Token Header**:
   - Standard `auth.json` or `OPENAI_API_KEY` environment variables might not properly propagate to custom named providers in Codex CLI.
   - Use `experimental_bearer_token = "<9router_api_key>"` inside the `[model_providers.<name>]` block in `config.toml`.
5. **Agent Role Definition Warning**:
   - Defining `[agents.subagent]` without a `description` field causes `warning: Ignoring malformed agent role definition: agent role subagent must define a description`.
   - Always supply `description = "..."`.
6. **Stale Interactive Session 401 Reconnect Error**:
   - If Codex CLI was running an existing interactive session attached to OpenAI upstream (`wss://api.openai.com/v1/responses`), switching `~/.codex/config.toml` to 9router won't affect the already-running websocket reconnect loop and will show `Unexpected status 401 Unauthorized: invalid_api_key`. Terminate the running session (`Ctrl+C` or kill) and restart Codex to instantiate the new provider connection.
