# Codex / ChatGPT Backend Usage & Quota Inspection

How to check Codex (ChatGPT Plus / Team / Pro) real-time rate limit windows and usage directly from 9router credentials.

## Credential Extraction

Codex tokens are stored in `providerConnections` in 9router SQLite:

```python
import sqlite3, json, requests

db = sqlite3.connect("/root/.9router/db/data.sqlite")
cur = db.cursor()
cur.execute("SELECT data FROM providerConnections WHERE provider='codex' OR id LIKE '%codex%';")
row = cur.fetchone()
data = json.loads(row[0])
acc_token = data.get("accessToken")
account_id = data.get("providerSpecificData", {}).get("chatgptAccountId")
```

## Upstream Usage Endpoint

Endpoint: `https://chatgpt.com/backend-api/wham/usage`

### Required Headers

Calling this endpoint without `chatgpt-account-id` or with default Python `User-Agent` returns **HTTP 403 Forbidden**.

```python
headers = {
    "Authorization": f"Bearer {acc_token}",
    "chatgpt-account-id": account_id,
    "User-Agent": "codex-cli/0.1.0 (external)"
}

r = requests.get("https://chatgpt.com/backend-api/wham/usage", headers=headers, timeout=10)
usage = r.json()
```

## Response Structure

```json
{
  "user_id": "user-xxx",
  "account_id": "xxx",
  "email": "user@example.com",
  "plan_type": "plus",
  "rate_limit": {
    "allowed": true,
    "limit_reached": false,
    "primary_window": {
      "used_percent": 5,
      "limit_window_seconds": 18000,
      "reset_after_seconds": 16749,
      "reset_at": 1787685238
    },
    "secondary_window": {
      "used_percent": 1,
      "limit_window_seconds": 604800,
      "reset_after_seconds": 603549,
      "reset_at": 1788272038
    }
  },
  "credits": {
    "has_credits": false,
    "unlimited": false,
    "overage_limit_reached": false,
    "balance": "0"
  }
}
```

- `primary_window`: 5-hour rolling window (`limit_window_seconds = 18000`).
- `secondary_window`: 7-day rolling window (`limit_window_seconds = 604800`).
- `reset_at`: Unix epoch seconds for when the window resets.
