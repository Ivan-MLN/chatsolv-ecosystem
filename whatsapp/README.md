# ChatSolv WhatsApp Gateway Service 📱
*Standalone WhatsApp Web Socket Gateway & HMAC Webhook Dispatcher*

---

## 🌟 1. Pengenalan WhatsApp Service

**ChatSolv WhatsApp Service** adalah microservice mandiri yang dibangun menggunakan bahasa **Go (Golang)** dan engine [whatsmeow](https://go.mau.fi/whatsmeow). Service ini bertindak sebagai jembatan komunikasi langsung antara backend ChatSolv dan protokol soket WhatsApp Web resmi.

Service ini sengaja dipisahkan dari backend control plane (`serverside`) dengan tujuan:
1. **Isolasi Kegagalan**: Memastikan beban tinggi atau diskoneksi soket WhatsApp tidak mempengaruhi API utama.
2. **Efisiensi Memori**: Menjalankan engine soket Web yang ringan dengan database SQLite terisolasi per channel.
3. **Keamanan Maksimal**: Menggunakan autentikasi berbasis tanda tangan kriptografis **HMAC-SHA256** untuk seluruh interaksi internal.

---

## 🏗️ 2. Alur Kerja Arsitektur & Callback

```text
[ ChatSolv Backend ]
       │
       │ 1. POST /internal/v1/channels/connect (HMAC Signed)
       ▼
[ WhatsApp Gateway Service ] ◄──► [ WhatsApp Web Protocol (whatsmeow) ]
       │                                          ▲
       │                                          │ (Customer Kirim WA)
       │ 2. POST /internal/v1/messages/incoming   ▼
       │    (HMAC Signed Webhook Event)      [ Customer WA ]
       ▼
[ ChatSolv Backend ]
```

### Tugas & Tanggung Jawab Utama:
- Mengelola lifecycle session WhatsApp Web per channel (Pairing QR Code, Auto-Reconnect, Disconnect).
- Menyimpan sesi login (device keys, auth tokens) ke file database SQLite mandiri per channel.
- Mengunduh attachment media (gambar, audio, dokumen) dan meneruskannya ke backend secara aman.
- Memfilter pesan hanya untuk obrolan 1-on-1 (Direct Message) serta mengabaikan pesan grup (`@g.us`), channel broadcast, dan newsletter.

---

## 📂 3. Penjelasan Detail Setiap File & Fungsinya

```text
whatsapp/
├── cmd/
│   └── server/
│       └── main.go                 # Entrypoint server, sinkronisasi versi WhatsApp Web, inisialisasi whatsmeow manager, HTTP routing, dan graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Validasi & parsing environment variables (PORT, INTERNAL_SERVICE_SECRET, BACKEND_URL, DB_ROOT)
│   ├── server/
│   │   ├── handler.go              # HTTP Handler untuk endpoint /internal/v1/channels/connect, /disconnect, /status, dan /health
│   │   ├── middleware.go           # Middleware validasi HMAC-SHA256 (memeriksa timestamp 5-menit dan signature body)
│   │   ├── server.go               # Inisialisasi HTTP server net/http standar Go
│   │   └── server_test.go          # Unit test untuk HTTP endpoints dan middleware security
│   ├── whatsapp/
│   │   ├── manager.go              # Koordinator multi-channel sesi, pembuatan QR code pairing, reconnect loop
│   │   ├── session.go              # Logika state koneksi individual channel (Device JID, EventHandler whatsmeow)
│   │   ├── session_test.go         # Unit test untuk simulasi session lifecycle
│   │   └── store.go                # Konstruktor SQLite container berbasis driver modernc (tanpa cgo)
│   └── callback/
│       ├── client.go               # HTTP Client yang mengirim webhook event HMAC-signed ke backend
│       ├── client_test.go          # Unit test untuk verifikasi signature webhook
│       ├── event.go                # Parser event pesan masuk WhatsApp, ekstrak teks, pengirim, dan media
│       └── types.go                # Definisi struct DTO webhook payload
├── data/
│   └── sessions/                   # Direktori penyimpanan file SQLite database (.db) per channel connection ID
├── .env.example                    # Template environment variables
├── Makefile                        # Skrip automasi build, run, test, dan code formatting
├── go.mod & go.sum                 # Definisi dependensi Go
└── README.md                       # Dokumentasi resmi WhatsApp service
```

---

## 🔒 4. Spesifikasi Keamanan Internal API (HMAC-SHA256)

Seluruh route di bawah `/internal/v1/*` diwajibkan menyertakan header otentikasi HMAC:

### Headers Wajib:
- `X-ChatSolv-Timestamp`: Format RFC3339 (misal: `2026-09-03T14:30:00Z`). Replay window dibatasi maksimum **5 menit**.
- `X-ChatSolv-Signature`: Hash heksadesimal dari:
  ```text
  hex(HMAC-SHA256(INTERNAL_SERVICE_SECRET, timestamp + "." + request_body))
  ```

### Daftar Endpoint:
| Method | Path | Deskripsi | Auth |
| :--- | :--- | :--- | :--- |
| `POST` | `/internal/v1/channels/connect` | Memulai sesi WhatsApp & mengembalikan QR code string jika belum login | HMAC |
| `POST` | `/internal/v1/channels/disconnect` | Memutus koneksi soket & logout sesi WhatsApp | HMAC |
| `GET` | `/internal/v1/channels/status?channel_id=xxx` | Memeriksa status koneksi channel saat ini (`connected`, `waiting_pairing`, `disconnected`) | HMAC |
| `GET` | `/health` | Liveness probe endpoint | No Auth |

---

## 🛠️ 5. Panduan Developer & Menjalankan Service (DX)

```bash
# 1. Konfigurasi Environment
cp .env.example .env
# Pastikan INTERNAL_SERVICE_SECRET minimal 32 karakter dan sama persis dengan backend

# 2. Menjalankan Service secara Lokal
make run

# 3. Menjalankan Unit Tests
make test

# 4. Melakukan Format Kode Go
make fmt

# 5. Build Binary Produksi
make build
```
