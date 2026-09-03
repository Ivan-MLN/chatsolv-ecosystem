# Custom OpenAI-Compatible Provider Registration & Routing in 9router

## Prerequisites & Table Requirements

In 9router, custom OpenAI-compatible endpoints require matching entries in two SQLite tables (`/root/.9router/db/data.sqlite`):
1. `providerNodes`: Maps the provider prefix and base URL.
2. `providerConnections`: Holds credentials, `defaultModel`, and status metadata.

Both tables MUST use the same `id` string (e.g. `openai-compatible-chat-<unique-id>`).

## Python Registration Script

```python
import sqlite3, json, uuid

db = sqlite3.connect('/root/.9router/db/data.sqlite')
cursor = db.cursor()

node_id = f"openai-compatible-chat-{str(uuid.uuid4())}"
prefix = "myprefix"
base_url = "https://api.example.com/v1"
api_key = "sk-xxx"
default_model = "claude-opus-5"

# 1. Register in providerNodes (REQUIRED for prefix resolution)
node_config = {
    "prefix": prefix,
    "apiType": "chat",
    "baseUrl": base_url
}
cursor.execute("""
    INSERT INTO providerNodes (id, type, name, data, createdAt, updatedAt)
    VALUES (?, 'openai-compatible', 'MyProvider', ?, datetime('now'), datetime('now'))
""", (node_id, json.dumps(node_config)))

# 2. Register in providerConnections (Credential & state)
connection_data = {
    "defaultModel": default_model,
    "apiKey": api_key,
    "testStatus": "active",
    "providerSpecificData": {
        "prefix": prefix,
        "apiType": "chat",
        "baseUrl": base_url,
        "nodeName": "MyProvider",
        "connectionProxyEnabled": False,
        "connectionProxyUrl": "",
        "connectionNoProxy": ""
    },
    "lastError": None,
    "errorCode": None,
    "lastErrorAt": None,
    "backoffLevel": 0
}
cursor.execute("""
    INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt)
    VALUES (?, ?, 'apikey', 'MyProvider', 1, ?, datetime('now'), datetime('now'))
""", (str(uuid.uuid4()), node_id, json.dumps(connection_data)))

db.commit()
```

After modifying SQLite, restart the service:
```bash
systemctl restart 9router.service
```

## Routing & Model Aliasing

- **Direct Call**: Call `<prefix>/<default_model>` (e.g., `myprefix/claude-opus-5`).
- **Combos Integration**: Insert into `combos` table to map alias names or group providers into fallback chains.
- **Returned vs Fallback**:
  - `Returned`: Upstream primary target answered directly.
  - `Fallback`: 9router or upstream proxy dynamically re-routed the request to a fallback model because the requested model was unavailable or locked.
