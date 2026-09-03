# Kiro CLI Auth Extraction & 9router Integration

## Kiro CLI Auth Location

On Linux VPS, `kiro-cli` stores session and OAuth tokens in its local SQLite database:

- Database: `/root/.local/share/kiro-cli/data.sqlite3`
- Table: `auth_kv`
- Key: `kirocli:social:token`

Data JSON structure:
- `access_token`: Active AWS CodeWhisperer/Kiro bearer token (`aoaAAAAAG...`)
- `refresh_token`: OAuth refresh token (`aorAAAAAG...`)
- `profile_arn`: AWS CodeWhisperer profile ARN (`arn:aws:codewhisperer:...`)
- `expires_at`: Token expiry ISO string
- `provider`: Social OAuth provider (`google`, `builder-id`, etc.)

## Direct 9router SQLite Provider Insertion (Bypassing CLI Auth Token Check)

9router's Next.js OAuth API endpoints (`/api/oauth/kiro/*`) enforce `x-9r-cli-token` header validation (checks for token in `/root/.9router/9r-cli-auth`). If the CLI token file does not exist, requests return HTTP 403 `{"error":"Local only: CLI token required"}`.

To bypass this and connect Kiro directly:

```python
import sqlite3, json, uuid
from datetime import datetime, timezone

# 1. Read Kiro CLI credentials from data.sqlite3
conn_kiro = sqlite3.connect('/root/.local/share/kiro-cli/data.sqlite3')
cur_kiro = conn_kiro.cursor()
cur_kiro.execute('SELECT value FROM auth_kv WHERE key = "kirocli:social:token";')
kiro_data = json.loads(cur_kiro.fetchone()[0])

access_token = kiro_data.get('access_token') or kiro_data.get('accessToken')
refresh_token = kiro_data.get('refresh_token') or kiro_data.get('refreshToken')
profile_arn = kiro_data.get('profile_arn') or kiro_data.get('profileArn') or ''
expires_at = kiro_data.get('expires_at') or kiro_data.get('expiresAt') or ''

now_iso = datetime.now(timezone.utc).isoformat().replace('+00:00', 'Z')

provider_data = {
    'accessToken': access_token,
    'refreshToken': refresh_token,
    'expiresAt': expires_at or now_iso,
    'authMethod': 'social',
    'providerSpecificData': {
        'profileArn': profile_arn,
        'authMethod': 'social'
    }
}

# 2. Insert/Update in 9router database
conn_9r = sqlite3.connect('/root/.9router/db/data.sqlite')
cur_9r = conn_9r.cursor()

cur_9r.execute('SELECT id FROM providerConnections WHERE provider = "kiro";')
existing = cur_9r.fetchall()
if existing:
    for (eid,) in existing:
        cur_9r.execute(
            'UPDATE providerConnections SET data = ?, isActive = 1, updatedAt = ? WHERE id = ?',
            (json.dumps(provider_data), now_iso, eid)
        )
else:
    conn_id = str(uuid.uuid4())
    cur_9r.execute(
        'INSERT INTO providerConnections (id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)',
        (conn_id, 'kiro', 'oauth', 'kiro-cli', 'kiro-cli@local', 1, 1, json.dumps(provider_data), now_iso, now_iso)
    )

conn_9r.commit()
```

## Refreshing Kiro Tokens via Kiro Desktop Auth Endpoint

Kiro social OAuth tokens can be refreshed directly against the auth endpoint:

- **Endpoint**: `POST https://prod.us-east-1.auth.desktop.kiro.dev/refreshToken`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Accept": "application/json",
    "User-Agent": "kiro-cli/1.0.0"
  }
  ```
- **Body**: `{"refreshToken": "<aorAAAAAG...>"}`
- **Response**: `{"accessToken": "<aoa...>", "expiresIn": 3600, "profileArn": "..."}`

## 9router Kiro Import Endpoints (Requires CLI Auth Token)

9router provides native endpoints to register and connect Kiro accounts when called with header `x-9r-cli-token`:

### 1. Direct Refresh Token Import Endpoint
- **URL**: `POST http://127.0.0.1:20128/api/oauth/kiro/import`
- **Payload**:
  ```json
  {
    "refreshToken": "aorAAAAAG...",
    "profileArn": "arn:aws:codewhisperer:..."
  }
  ```
- **Behavior**: 9router exchanges the `refreshToken` with AWS SSO endpoint, extracts user email from JWT, registers `kiro` provider in `providerConnections` with `authType: "oauth"`, and sets `testStatus: "active"`.

### 2. Auto-Import from AWS SSO Cache
- **URL**: `POST http://127.0.0.1:20128/api/oauth/kiro/auto-import`
- **Behavior**: Scans `~/.aws/sso/cache/*.json` for tokens matching prefix `aorAAAAAG` and profile ARN from `~/.config/Kiro/User/globalStorage/kiro.kiroagent/profile.json`.

### 3. CLI Proxy Import
- **URL**: `POST http://127.0.0.1:20128/api/oauth/kiro/import-cli-proxy`
- Used for importing external IDP configurations with explicit `token_endpoint`, `client_id`, `scopes`, etc.
