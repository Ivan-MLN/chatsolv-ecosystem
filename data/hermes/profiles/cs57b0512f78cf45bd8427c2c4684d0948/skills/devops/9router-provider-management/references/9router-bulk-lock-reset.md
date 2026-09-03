# Bulk Clearing 9router Model Locks & Error Backoffs

When an upstream custom provider (e.g. Freebuff, Ipeenk, AgentRouter, or external OpenAI-compatible endpoints) returns transient HTTP 403/400 errors or non-SSE stream mismatches, 9router automatically places individual model locks inside `providerConnections.data` (e.g. `modelLock_openai/gpt-5.6-luna-max` or `modelLock_claude-opus-5`).

Even after upstream recovers or proxy is fixed, 9router will continue rejecting requests with `[provider/model] Unavailable (reset after Xs)` or `Provider error (reset after 30s)` until the model lock expires or is manually cleared.

## Python Script to Clear All Model Locks for a Provider

```python
import sqlite3, json

db_path = '/root/.9router/db/data.sqlite'
provider_name = 'AgentRouter'  # Or any provider name ('lol', 'Ipeenk AI', etc.)

db = sqlite3.connect(db_path)
cursor = db.cursor()

cursor.execute("SELECT id, data FROM providerConnections WHERE name = ?", (provider_name,))
row = cursor.fetchone()
if row:
    conn_id, data_str = row
    d = json.loads(data_str)
    
    # Remove all modelLock_* keys
    keys_to_del = [k for k in d.keys() if k.startswith('modelLock_')]
    for k in keys_to_del:
        del d[k]
        
    d['errorCode'] = None
    d['lastError'] = None
    d['lastErrorAt'] = None
    d['backoffLevel'] = 0
    d['testStatus'] = 'active'
    
    cursor.execute("UPDATE providerConnections SET data = ? WHERE id = ?", (json.dumps(d), conn_id))
    db.commit()
    print(f"Cleared {len(keys_to_del)} model locks and reset backoff for '{provider_name}'.")

# Always restart 9router after DB edits
# systemctl restart 9router.service
```

## SQL Equivalent for Single Model Unlocking

```sql
UPDATE providerConnections 
SET data = json_remove(
  json_replace(
    data,
    '$.errorCode', json('null'),
    '$.backoffLevel', 0,
    '$.lastError', json('null'),
    '$.lastErrorAt', json('null'),
    '$.testStatus', 'active'
  ),
  '$.modelLock_claude-opus-5'
)
WHERE name = 'AgentRouter';
```
