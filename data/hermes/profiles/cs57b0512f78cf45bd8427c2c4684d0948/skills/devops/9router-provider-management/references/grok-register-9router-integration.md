# Grok-Register & xAI Token Integration with 9router

Integration pattern between [Charles-0509/Grok-Register](https://github.com/Charles-0509/Grok-Register) and 9router SQLite provider database.

## Architecture

1. **Grok-Register Runtime**:
   - Binary CLI: `/usr/local/bin/grok`
   - Data & Config: `~/.grok/config.env`, `~/.grok/outputs/<run-id>/`
   - Output formats: `grok2api/tokens.txt` (single-line session tokens), `CPA/*.json`, `SSO/*.json`.
   - Clearance stack: WARP socks proxy (`127.0.0.1:40000`), Privoxy HTTP proxy (`127.0.0.1:40080`), FlareSolverr (`127.0.0.1:8191`).
   - Requires Docker Compose v2 (v1 fails on variable interpolation syntax `${VAR:-default}`).

2. **9router xAI Provider Connection Schema**:
   - Database: `/root/.9router/db/data.sqlite`
   - Table: `providerConnections`
   - Provider identifier: `xai` (or `grok-cli` for CLI OAuth)
   - `authType`: `apiKey`
   - `data` JSON payload:
     ```json
     {
       "apiKey": "<sso_token_or_api_key>",
       "testStatus": "active",
       "lastRefreshAt": "2026-08-23T17:00:00.000Z"
     }
     ```

## Synchronization Script (`/usr/local/bin/grok-sync-9router`)

```python
#!/usr/bin/env python3
import json
import sqlite3
import sys
import uuid
from datetime import datetime
from pathlib import Path

DB_PATH = Path.home() / ".9router" / "db" / "data.sqlite"
GROK_OUTPUTS_DIR = Path.home() / ".grok" / "outputs"

def get_latest_output_dir():
    if not GROK_OUTPUTS_DIR.exists():
        return None
    dirs = [d for d in GROK_OUTPUTS_DIR.iterdir() if d.is_dir()]
    return sorted(dirs, key=lambda x: x.name, reverse=True)[0] if dirs else None

def sync_tokens_to_9router(output_dir=None):
    if not output_dir:
        output_dir = get_latest_output_dir()
    if not output_dir or not DB_PATH.exists():
        return 0

    tokens_txt = output_dir / "grok2api" / "tokens.txt"
    cpa_dir = output_dir / "CPA"
    found_accounts = []

    if cpa_dir.exists():
        for cpa_file in cpa_dir.glob("*.json"):
            try:
                with open(cpa_file, "r") as f:
                    data = json.load(f)
                    if isinstance(data, list):
                        found_accounts.extend(data)
                    elif isinstance(data, dict):
                        found_accounts.append(data)
            except Exception:
                pass

    if not found_accounts and tokens_txt.exists():
        with open(tokens_txt, "r") as f:
            for idx, token in enumerate(f.read().splitlines()):
                token = token.strip()
                if token:
                    found_accounts.append({
                        "token": token,
                        "email": f"grok-user-{idx+1}@auto.x.ai"
                    })

    conn = sqlite3.connect(str(DB_PATH))
    cursor = conn.cursor()
    imported = 0
    now_iso = datetime.utcnow().isoformat() + "Z"

    for acc in found_accounts:
        token = acc.get("token") or acc.get("session_token") or acc.get("api_key") or acc.get("sso_token")
        email = acc.get("email") or acc.get("username") or f"grok-{uuid.uuid4().hex[:6]}"
        if not token:
            continue

        cursor.execute("SELECT id FROM providerConnections WHERE provider = 'xai' AND name = ?", (email,))
        row = cursor.fetchone()
        conn_data = {"apiKey": token, "testStatus": "active", "lastRefreshAt": now_iso}

        if row:
            cursor.execute(
                "UPDATE providerConnections SET data = ?, isActive = 1, updatedAt = ? WHERE id = ?",
                (json.dumps(conn_data), now_iso, row[0])
            )
        else:
            conn_id = str(uuid.uuid4())
            cursor.execute(
                "INSERT INTO providerConnections (id, provider, authType, name, isActive, data, createdAt, updatedAt) VALUES (?, ?, ?, ?, 1, ?, ?, ?)",
                (conn_id, "xai", "apiKey", email, json.dumps(conn_data), now_iso, now_iso)
            )
        imported += 1

    conn.commit()
    conn.close()

    if imported > 0:
        import subprocess
        try:
            subprocess.run(["systemctl", "restart", "9router.service"], check=False)
        except Exception:
            pass

    return imported
```

## Troubleshooting & Pitfalls

- **Next.js Action ID 404 on xAI Registration Tooling**: Grok/xAI web signup frequently invalidates Next.js Server Action IDs (`404 Server action not found`). When automated register tools fallback to hardcoded action IDs (`+default_action`), registration fails completely until upstream tool authors reverse-engineer new action hashes and flight payload signatures.
- **Docker Compose v1 vs v2**: Older `docker-compose` (Python/v1) fails on `clearance/docker-compose.yml` with `Invalid interpolation format for "warp-proxy" option in service "services": "127.0.0.1:${WARP_SOCKS_PORT:-40000}:1080"`. Ensure Docker Compose v2 plugin is installed at `/usr/local/lib/docker/cli-plugins/docker-compose`.
- **Cloudflare 403 & Next.js Action ID Scraper**: Grok signup uses Next.js server actions. If Cloudflare blocks direct protocol probes (HTTP 403), use FlareSolverr on `127.0.0.1:8191` to extract signup chunks and dynamically scrape valid action IDs.
