# WhatsApp Bot Service Integration Map

Status: `ACTIVE — BOT SERVICE IMPLEMENTED`

Go backend bertindak sebagai control plane. Implementasi protokol WhatsApp, QR pairing, session persistence, reconnect, encryption, dan pengiriman pesan harus berada di WhatsApp Bot Service terpisah.

## Yang sudah tersedia di Go backend

- Tenant-aware channel list.
- Owner/admin authorization untuk connect dan disconnect.
- Subscription channel quota check.
- PostgreSQL channel metadata dan status.
- HMAC-signed HTTP client menuju Bot Service.
- Timeout request Bot Service 15 detik.

Route backend berikut sudah terdaftar dan terhubung ke Bot Service:

| Backend route | Status |
|---|---|
| `GET /api/v1/channels?workspace_id=:workspaceID` | ACTIVE; dapat membaca metadata channel di PostgreSQL |
| `POST /api/v1/channels/whatsapp/connect?workspace_id=:workspaceID` | ACTIVE; membuat QR atau pairing code |
| `DELETE /api/v1/channels/:id` | ACTIVE; menghentikan dan menghapus session |

## File backend yang menjadi integration points

| File | Peran |
|---|---|
| `cmd/server/main.go` | Wiring `channel.Repository`, `HMACBotClient`, handler, dan HTTP routes |
| `internal/channel/service.go` | Orkestrasi authorization, quota, create channel, connect, status update, dan delete |
| `internal/channel/disconnect.go` | HMAC request untuk disconnect session eksternal |
| `internal/channel/handler.go` | Public dashboard HTTP contract untuk list/connect/delete |
| `internal/channel/postgres.go` | Adapter PostgreSQL tenant-aware untuk channel |
| `db/queries/channels.sql` | Query create/list/status/delete serta authorization mutation |
| `generated/sqlc/channels.sql.go` | Generated code; jangan diedit manual |
| `internal/config/config.go` | Membaca URL Bot Service dan shared HMAC secret |
| `.env.example` | `WHATSAPP_SERVICE_URL` dan `INTERNAL_SERVICE_SECRET` |
| `docs/FRONTEND_ROUTES.md` | Status route untuk handoff frontend |
| `docs/PRD.md` | Source of truth arsitektur dan WhatsApp flows |

## Contract Bot Service

Base URL berasal dari `WHATSAPP_SERVICE_URL`, default development `http://localhost:4010`.

### Connect

```http
POST /internal/v1/channels/connect
Content-Type: application/json
X-ChatSolv-Timestamp: <RFC3339>
X-ChatSolv-Signature: <hex HMAC-SHA256>

{"channel_id":"channel-uuid","phone_number":"628123456789"}
```

Expected response:

```json
{
  "data": {
    "session_id": "bot-session-id",
    "status": "waiting_pairing",
    "qr": "qr-pairing-data",
    "pairing_code": "ABCD-EFGH"
  }
}
```

### Disconnect

```http
POST /internal/v1/channels/disconnect
Content-Type: application/json
X-ChatSolv-Timestamp: <RFC3339>
X-ChatSolv-Signature: <hex HMAC-SHA256>

{"channel_id":"channel-uuid"}
```

Expected response: any `2xx` status.

Signature input untuk kedua request:

```text
HMAC-SHA256(INTERNAL_SERVICE_SECRET, timestamp + "." + raw_request_body)
```

## Status Implementasi

WhatsApp Bot Service diimplementasikan di folder `whatsapp` menggunakan Go dan library `whatsmeow` dengan arsitektur SQLite per channel session.

Semua internal callback routes berikut aktif dan terintegrasi antara `backend` dan `whatsapp`:

| Route | Method | Arah | Fungsi |
|---|---|---|---|
| `/internal/v1/channels/connect` | `POST` | Backend -> Bot | Mulai pairing WhatsApp / generate QR code |
| `/internal/v1/channels/disconnect` | `POST` | Backend -> Bot | Disconnect WhatsApp session |
| `/internal/v1/channels/status` | `POST` | Bot -> Backend | Update status channel (`connected`, `disconnected`, dll) |
| `/internal/v1/channels/events` | `POST` | Bot -> Backend | Event pairing (`pair_success`, `qr_refresh`, dll) |
| `/internal/v1/messages/incoming` | `POST` | Bot -> Backend | Forward pesan customer WhatsApp ke AI runtime |

Semua interaksi di atas diamankan dengan HMAC SHA-256 (`INTERNAL_SERVICE_SECRET`) dan timestamp tolerance 5 menit.

## Menjalankan Layanan

1. Pastikan `INTERNAL_SERVICE_SECRET` di `Projects/backend/.env` dan `Projects/whatsapp/.env` bernilai identik (minimal 32 byte).
2. Jalankan ChatSolv Backend:
   ```bash
   cd /home/nldt/chatsolv-ecosystem/backend && make run
   ```
3. Jalankan WhatsApp Service:
   ```bash
   cd /home/nldt/chatsolv-ecosystem/whatsapp && make run
   ```
