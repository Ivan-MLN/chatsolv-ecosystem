# Genspark LLM Proxy Integration in 9router

## Overview
Genspark provides an OpenAI-compatible LLM proxy endpoint supporting Anthropic Claude models (including `claude-fable-5`, `claude-opus-5`, `claude-sonnet-5`, `claude-sonnet-4-6`, `claude-haiku-4-5`), GPT models, and reasoning models.

- **Base URL**: `https://www.genspark.ai/api/llm_proxy/v1`
- **CLI / Config Location**: Installed via `npm install -g @genspark/cli`. API key stored at `~/.genspark-tool-cli/config.json` (key field: `api_key`).
- **Models Endpoint**: `GET https://www.genspark.ai/api/llm_proxy/v1/models`

## Checking Account Info & Credit Balance

Genspark API keys can be probed for credit balance, user email, and plan via the CLI command `gsk me` (or `genspark me` / `genspark login-info`):

```bash
GSK_API_KEY="gsk-xxx" gsk me
```
Output JSON format:
```json
{
  "status": "ok",
  "message": "success",
  "data": {
    "email": "user@example.com",
    "name": "User Name",
    "plan": "plus",
    "personal_plan": "plus",
    "credit_balance": 10100
  }
}
```

When an account credit balance reaches `0` (or returns `credit_exhausted`), remove the exhausted connection from 9router SQLite `providerConnections` table and restart `9router.service` so 9router does not attempt routing traffic to empty accounts.

## Upstream Permission Flagging / Ban Behavior
Genspark may flag or restrict accounts from using LLM proxy endpoints (`/api/llm_proxy/v1/*`) or CLI tools even when `credit_balance > 0` and plan is `plus`:
- **HTTP 200 with Block Message**: Calling `/api/llm_proxy/v1/chat/completions` returns HTTP 200, but the completion content is:
  `"This account is not permitted to use the Genspark API. If you believe this is an error, please contact support."`
- **CLI Tool Endpoints Ban**: Calling `/api/tool_cli/<tool>` (e.g. `web_search`) with `X-Api-Key` and `X-GSK-CLI-Version` returns:
  `{"status":"error","message":"This account is not permitted to use the Genspark API. If you believe this is an error, please contact support."}`
  This confirms the ban is account-wide across both LLM proxy and Tool CLI APIs, not a header/client mismatch.
- **Model Validation vs Completion Ban**: Even on banned accounts, Genspark's upstream model name validator remains active. Sending an unsupported model name returns HTTP 400 with the live catalog of allowed models (`Allowed models: gpt-5, claude-opus-5, ...`), but any valid model name returns the 200 block message.
- **Zero Balance CLI Constraint**: For accounts with 0 credits, CLI commands error with:
  `"The Genspark CLI requires a paid plan or a credit balance of at least 500."`

When testing Genspark connections in 9router, always verify actual response content rather than checking HTTP status code alone.

## Upstream Model 401 & False "Unavailable" Provider Lock
Individual models on Genspark (e.g., `claude-fable-5`) may fail upstream on Genspark's backend with:
```
401 Unauthorized - litellm.AuthenticationError: AnthropicException: API key is invalid
```
When this occurs on a single model, 9router mistakes the 401 error as an entire account authentication failure, marking the provider connection as `testStatus: "unavailable"` with `errorCode: 401` and an active backoff countdown timer (`⏱ Xs`) in the Web UI dashboard.

### Diagnosis & Recovery:
1. Probe working models directly against Genspark API (`claude-sonnet-4-6`, `claude-opus-5`, `gpt-5.5`).
2. If working models succeed, clear the false error state in 9router SQLite:
```python
import sqlite3, json

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cursor = db.cursor()

cursor.execute('''
    UPDATE providerConnections 
    SET data = json_set(
        data, 
        '$.testStatus', 'active', 
        '$.errorCode', null, 
        '$.lastError', null, 
        '$.lastErrorAt', null, 
        '$.backoffLevel', 0,
        '$.modelLock_claude-fable-5', null
    ),
    updatedAt = datetime('now')
    WHERE name LIKE '%Genspark%'
''')
db.commit()
```
3. Restart service: `systemctl restart 9router.service`.
4. Instruct upstream clients/combos to avoid routing to broken models (`claude-fable-5`).

## Claude Code & Thinking Schema Compatibility (422 Error Fix)
When Claude Code connects to 9router (`/v1/messages`), it sends:
```json
{
  "thinking": { "type": "adaptive" }
}
```
Genspark's OpenAI-compatible backend schema enforces `body.thinking.type` to be strictly `"enabled"` or `"disabled"`. If 9router forwards `"adaptive"`, Genspark returns:
```json
{
  "status": -2,
  "message": "Request parameter validation failed",
  "errors": [
    {
      "type": "literal_error",
      "loc": ["body", "thinking", "type"],
      "msg": "Input should be 'enabled' or 'disabled'",
      "input": "adaptive"
    }
  ]
}
```

### Fix in 9router Transformer (`chunks/8499.js`)
In 9router's translation chunk (`/usr/local/lib/node_modules/9router/app/.next-cli-build/server/chunks/8499.js`), modify `case "claude-adaptive":`:
```javascript
// Before
case "claude-adaptive": {
  if (h && i) { b.thinking = { type: "disabled" }; break; }
  b.thinking = { type: "adaptive" };
  let a = n(j);
  b.output_config = { effort: "xhigh" === a ? "high" : a };
  break;
}

// After (maps to {type: "enabled"} for compatibility with OpenAI proxies like Genspark)
case "claude-adaptive": {
  if (h && i) { b.thinking = { type: "disabled" }; break; }
  b.thinking = { type: "enabled" };
  let a = n(j);
  b.output_config = { effort: "xhigh" === a ? "high" : a };
  break;
}
```
After editing, restart `systemctl restart 9router.service`.

## 9router Registration

Custom OpenAI-compatible provider in 9router requires:
1. `providerNodes` entry with `prefix: "genspark"` and `baseUrl: "https://www.genspark.ai/api/llm_proxy/v1"`
2. One or more `providerConnections` rows linked to the same `node_id` for multi-account pooling and automatic failover.

### Python Registration Snippet

```python
import sqlite3, json, uuid

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cursor = db.cursor()

prefix = 'genspark'
node_name = 'Genspark'
node_id = f'openai-compatible-chat-{str(uuid.uuid4())}'
base_url = 'https://www.genspark.ai/api/llm_proxy/v1'
api_key = 'gsk-xxx'

# 1. Register Node
node_config = {
    'prefix': prefix,
    'apiType': 'chat',
    'baseUrl': base_url
}
cursor.execute('''
    INSERT INTO providerNodes (id, type, name, data, createdAt, updatedAt)
    VALUES (?, 'openai-compatible', ?, ?, datetime('now'), datetime('now'))
''', (node_id, node_name, json.dumps(node_config)))

# 2. Register Account Connection(s)
connection_data = {
    'defaultModel': 'claude-sonnet-4-6',
    'apiKey': api_key,
    'testStatus': 'active',
    'providerSpecificData': {
        'prefix': prefix,
        'apiType': 'chat',
        'baseUrl': base_url,
        'nodeName': node_name,
        'connectionProxyEnabled': False,
        'connectionProxyUrl': '',
        'connectionNoProxy': ''
    },
    'lastError': None,
    'errorCode': None,
    'lastErrorAt': None,
    'backoffLevel': 0
}
cursor.execute('''
    INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt)
    VALUES (?, ?, 'apikey', ?, 1, ?, datetime('now'), datetime('now'))
''', (str(uuid.uuid4()), node_id, node_name, json.dumps(connection_data)))

db.commit()
```

### Adding Additional Accounts & Auto-Pruning Exhausted Accounts (Bulk Script)

To add accounts to the pool, verify balance via `genspark me`, prune 0-credit accounts, and restart `9router.service`:

```javascript
// Run via nael-ai eval MCP (:5778) or node script
const sqlite3Mod = await import("node:sqlite");
const cryptoMod = await import("node:crypto");
const cpMod = await import("node:child_process");

const db = new sqlite3Mod.DatabaseSync("/root/.9router/db/data.sqlite");
const newKeys = ["gsk-key1...", "gsk-key2..."];

const nodeRow = db.prepare("SELECT id, data FROM providerNodes WHERE id LIKE ?").get("openai-compatible-chat-%");
const nodeId = nodeRow.id;
const nodeData = JSON.parse(nodeRow.data);

const existingConns = db.prepare("SELECT id, name, data FROM providerConnections WHERE provider = ?").all(nodeId);
const allKeys = [];
for (const c of existingConns) {
  const d = JSON.parse(c.data);
  allKeys.push({ id: c.id, name: c.name, apiKey: d.apiKey });
}

for (const sk of newKeys) {
  if (!allKeys.some(k => k.apiKey === sk)) {
    allKeys.push({ id: cryptoMod.randomUUID(), name: `Genspark (Akun ${allKeys.length + 1})`, apiKey: sk });
  }
}

for (let i = 0; i < allKeys.length; i++) {
  const item = allKeys[i];
  const connData = {
    defaultModel: "claude-sonnet-4-6",
    apiKey: item.apiKey,
    testStatus: "active",
    providerSpecificData: {
      prefix: nodeData.prefix,
      apiType: nodeData.apiType,
      baseUrl: nodeData.baseUrl,
      nodeName: item.name,
      connectionProxyEnabled: false,
      connectionProxyUrl: "",
      connectionNoProxy: ""
    },
    errorCode: null,
    backoffLevel: 0,
    "modelLock_claude-opus-5": null,
    lastError: null,
    lastErrorAt: null
  };

  const exists = db.prepare("SELECT id FROM providerConnections WHERE id = ?").get(item.id);
  if (exists) {
    db.prepare("UPDATE providerConnections SET name = ?, isActive = 1, data = ?, updatedAt = datetime() WHERE id = ?")
      .run(item.name, JSON.stringify(connData), item.id);
  } else {
    db.prepare("INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt) VALUES (?, ?, ?, ?, 1, ?, datetime(), datetime())")
      .run(item.id, nodeId, "apikey", item.name, JSON.stringify(connData));
  }
}

// Prune exhausted accounts
const allConns = db.prepare("SELECT id, data FROM providerConnections WHERE provider = ?").all(nodeId);
for (const c of allConns) {
  const d = JSON.parse(c.data);
  try {
    const out = cpMod.execSync(`GSK_API_KEY="${d.apiKey}" genspark me`).toString();
    const cred = JSON.parse(out)?.data?.credit_balance ?? 0;
    if (cred <= 0) {
      db.prepare("DELETE FROM providerConnections WHERE id = ?").run(c.id);
    }
  } catch (err) {}
}

cpMod.execSync("systemctl restart 9router.service");
```

## Model Verification
Test call format:
```bash
curl -s http://127.0.0.1:20128/v1/chat/completions \
  -H "Authorization: Bearer <9ROUTER_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"model": "genspark/claude-sonnet-5", "messages": [{"role": "user", "content": "hi"}], "stream": false}'
```
