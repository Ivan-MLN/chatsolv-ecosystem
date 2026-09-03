# Genspark LLM Proxy Provider Integration in 9router

## Overview
Genspark provides an OpenAI-compatible LLM proxy endpoint supporting various models (Claude 4.5/4.6/4.7/5, GPT-5/5.5/5.6-luna, DeepSeek V4 Pro, Kimi K3, etc.).

- **Base URL**: `https://www.genspark.ai/api/llm_proxy/v1`
- **CLI / Auth Storage**: `@genspark/cli` stores credentials in `~/.genspark-tool-cli/config.json` under the key `api_key`.
- **9router Prefix**: `genspark`

## Registration Script Example

```python
import sqlite3, json, uuid

cfg = json.load(open('/root/.genspark-tool-cli/config.json'))
api_key = cfg['api_key']
base_url = 'https://www.genspark.ai/api/llm_proxy/v1'
prefix = 'genspark'
node_name = 'Genspark'
node_id = f'openai-compatible-chat-{str(uuid.uuid4())}'

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cursor = db.cursor()

# 1. providerNodes (Registers prefix and routing)
node_config = {
    'prefix': prefix,
    'apiType': 'chat',
    'baseUrl': base_url
}
cursor.execute("""
    INSERT INTO providerNodes (id, type, name, data, createdAt, updatedAt)
    VALUES (?, 'openai-compatible', ?, ?, datetime('now'), datetime('now'))
""", (node_id, node_name, json.dumps(node_config)))

# 2. providerConnections (Holds API key and state; authType NOT NULL constraint)
conn_id = str(uuid.uuid4())
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
cursor.execute("""
    INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt)
    VALUES (?, ?, 'apiKey', ?, 1, ?, datetime('now'), datetime('now'))
""", (conn_id, node_id, node_name, json.dumps(connection_data)))

db.commit()
```

## Troubleshooting & Known Quirks

### 1. Upstream Permission Flagging (HTTP 200 False Positive)
When an account's developer API access is suspended or restricted by Genspark, requests to `https://www.genspark.ai/api/llm_proxy/v1/chat/completions` and direct CLI tool calls (`/api/tool_cli/web_search`) return HTTP status `200 OK` with JSON error payload:
```json
{
  "choices": [{
    "message": {
      "content": "This account is not permitted to use the Genspark API. If you believe this is an error, please contact support."
    }
  }]
}
```
`gsk me` will still report `status: "ok"`, `plan: "plus"`, and active credits (> 0).
Always inspect `response.json()['choices'][0]['message']['content']` during key health verification.

### 2. Upstream Single-Model 401 & 9router "Unavailable" Backoff
Certain models (such as `claude-fable-5`) may fail upstream on Genspark's litellm routing with:
```json
{"error":{"message":"Error code: 401 - {'error': {'message': 'litellm.AuthenticationError: AnthropicException - API key is invalid.'}}","code":"server_error","status":401}}
```
This is an upstream Genspark-to-Anthropic configuration error, NOT an invalid Genspark user API key.
However, 9router receives HTTP 401 and treats the entire provider connection as invalid, updating `testStatus` to `"unavailable"` and applying an error lock.

**Recovery Script:**
```python
import sqlite3, json

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cursor = db.cursor()
row = cursor.execute('SELECT id, data FROM providerConnections WHERE name LIKE "%Genspark%"').fetchone()
if row:
    cid, data_str = row
    data = json.loads(data_str)
    data['testStatus'] = 'active'
    data['errorCode'] = None
    data['lastError'] = None
    data['lastErrorAt'] = None
    data['backoffLevel'] = 0
    for k in list(data.keys()):
        if k.startswith('modelLock_'):
            data[k] = None
    cursor.execute('UPDATE providerConnections SET data = ?, updatedAt = datetime("now") WHERE id = ?', (json.dumps(data), cid))
    db.commit()
```
Followed by `systemctl restart 9router.service`.

## Service Restart & Verification

```bash
systemctl restart 9router.service

# List models
curl -s http://127.0.0.1:20128/v1/models | jq -r '.data[].id' | grep -i genspark

# Test completion
curl -s http://127.0.0.1:20128/v1/chat/completions \
  -H "Authorization: Bearer <9router_key>" \
  -H "Content-Type: application/json" \
  -d '{"model": "genspark/claude-sonnet-4-6", "messages": [{"role": "user", "content": "ping"}], "stream": false}'
```
