# ChatSolv WhatsApp Bot Service

Layanan Go terpisah yang menangani koneksi WhatsApp menggunakan [whatsmeow](https://go.mau.fi/whatsmeow).

Backend ChatSolv **tidak** menjalankan protokol WhatsApp secara langsung. Service ini yang bertanggung jawab atas:

- QR pairing dan session management per channel
- Reconnect otomatis
- Enkripsi end-to-end (ditangani whatsmeow)
- Meneruskan pesan masuk ke backend ChatSolv
- Menerima perintah connect/disconnect dari backend

## Arsitektur

```
ChatSolv Backend
    ↓ POST /internal/v1/channels/connect   (HMAC signed)
WhatsApp Bot Service  <──>  WhatsApp Servers
    ↓ POST /internal/v1/messages/incoming  (HMAC signed)
ChatSolv Backend
```

## Setup

```bash
cp .env.example .env
# Edit INTERNAL_SERVICE_SECRET (harus sama dengan backend)
make run
```

## Internal API

Semua route kecuali `/health` diproteksi dengan HMAC-SHA256.

Headers yang diperlukan:
- `X-ChatSolv-Timestamp`: RFC3339 timestamp
- `X-ChatSolv-Signature`: hex(HMAC-SHA256(secret, timestamp + "." + body))
- Replay window: 5 menit

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/internal/v1/channels/connect` | Mulai session WA, return QR |
| POST | `/internal/v1/channels/disconnect` | Putus session WA |
| GET  | `/internal/v1/channels/status?channel_id=xxx` | Cek status session |
| GET  | `/health` | Health check (no auth) |

### Connect Request

```json
{"channel_id": "uuid-dari-backend"}
```

### Connect Response

```json
{
  "success": true,
  "data": {
    "session_id": "uuid-dari-backend",
    "status": "waiting_pairing",
    "qr": "qr-string-untuk-di-scan"
  }
}
```

Jika device sudah pernah paired sebelumnya, `status` akan `"connected"` dan `qr` kosong.

## Callback ke Backend

Service ini memanggil backend ChatSolv secara otomatis ketika ada event:

| Event | Backend Route |
|-------|--------------|
| Pesan masuk | `POST /internal/v1/messages/incoming` |
| Status berubah (connect/disconnect/logout) | `POST /internal/v1/channels/status` |
| QR refresh / pair success | `POST /internal/v1/channels/events` |

## Session Storage

Setiap channel memiliki SQLite database terpisah di `DB_ROOT/<channel_id>.db`.
whatsmeow mengelola skema dan migrasinya secara otomatis.

## Environment Variables

| Variable | Default | Keterangan |
|----------|---------|------------|
| `PORT` | `4010` | Port HTTP |
| `INTERNAL_SERVICE_SECRET` | — | **Required.** Min 32 bytes. Harus sama dengan backend |
| `BACKEND_URL` | `http://localhost:3000` | URL backend ChatSolv |
| `DB_ROOT` | `./data/sessions` | Direktori SQLite databases |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `CALLBACK_TIMEOUT` | `10s` | Timeout callback ke backend |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
