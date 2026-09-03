# ChatSolv Backend Route Request Tests (`node:fetch`)

Direktori ini berisi script pengujian HTTP request interaktif menggunakan **Node.js native `fetch`** (`node:fetch`) untuk menguji dan menampilkan respons JSON dari seluruh endpoint backend ChatSolv.

---

## 1. Daftar File Script per Domain

| File | Endpoint yang Diuji | Deskripsi |
|---|---|---|
| `tests/health.js` | `/health`, `/ready`, `/health/live`, `/health/ready` | Liveness & readiness probe |
| `tests/auth.js` | `/api/v1/auth/register`, `/login`, `/refresh`, `/forgot-password` | Otentikasi & rotasi token |
| `tests/dashboard.js` | `/api/v1/me`, `/api/v1/dashboard` | Profil user & ringkasan operasional |
| `tests/workspaces.js` | `/api/v1/workspaces`, `/workspaces/:id`, `/subscription`, `/workspace` | CRUD workspace & subscription plan |
| `tests/agents.js` | `/api/v1/agent`, `/agent/profile`, `/agent/personality`, `/agents/:id/*`, `/agent/test` | Konfigurasi AI agent & playground |
| `tests/business.js` | `/api/v1/business`, `/settings/workspaces/:id/business`, `/policies` | Info bisnis & kebijakan operasional |
| `tests/channels.js` | `/api/v1/channels`, `/channels/whatsapp/connect`, `/channels/:id` | Integrasi WhatsApp Bot & pairing |
| `tests/knowledge.js` | `/api/v1/knowledge`, `/knowledge/text`, `/knowledge/faqs` | Knowledge base ingestion |
| `tests/conversations.js` | `/api/v1/conversations`, `/conversations/:id`, `/messages`, `/mode` | Inbox chat & human takeover mode |
| `tests/api_keys.js` | `/api/v1/api-keys` (GET, POST, DELETE) | Developer secret API keys |
| `tests/webhooks.js` | `/api/v1/webhooks` (GET, POST, PATCH, DELETE) | Developer webhooks & AES-GCM secret |
| `tests/public_sessions.js` | `/api/v1/agent-sessions`, `/messages`, `/stream` (SSE) | Public widget visitor sessions |
| `tests/internal_hmac.js` | `/internal/v1/channels/*`, `/messages/incoming`, `/agents/*` | Microservice WhatsApp HMAC routes |
| `tests/all_routes.js` | Seluruh 65 endpoint | Runner otomatis untuk menjalankan seluruh file di atas |

---

## 2. Cara Menjalankan

### Prasyarat:
1. Server ChatSolv Backend aktif (default: `http://127.0.0.1:3000`).
2. Node.js v18+ terpasang (menggunakan native `fetch`).

### Menjalankan Seluruh Route Sekaligus:
```bash
node tests/all_routes.js
```

### Menjalankan Per Domain / Module:
```bash
# 1. Test Health
node tests/health.js

# 2. Test Auth Flow
node tests/auth.js

# 3. Test Workspaces
node tests/workspaces.js

# 4. Test AI Agent Settings
node tests/agents.js

# 5. Test Business Policies
node tests/business.js

# 6. Test WhatsApp Channels
node tests/channels.js

# 7. Test Conversations & Mode Switch
node tests/conversations.js

# 8. Test API Keys
node tests/api_keys.js

# 9. Test Webhooks
node tests/webhooks.js

# 10. Test Public Website Widget Sessions
node tests/public_sessions.js

# 11. Test Internal Microservice HMAC
node tests/internal_hmac.js
```

### Kustomisasi Environment Variables (opsional):
```bash
BASE_URL="http://127.0.0.1:3000" node tests/auth.js
INTERNAL_SERVICE_SECRET="replace-with-at-least-32-random-bytes" node tests/internal_hmac.js
```
