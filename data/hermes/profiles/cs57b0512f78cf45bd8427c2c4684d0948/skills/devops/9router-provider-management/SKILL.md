---
name: 9router-provider-management
description: "Use when debugging 9router provider errors or reconfig."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [9router, llm-router, sqlite, provider, proxy]
---

# 9router Provider Management

Use when a 9router provider throws `No active credentials for provider: <id>`, 403, or needs baseUrl/proxy reconfiguration.

## Key paths

- Runtime dir: `/root/.9router/`
- Database: `/root/.9router/db/data.sqlite`
- Table: `providerConnections`
- Systemd Service: `/etc/systemd/system/naelrouter.service` (or `9router.service` — check `systemctl list-units | grep -i router`, runs on port 20128)
- App Location: `/usr/local/lib/node_modules/9router/app`

## Schema

```sql
CREATE TABLE providerConnections (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,   -- e.g. "anthropic-compatible-<uuid>"
  authType TEXT NOT NULL,
  name TEXT,
  isActive INTEGER DEFAULT 1,  -- 0 = disabled → "No active credentials"
  data TEXT NOT NULL,           -- JSON: apiKey, baseUrl, errorCode, backoffLevel
  createdAt TEXT,
  updatedAt TEXT
);
```

## Diagnosis

```bash
# List all + active state
sqlite3 /root/.9router/db/data.sqlite \
  "SELECT id, provider, name, isActive FROM providerConnections;"

# Check if a specific provider/model is configured in connections or combos
sqlite3 /root/.9router/db/data.sqlite \
  "SELECT id, provider, name, isActive FROM providerConnections WHERE provider LIKE '%<query>%' OR name LIKE '%<query>%';"
sqlite3 /root/.9router/db/data.sqlite \
  "SELECT id, name, models FROM combos WHERE name LIKE '%<query>%' OR models LIKE '%<query>%';"
curl -s http://127.0.0.1:20128/v1/models | jq -r '.data[].id' | grep -i "<query>"

# Read full data JSON for one provider
sqlite3 /root/.9router/db/data.sqlite \
  "SELECT isActive, data FROM providerConnections WHERE id = '<id>';"

# Test specific provider model execution directly
curl -s http://127.0.0.1:20128/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "<model_id>", "messages": [{"role": "user", "content": "hi"}]}'
```

## Provider Deletion & Proxy Cleanup

To permanently remove unwanted providers and clean up associated proxy processes:

```bash
# 1. Delete provider rows from SQLite
sqlite3 /root/.9router/db/data.sqlite \
  "DELETE FROM providerConnections WHERE id IN ('<id1>', '<id2>');"

# 2. Stop & remove associated PM2 proxy process if applicable
pm2 stop proxy-<name> && pm2 delete proxy-<name> && pm2 save

# 3. Restart 9router service to refresh web UI & internal state
systemctl restart 9router.service
```

## Token Saver Configuration (settings table)

9router settings stored in `data.sqlite` table `settings` (`id = 1`, JSON string column `data`):

- **RTK (`rtkEnabled`)**: Compresses CLI tool output (git diff, grep, ls, tree, logs) by 60–90% before sending to LLM.
- **Headroom (`headroomEnabled`, `headroomUrl`)**: Connects to context-compression proxy (`http://localhost:8787` or Docker sidecar).
- **Caveman (`cavemanEnabled`, `cavemanLevel`: `lite` | `full` | `ultra`)**: Alters system prompt to mandate terse LLM output style (~65–87% output token reduction).
  - `lite`: Removes filler/pleasantries, keeps grammar intact.
  - `full`: Removes articles & extra connectors, fragmented sentences allowed.
  - `ultra`: Telegraphic style, maximum output compression.
- **Ponytail (`ponytailEnabled`, `ponytailLevel`: `lite` | `full` | `ultra`)**: Biases code generation toward minimal implementation (YAGNI, stdlib first).
  - `lite`: Prefers simpler code structures.
  - `full`: Enforces stdlib/native functions over external deps.
  - `ultra`: YAGNI extremist; prioritizes deleting code over adding new code.

## Media / Search / Fetch Provider Discovery

To inspect active web search, crawl, and fetch providers registered in 9router:
```bash
sqlite3 /root/.9router/db/data.sqlite \
  "SELECT id, provider, name, isActive, data FROM providerConnections WHERE provider IN ('serper', 'linkup', 'firecrawl');"
```
Registered providers include `serper`, `linkup`, and `firecrawl`. Note: Media/Search providers configured in 9router's SQLite DB serve external client endpoints routed through 9router; agent search calls can either query 9router or extract API keys directly from `providerConnections.data` JSON (e.g. `json_extract(data, '$.apiKey')`) for direct provider API calls.


## Adding a Custom OpenAI-Compatible Provider (WORKING METHOD)

**Critical discovery**: Inserting into `providerConnections` alone is NOT enough — 9router resolves the provider prefix via the **`providerNodes`** table. Without a matching `providerNodes` row, calls fail with `No active credentials for provider: <prefix>` even though models appear in `/v1/models`.

Both tables must share the SAME `provider` / `id` string.

```python
import sqlite3, json, uuid

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cur = db.cursor()

node_id = 'openai-compatible-chat-<unique-slug>'   # same value in BOTH tables
prefix  = 'myprov'
base    = 'https://api.example.com/v1'

# 1. providerNodes (REQUIRED — this is what registers the prefix)
cur.execute("""INSERT INTO providerNodes (id, type, name, data, createdAt, updatedAt)
  VALUES (?, 'openai-compatible', 'MyProvider', ?, datetime('now'), datetime('now'))""",
  (node_id, json.dumps({'prefix': prefix, 'apiType': 'chat', 'baseUrl': base})))

# 2. providerConnections (holds the credential)
cur.execute("""INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt)
  VALUES (?, ?, 'apikey', 'MyProvider', 1, ?, datetime('now'), datetime('now'))""",
  (str(uuid.uuid4()), node_id, json.dumps({
    'defaultModel': 'model-id',
    'apiKey': 'sk-xxx',
    'testStatus': 'active',
    'providerSpecificData': {'prefix': prefix, 'apiType': 'chat', 'baseUrl': base,
                             'nodeName': 'MyProvider', 'connectionProxyEnabled': False,
                             'connectionProxyUrl': '', 'connectionNoProxy': ''},
    'lastError': None, 'errorCode': None, 'lastErrorAt': None, 'backoffLevel': 0})))
db.commit()
```

Then `systemctl restart 9router.service` and test `<prefix>/<model>`.

**Note on model IDs**: `/v1/models` echoes the upstream ID verbatim (e.g. `ipeenk/ipeenkclaude/claude-opus-5`), but the callable route is `<prefix>/<defaultModel>` (e.g. `ipeenk/claude-opus-5`). Some gateways (Ipeenk) accept both the bare and namespaced upstream model name.

## Fix: isActive = 0 → "No active credentials"

```bash
sqlite3 /root/.9router/db/data.sqlite \
  "UPDATE providerConnections SET isActive = 1 WHERE id = '<id>';"
```

## Fix: redirect baseUrl to local proxy + reset errorCode

```bash
sqlite3 /root/.9router/db/data.sqlite "UPDATE providerConnections
  SET isActive = 1,
      data = json_replace(
        data,
        '$.providerSpecificData.baseUrl', 'http://127.0.0.1:<PORT>/v1',
        '$.errorCode', json('null')
      )
  WHERE id = '<id>';"
```

## Fix: reset error backoff fully

```bash
sqlite3 /root/.9router/db/data.sqlite "UPDATE providerConnections
  SET data = json_replace(
    data,
    '$.errorCode', json('null'),
    '$.backoffLevel', 0,
    '$.lastError', json('null'),
    '$.lastErrorAt', json('null')
  )
  WHERE id = '<id>';"
```

## After edits

9router reads SQLite per-request for proxy routing, **but the Web UI requires a service restart to reflect new provider entries or schema changes**:

```bash
systemctl restart 9router.service
```

Verify via models endpoint:
```bash
curl -s -H "Authorization: Bearer <key>" http://127.0.0.1:20128/v1/models
```

## Adding External Providers (Limitation)

9router stores provider credentials in `providerConnections` table, but **does not auto-expose external OpenAI-compatible providers as `/v1/models` endpoints or make them available in web UI combos** without code changes. Built-in providers (`kiro`, `antigravity`, `serper`, `linkup`, `firecrawl`) are hardcoded in the codebase.

**Database insertion alone is insufficient** — models from custom providers will NOT appear in `curl http://127.0.0.1:20128/v1/models` output even after `systemctl restart 9router.service`.

**Recommended approach:** Register external providers as **Hermes fallback providers** instead (see `references/hermes-custom-providers.md` for complete workflow). This makes them available to Hermes when 9router models fail, without requiring 9router source code modifications.

**Database insertion method (for reference only, does not expose models in 9router):**
```bash
# Insert into database
sqlite3 /root/.9router/db/data.sqlite "INSERT INTO providerConnections 
  (id, provider, authType, name, isActive, data, createdAt, updatedAt) 
VALUES 
  ('$(uuidgen)', 'lapakvip', 'apikey', 'LakapVIP Router', 1, 
   '{\"baseUrl\":\"https://router.lapakvip.com/api/v1\",\"apiKey\":\"lv-xxx\"}', 
   datetime('now'), datetime('now'));"

# Restart service
systemctl restart 9router.service
```

**Result**: Provider is stored and `isActive = 1`, but models won't appear in `curl http://127.0.0.1:20128/v1/models` unless the provider type is registered in 9router source code. Use external providers directly (not routed through 9router) or fork/extend 9router codebase.

**Model Name Conventions**: 9router models use provider prefixes. Example: `kr/claude-sonnet-4.5` (NOT `claude-sonnet-4.5`). Check available models via:
```bash
curl -s http://127.0.0.1:20128/v1/models | jq -r '.data[].id'
```

## Pitfalls

- **Genspark apiType (`chat` vs `responses`) causing Empty Completion Content**: In 9router `providerNodes` and `providerConnections.data.providerSpecificData`, setting `apiType: "responses"` for OpenAI-compatible providers like Genspark (`https://www.genspark.ai/api/llm_proxy/v1`) causes 9router to route requests via OpenAI Responses API format, resulting in HTTP 200 with empty assistant content (`"content": ""`). Ensure `apiType` is set to `"chat"` in both `providerNodes.data` and `providerConnections.data.providerSpecificData`, then restart `9router.service`.
- **OpenAI-Compatible Streaming Performance & Benchmark Context**: In 9router benchmarks, Gemini 3.7 Flash High over Antigravity achieves ~850-950 TPS after a brief thinking phase (~4-6s TTFT), whereas flagship Claude proxy models (such as `genspark/claude-fable-5` or `genspark/claude-sonnet-4-6`) stream at ~40-80 TPS with ~2.5-6s TTFT.
- **Claude Code Thinking Parameter Validation 422 (`Input should be 'enabled' or 'disabled', input: 'adaptive'`)**: When routing Claude Code requests to custom OpenAI-compatible providers (like Genspark `genspark/claude-opus-5`), Claude Code sends Anthropic native thinking format `{"thinking": {"type": "adaptive"}}`. In 9router's translation logic (`chunks/8499.js` in function `t` / module `52136`), model names matching Claude patterns trigger `case "claude-adaptive": b.thinking = { type: "adaptive" }`. Genspark's OpenAI-compatible schema strictly requires literal `'enabled'` or `'disabled'`, resulting in HTTP 422 `Request parameter validation failed`. Fix: patch 9router's `chunks/8499.js` under `case "claude-adaptive"` so that non-disabled thinking is converted to `{ type: "enabled" }`, then `systemctl restart 9router.service`.
- **Antigravity 403 Validation Required (`Verify your account to continue`)**: Google CloudCode / Antigravity periodically triggers security checkpoints for OAuth accounts under heavy use, returning HTTP 403 `VALIDATION_REQUIRED` with a Google login verification URL. Flash models (e.g. `ag/gemini-3.7-flash-high`) often remain operational while Pro/Claude models fail until the Google account is re-authenticated or verified via web login.
- **Antigravity 429 Rate Limit (`Resource has been exhausted (reset after 8s)`)**: Occurs during rapid/burst tool-calling loops when only 1 Google account is configured in the Antigravity provider. This is an upstream CloudCode PA API burst RPM limit, not monthly token quota depletion. Client agents (like Hermes) do not have built-in quota bypass mechanisms; rate limit resilience must be handled at the router/proxy tier. Fix: Add multiple Google accounts (3–5 accounts) to the `antigravity` pool in 9router so it automatically load-balances and rotates across accounts when an account hits the 8s cooldown. If an account gets locked in 9router SQLite after a 429, reset `errorCode` to null, `backoffLevel` to 0, and clear `modelLock_*` in `providerConnections.data`.
- **Antigravity Model Name Aliases & Flash Proxy (:20130)**: When external clients send OpenAI-style model IDs like `antigravity/gemini-3.7-flash-high` or `gemini-3.7-flash-high` (instead of 9router's internal `ag/gemini-3.7-flash-high`), requests fail unless translated. A dedicated proxy (`/root/proxy-antigravity-flash/index.mjs` on port 20130) intercepts requests, normalizes model names to `ag/*`, authenticates custom keys (`sk-gemini-*`), and forwards to 9router (:20128).
- **Newly Released Antigravity / Upstream Models (e.g. Gemini 3.7 / 3.8 Flash)**: Antigravity in 9router routes to Google's internal CloudCode PA API (`daily-cloudcode-pa.googleapis.com`). Models like `ag/gemini-3.8-flash-high` and `ag/gemini-3.7-flash-high` map upstream to tiered endpoints (e.g. `gemini-3.8-flash-tiered` with high reasoning budget). If an unexposed model ID (e.g. `ag/gemini-3.9-flash-high`) is queried, upstream returns 404 or 429 quota errors. Always check supported models via `curl -s http://127.0.0.1:20128/v1/models | jq -r '.data[].id'`.
- **Custom Provider Web UI Visibility**: Inserting a custom OpenAI-compatible provider (e.g., `lapakvip`, `router.lapakvip.com`) into `providerConnections` table will store credentials successfully, but the provider **will NOT appear in 9router Web UI dashboard or be usable in combos** unless the provider type is hardcoded in 9router source code. Built-in providers (`kiro`, `antigravity`, `serper`, `linkup`, `firecrawl`) are the only ones that render in UI. Custom providers can only be used as **direct Hermes fallback providers** (see `references/hermes-custom-providers.md`), not routed through 9router's combo system. To expose custom providers in 9router combos, fork and extend the codebase.
- **Cloudflare / Third-Party OpenAI Proxies User-Agent Blocking (e.g. `aspai.top`)**: Calls made via standard `urllib.request` without an explicit `User-Agent` header (defaulting to `Python-urllib/...`) may be blocked with HTTP 403 Forbidden by Cloudflare. Setting a browser/curl User-Agent header (e.g., `User-Agent: curl/7.74.0`) resolves the HTTP 403 Forbidden error on third-party API endpoints.
- **Xiaomi MiMo API (`api.xiaomimomo.com/v1`)**: `GET /v1/models` returns HTTP 200 OK with a valid API key, but `POST /v1/chat/completions` returns HTTP 402 `insufficient_balance` when account balance or free quota is zero/unclaimed.
- **Ollama Cloud Read-Only Key Behavior**: An Ollama Cloud API key may successfully validate `GET /v1/models` (HTTP 200) while returning `401 Unauthorized` on `POST /v1/chat/completions` if the key lacks completion scope or the account has zero active credits.
- **Dashboard Topology Canvas Built-in Nodes**: Nodes like `MiMo Code Free`, `OpenCode Free`, `Kiro AI`, etc. shown on the 9router dashboard central graph/flowchart are hardcoded static UI catalog presets in the Next.js frontend bundle (`1321-54939b699b5f3d07.js`). Their presence on the visual canvas does NOT mean an account or provider is active or configured in SQLite (`providerConnections`). Grayed-out nodes without cyan glowing connection lines represent inactive catalog items and consume zero resources or traffic.
- **Dashboard Provider Visibility**: Custom/OpenAI-compatible providers will NOT render in the 9router Web UI unless the `provider` column uses an established prefix pattern like `openai-compatible-chat-<uuid>`. Arbitrary string prefixes like `nvidia-<uuid>` work at the API level but won't be shown in the Web UI dashboard. Additionally, `systemctl restart 9router.service` is required for newly inserted DB rows to show up in the Web UI.
- **Prefix stripping causing 404s**: For OpenAI-compatible providers like NVIDIA (`integrate.api.nvidia.com/v1`) where model IDs contain namespaces (e.g. `z-ai/glm-5.2`), setting a `prefix` in `providerSpecificData` in 9router's SQLite DB causes 9router to strip the prefix when forwarding requests upstream (sending `glm-5.2` instead of `z-ai/glm-5.2`), leading to `404 page not found`.
- **`Error: upstream non-SSE: 200`**: Occurs when upstream returns non-SSE JSON (or compressed brotli/gzip payload) on a streaming completion request. 9router expects SSE chunks (`data: ...`), so receiving plain JSON causes 9router to treat the provider as failed and lock it for 30s. Fix: Use a local proxy to strip `accept-encoding`, set `content-type: text/event-stream`, and wrap JSON in SSE format (`data: <json>\n\ndata: [DONE]\n\n`). See `references/agentrouter-proxy.md`.
- **Invalid SSE response for non-streaming request**: Occurs when 9router sends a non-streaming request (`"stream": false`), but the proxy or upstream returns `Content-Type: text/event-stream`. Fix: Ensure your proxy script checks `reqBody.stream === true` and sets `application/json` for non-streaming requests.
- **AgentRouter Language / Content Filtering**: AgentRouter enforces strict filtering on non-English queries (e.g., Indonesian prompts), returning HTTP 400 `content-blocked`. For maximum stability when testing or routing via `agentrouter/claude-opus-5`, use English prompts.
- **Empty Stream / `no finish_reason` Error**: `Provider returned an empty stream with no finish_reason` occurs when upstream AgentRouter returns non-SSE JSON error (e.g. `401 unauthorized_client_error`, invalid token `无效的令牌`, or out-of-quota token). 9router expects SSE chunks (`data: ...`), so a plain JSON error response causes an empty stream failure. Check proxy User-Agent (`claude-cli/0.1.0 (external)`) and verify token validity with `curl -H "User-Agent: claude-cli/0.1.0 (external)" https://agentrouter.org/v1/models`.
- **Model Lock & Upstream Model Name Mismatches**: When upstream custom OpenAI/Anthropic-compatible endpoints enforce full model path identifiers (e.g., `ipeenkups/claude-opus-5` instead of `claude-opus-5`), 9router may receive a 403/404 error and place a lock under `modelLock_<model>` inside the provider's `data` JSON column (e.g., `modelLock_claude-opus-5`). Updating `defaultModel` alone is insufficient; you must also reset error states and clear model lock entries in SQLite or restart 9router:
  ```sql
  UPDATE providerConnections 
  SET data = json_set(
    json_replace(
      data,
      '$.defaultModel', 'ipeenkups/claude-opus-5',
      '$.errorCode', json('null'),
      '$.backoffLevel', 0,
      '$.lastError', json('null'),
      '$.lastErrorAt', json('null'),
      '$.testStatus', 'active'
    ),
    '$.modelLock_claude-opus-5', json('null')
  )
  WHERE id = '<id>';
  ```
- **Groq vs Grok Provider Naming**: 9router stores xAI Grok CLI accounts under provider `grok-cli` (`grok-4.6`, etc.), whereas Groq Cloud is an independent API provider (usually configured in Hermes `.env` for Whisper STT or registered as an OpenAI-compatible provider). Searching 9router SQLite for `groq` will not return xAI `grok-cli` entries.
- **Intra-Provider Multi-Account Pooling vs Cross-Provider Combo Fallback**:
  - **Intra-provider (same provider node)**: Multiple accounts registered under the same `provider` node in `providerConnections` form a pool. If account 1 hits rate limits (429) or quota depletion (402/error), 9router's connection selector (`c1`) automatically locks the failed connection and fails over to the next available active connection in the pool without client intervention.
  - **Cross-provider failover**: Direct model calls (e.g. `genspark/claude-sonnet-4-6`) cannot jump across providers when all accounts in that provider's pool are exhausted. To enable cross-provider resilience (e.g., Genspark -> Antigravity/Gemini), define an entry in the `combos` table with an ordered fallback array (e.g., `["genspark/claude-sonnet-4-6", "ag/gemini-3.7-flash-high"]`).
- **Combo Fallback & Stall Behaviour**: When an upstream provider in a 9router `combo` stalls or times out during completion streaming (e.g. server side hangs without sending SSE chunks), 9router automatically sets `testStatus: "unavailable"` and fails over to the next provider in the combo array (e.g. falling back to `gemini-gratisan`).
- **Model ID Format Mismatches (Claude Code via 9router)**: Claude Code may send model IDs with incorrect formatting (e.g., `kiro/claude-sonnet-4_5` with underscore instead of dash, or `kiro/` prefix instead of `kr/`). Upstream Kiro API rejects these with HTTP 400 `Invalid model ID. Please select a different model to continue.` The error surfaces in 9router logs as `[AUTH] kiro | all 5 accounts locked for claude-sonnet-4_5 (reset after Xs) | lastError=[400]: {"message":"Invalid model ID..."}`. **Root cause**: Client (Claude Code) sends wrong format; 9router forwards as-is; upstream rejects before alias mapping can occur. **Fix**: Edit `/root/.claude/settings.json` and set `"ANTHROPIC_DEFAULT_SONNET_MODEL": "kr/claude-sonnet-4.5"`, then restart Claude Code. Adding database aliases (`INSERT INTO combos ... kiro/claude-sonnet-4_5 -> kr/claude-sonnet-4.5`) does NOT work because upstream validation happens before 9router's combo resolution. Request normalization proxies can intercept and rewrite model IDs, but config patch is simpler. See `references/claude-code-model-id-fix.md` for config location, normalization regex pattern, and proxy implementation.
- **Verifying Model Availability via Curl**: Do not claim or report that a 9router model alias/fallback (e.g. `deepseek-gratisan` or `claude-gratisan`) is working without executing a direct `curl` completion test to `http://127.0.0.1:20128/v1/chat/completions`. Upstream providers may return hidden 429 rate limit errors (`FreeUsageLimitError`) or credential errors that only surface upon actual completion calls.
- **Genspark API Permission Flagging (HTTP 200 False Positive)**: Upstream Genspark may return HTTP 200 with the message `"This account is not permitted to use the Genspark API. If you believe this is an error, please contact support."` inside the completion payload even when account balance is positive (> 0 credits) and plan is Plus. Always inspect the generated message content when validating Genspark keys, not just HTTP status code 200.
- **Genspark Single-Model Upstream 401 False "Unavailable" Lock**: Specific upstream models on Genspark (e.g. `claude-fable-5`) may fail with `401 Unauthorized` / `litellm.AuthenticationError: AnthropicException: API key is invalid` on Genspark's backend even when the account API key is perfectly valid for other models (`claude-sonnet-4-6`, `claude-opus-5`, `gpt-5.5`). When this happens, 9router mistakes the 401 for an invalid account credential and marks the connection `testStatus: "unavailable"` with `errorCode: 401`. Fix: Verify other models directly against Genspark API; if they work, reset `testStatus: "active"`, set `errorCode: null`, `lastError: null`, `lastErrorAt: null`, `backoffLevel: 0`, and clear any `modelLock_<model>` in `providerConnections.data`, then restart `9router.service`.
- **`providerConnections` Table Insert Constraint (`authType NOT NULL`)**: When programmatically inserting records into `providerConnections` via SQLite, `authType` is a `NOT NULL` column (typically set to `'apiKey'` or `'oauth'`). Omitting `authType` will raise `sqlite3.IntegrityError: NOT NULL constraint failed: providerConnections.authType`.
- **9router Combos Routing Aliases to Underlying Models**: Combos in 9router (stored in `combos` table in `data.sqlite`) map alias names (e.g. `claude-gratisan`) to other combo names or direct models (e.g. `["gemini-gratisan"]` -> `["gemini-3.7-flash"]`). When debugging model configuration in dependent applications (like WhatsApp bots / ACP integration), inspect the `combos` table to see the true underlying target model. To add a new model alias or re-point a combo:
  ```sql
  -- Create new combo mapping alias to direct model (e.g. gemini/gemini-3.7-flash)
  INSERT INTO combos (id, name, kind, models, createdAt, updatedAt)
  VALUES ('<uuid>', 'gemini-3.7-flash', '', '["gemini/gemini-3.7-flash"]', datetime('now'), datetime('now'));

  -- Update existing combo alias target
  UPDATE combos SET models = '["gemini-3.7-flash"]' WHERE name = 'gemini-gratisan';
  ```
  After modifying `combos`, restart 9router via `systemctl restart 9router.service`.
- **AgentRouter `content-blocked` (HTTP 400)**: Upstream AgentRouter models (such as `openai-compatible-chat-agentrouter-1/claude-opus-5`) may return HTTP 400 with `content-blocked (request id: ...)` and temporary cooldowns (`reset after 30s`). When benchmarking or testing, catch HTTP 400 responses to identify upstream account/content filter blocks rather than network timeouts.
- **Claude Code CLI Non-Interactive Execution Timeout**: Invoking `claude -p "..." --print` (`/root/.local/bin/claude`) in background non-interactive scripts/subprocesses can hang or time out due to TTY / interactive session requirements. For automated benchmarking or programmatic LLM calls, send HTTP requests directly to 9router's OpenAI-compatible API endpoint (`http://127.0.0.1:20128/v1/chat/completions`).
- **Antigravity `onboardUser done but no project_id in response` (HTTP 403)**: Antigravity OAuth provider in 9router fetches an internal synthetic project ID (e.g. `synthetic-list-lt8c4`) via `v1internal:onboardUser`. Manually created Google Cloud Console projects are ignored. If `projectId` in `providerConnections.data` is empty (`""`), 9router calls to Antigravity fail with HTTP 403. Fix: ensure the Google account has accepted Gemini Code Assist / Cloud Code terms of service in GCP Console. After doing so, reset `errorCode` to null, set `testStatus: "active"`, clear `modelLock_*` fields in `data`, and restart `9router.service` (`systemctl restart 9router.service`) or re-authenticate.
- **Multi-User / Non-Root `[DB] No SQLite driver available` Error**: 9router relies on Node v22+ (for native `node:sqlite`), `bun:sqlite`, `better-sqlite3`, or `sql.js`. If Node v22 is installed inside `/root/.hermes/node` (mode `700` on `/root`), running 9router under a non-root Linux user fails with `Permission denied` on `/usr/local/bin/node` and falls back to system Node v20 (`/usr/bin/node`), triggering `[DB] No SQLite driver available`. Fix: copy Node v22 to `/opt/node22` (`chmod -R 755 /opt/node22`) and update system symlinks (`ln -sf /opt/node22/bin/node /usr/local/bin/node`).
- **9router SQLite Table Names**: Note that provider connections are stored in table `providerConnections` (not `providers`). Searching SQLite with `SELECT * FROM providers` will result in `OperationalError: no such table: providers`.
- **Non-Streaming Requests (`stream: false`)**: When making HTTP calls to 9router `/v1/chat/completions` using standard HTTP clients (such as `urllib.request` or `requests`) without an SSE stream parser, explicitly set `"stream": false` in the JSON payload. Omitting `"stream": false` may return server-sent events (`data: {...}`), causing standard `json.loads()` to fail with `Expecting value: line 1 column 1 (char 0)`.
- **Prompt Caching Usage Log Metric Discrepancy (`promptTokens` vs `cached_tokens`)**: When querying models with upstream prompt caching enabled (e.g. Anthropic / Genspark), 9router records only uncached delta tokens in the primary `promptTokens` column in `usageHistory` (which can show as low as 2 tokens). The full context length is recorded in the `tokens` JSON column under `cached_tokens` (e.g. `{"prompt_tokens": 2, "cached_tokens": 54366}`). This indicates active prompt cache hits, not prompt truncation.
- **Local API Authentication Key Retrieval (`apiKeys` Table)**: When testing 9router endpoints locally (`http://127.0.0.1:20128/v1/chat/completions`), requests require a valid Bearer token. Query the active API key directly from SQLite: `sqlite3 /root/.9router/db/data.sqlite "SELECT key FROM apiKeys WHERE isActive = 1 LIMIT 1;"`.
- `isActive = 0` is the silent culprit for `No active credentials` — check this FIRST before assuming auth/key issues.
- Non-null `errorCode` keeps provider in backoff even after fixing auth — always reset it.
- Use `json_replace()` for surgical JSON edits, not string concat.
- WAL mode is active on the DB; quick single-row updates during off-peak are safe.

## References

- `references/codex-9router-integration.md` — Setup, TOML configuration (`wire_api = "responses"`, `experimental_bearer_token`), and agent role structure for running Codex CLI with 9router local endpoints.
- `references/codex-usage-monitoring.md` — Inspect Codex (ChatGPT Plus/Team/Pro) real-time rate limit windows (5h/7d) via `wham/usage` with required headers (`chatgpt-account-id`, User-Agent `codex-cli/0.1.0 (external)`).
- `references/9router-service-autostart.md` — Setup guides for 9router daemon auto-start & background persistence across OS platforms (systemd, PM2, Windows Task Scheduler / NSSM).
- `references/9router-custom-server-architecture.md` — Systemd service config, custom-server HTTP/2 wrapper, token refresher, and direct Next.js startup flow.
- `references/kiro-auth-import.md` — Kiro CLI SQLite token extraction (`data.sqlite3`), JSON payload structure, and 9router import endpoints.
- `references/9router-bulk-lock-reset.md` — Python script & SQL query for clearing batch model locks (`modelLock_*`) and resetting error backoffs on 9router custom providers when upstream accounts recover.
- `references/reverse-proxy-llm-benchmarking.md` — Reverse-proxy LLM authenticity probing, CLI wrapper detection (Kiro/Amazon Q prompt overhead ~7k tokens), and API metadata inspection.
- `references/agentrouter-proxy.md` — AgentRouter 401 bypass via header proxy (8743), token model scope limitation causing 9router UI validation 500 error, direct SQLite fix.
- `references/chatgptweb-proxy.md` — Setup, process location (`/root/proxy-chatgptweb/index.mjs` on :8745), 9router integration details, and verification commands for OpenAI-Web-Luna (`cgweb/gpt-5.6-luna`).
- `references/lumosel-proxy.md` — header-injection proxy for Lumosel (403 bypass via `claude-code/1.0.53` UA), pm2 setup, proxy script.
- `references/opencode-zen.md` — OpenCode Zen free models benchmark (`deepseek-v4-flash-free`, `ling-3.0-flash-free`), 1M context window, UA requirements (`opencode/1.0.0`), TPS stats.
- `references/nvidia-nemotron-api.md` — NVIDIA integrate API testing script & active Nemotron model status (Ultra 550B, Super 120B).
- `references/opencode-zen-free-models.md` — OpenCode Zen free models benchmark, context windows (1M tokens for DeepSeek V4 Flash / MiMo 2.5), and User-Agent requirements (`opencode/1.0.0`).
- `references/opencode-free-models.md` — OpenCode free tier endpoints, User-Agent requirements (HTTP 403 fix), and benchmark performance metrics (TPS/latency).
- `references/benchmark-llm-endpoints.md` — Python script for comparing OpenAI-compatible endpoints (latency, TTFT, TPS); includes working example comparing 9router local vs external providers.
- `references/grok-register-9router-integration.md` — Integration architecture between Grok-Register (Charles-0509/Grok-Register) and 9router SQLite `providerConnections` for xAI token sync.
- `references/genspark-and-zai-endpoints.md` — Genspark LLM proxy setup & multi-account pool; Z.AI GLM Coding Plan vs Standard API endpoint separation.
- `references/genspark-llm-proxy-integration.md` — Genspark LLM proxy (`https://www.genspark.ai/api/llm_proxy/v1`) integration, CLI key extraction (`~/.genspark-tool-cli/config.json`), multi-account pooling, and Claude flagship model testing.
- `references/hermes-custom-providers.md` — Complete workflow for adding external OpenAI-compatible providers to Hermes (not 9router) as fallback models using `hermes config set` commands.

## Scripts

- `scripts/benchmark_llm.py` — Benchmark multiple OpenAI-compatible endpoints (latency, TTFT, TPS). Edit CONFIGS array to test your endpoints.
