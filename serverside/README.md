# ChatSolv Backend Platform

Backend control plane ChatSolv menggunakan GoFiber, PostgreSQL, Redis, pgx/sqlc, Argon2id, dan JWT. Implementasi saat ini mencakup autentikasi, batas tenant workspace, RBAC, langganan, dan entitlement dasar.

## Fitur

- Registrasi user dengan normalisasi email dan password hashing Argon2id
- Login dengan access token JWT dan opaque refresh token
- Forgot/reset password dengan token single-use yang memiliki TTL
- Penyimpanan session dan temporary token di Redis
- Redis-backed rate limiting untuk seluruh endpoint auth
- Health check dan dependency readiness check
- Graceful shutdown, structured logging, dan security headers
- Pembuatan workspace dengan owner membership dan langganan awal `inactive` dalam satu transaksi
- JWT-protected workspace read/update dengan role owner/admin
- Subscription dan entitlement dasar yang selalu di-resolve melalui membership tenant

Refresh token bersifat opaque, hanya hash-nya yang disimpan di Redis, dan setiap pemakaian melakukan rotation single-use.

## Daftar endpoint

Handoff route lengkap untuk frontend—termasuk pemisahan endpoint aktif, planned, public website API, dan internal service API—tersedia di [`docs/FRONTEND_ROUTES.md`](docs/FRONTEND_ROUTES.md). Panduan detail penggunaan per route lengkap dengan request/response schema dan contoh curl tersedia di [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md).

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/health` | Memeriksa apakah proses API berjalan |
| `GET` | `/ready` | Memeriksa koneksi PostgreSQL dan Redis |
| `POST` | `/api/v1/auth/register` | Mendaftarkan user baru |
| `POST` | `/api/v1/auth/login` | Login dan mendapatkan token |
| `POST` | `/api/v1/auth/refresh` | Merotasi refresh token dan menerbitkan access token baru |
| `POST` | `/api/v1/auth/forgot-password` | Membuat password reset token |
| `POST` | `/api/v1/auth/reset-password` | Mengganti password menggunakan reset token |
| `GET` | `/api/v1/me` | Membaca identity dan daftar workspace membership user |
| `GET` | `/api/v1/dashboard?workspace_id=:workspaceID` | Membaca aggregate dashboard tenant |
| `POST` | `/api/v1/workspaces` | Membuat workspace, owner membership, dan langganan awal |
| `GET` | `/api/v1/workspaces/:workspaceID` | Membaca workspace jika user adalah member |
| `PATCH` | `/api/v1/workspaces/:workspaceID` | Mengubah workspace sebagai owner/admin |
| `GET` | `/api/v1/workspaces/:workspaceID/subscription` | Membaca status langganan dan entitlement workspace |
| `GET` | `/api/v1/workspace?workspace_id=:workspaceID` | Canonical route untuk membaca workspace |
| `PATCH` | `/api/v1/workspace?workspace_id=:workspaceID` | Canonical route untuk mengubah workspace |
| `GET` | `/api/v1/agents/:agentID/personality` | Membaca personality agent milik workspace user |
| `PATCH` | `/api/v1/agents/:agentID/personality` | Mengubah personality dan mengantrekan sinkronisasi agent |
| `GET` | `/api/v1/agents/:agentID/profile` | Membaca profil bot/agent |
| `PATCH` | `/api/v1/agents/:agentID/profile` | Mengubah profil bot dan mengantrekan sinkronisasi |
| `GET` | `/api/v1/agent?workspace_id=:workspaceID` | Canonical route untuk membaca default agent workspace |
| `PATCH` | `/api/v1/agent?workspace_id=:workspaceID` | Canonical route untuk mengubah nama agent |
| `GET` | `/api/v1/agent/profile?workspace_id=:workspaceID` | Canonical route untuk profil agent |
| `PATCH` | `/api/v1/agent/profile?workspace_id=:workspaceID` | Canonical route untuk mengubah profil agent |
| `GET` | `/api/v1/agent/personality?workspace_id=:workspaceID` | Canonical route untuk personality agent |
| `PATCH` | `/api/v1/agent/personality?workspace_id=:workspaceID` | Canonical route untuk mengubah personality agent |
| `POST` | `/api/v1/agent/test?workspace_id=:workspaceID` | Menjalankan isolated agent test playground |
| `GET` | `/api/v1/settings/workspaces/:workspaceID/business` | Membaca profil bisnis workspace |
| `PATCH` | `/api/v1/settings/workspaces/:workspaceID/business` | Mengubah profil bisnis workspace |
| `GET` | `/api/v1/settings/workspaces/:workspaceID/policies` | Membaca kebijakan bisnis workspace |
| `PATCH` | `/api/v1/settings/workspaces/:workspaceID/policies` | Mengubah kebijakan bisnis workspace |
| `GET` | `/api/v1/business?workspace_id=:workspaceID` | Canonical route untuk membaca business profile |
| `PATCH` | `/api/v1/business?workspace_id=:workspaceID` | Canonical route untuk mengubah business profile |
| `GET` | `/api/v1/channels?workspace_id=:workspaceID` | Membaca channel milik workspace |
| `POST` | `/api/v1/channels/whatsapp/connect?workspace_id=:workspaceID` | Memulai pairing WhatsApp — blocked sampai Bot Service dibuat |
| `DELETE` | `/api/v1/channels/:id` | Disconnect dan menghapus channel — blocked sampai Bot Service dibuat |
| `GET` | `/api/v1/api-keys?workspace_id=:workspaceID` | List API key metadata untuk owner/admin |
| `POST` | `/api/v1/api-keys?workspace_id=:workspaceID` | Membuat API key; raw secret hanya ditampilkan sekali |
| `DELETE` | `/api/v1/api-keys/:id` | Mencabut API key sebagai owner/admin |
| `GET` | `/api/v1/webhooks?workspace_id=:workspaceID` | List webhook endpoint untuk owner/admin |
| `POST` | `/api/v1/webhooks?workspace_id=:workspaceID` | Membuat webhook endpoint; secret hanya sekali |
| `PATCH` | `/api/v1/webhooks/:id` | Mengubah URL, events, atau status webhook |
| `DELETE` | `/api/v1/webhooks/:id` | Menghapus webhook endpoint |
| `POST` | `/api/v1/agent-sessions` | Membuat website session memakai secret API key |
| `POST` | `/api/v1/agent-sessions/:id/messages` | Pesan memakai ephemeral client token |
| `POST` | `/api/v1/agent-sessions/:id/messages/stream` | SSE memakai ephemeral client token |
| `GET` | `/api/v1/knowledge?workspace_id=:workspaceID` | Melihat daftar knowledge tenant |
| `GET` | `/api/v1/knowledge/:id` | Membaca metadata satu knowledge source |
| `PATCH` | `/api/v1/knowledge/:id` | Mengubah judul dan mengantrekan re-ingestion |
| `DELETE` | `/api/v1/knowledge/:id` | Mengantrekan penghapusan knowledge |
| `POST` | `/api/v1/knowledge/:id/retry` | Mengulang ingestion knowledge yang gagal |
| `POST` | `/api/v1/knowledge/documents` | Upload dokumen dan mengantrekan ingestion |
| `POST` | `/api/v1/knowledge/text` | Menambahkan knowledge teks |
| `POST` | `/api/v1/knowledge/faqs` | Menambahkan knowledge FAQ |
| `GET` | `/api/v1/conversations?workspace_id=:workspaceID` | List percakapan tenant dengan filter status/mode |
| `GET` | `/api/v1/conversations/:id` | Detail percakapan satu tenant |
| `GET` | `/api/v1/conversations/:id/messages` | Pesan percakapan dengan cursor pagination |
| `PATCH` | `/api/v1/conversations/:id/mode` | Handoff/resume percakapan (owner/admin) |
| `GET` | `/health/live` | Liveness probe alias |
| `GET` | `/health/ready` | Readiness probe alias dengan bounded dependency ping |
| `POST` | `/internal/v1/channels/events` | Internal WA event — BLOCKED BY BOT SERVICE |
| `POST` | `/internal/v1/channels/status` | Internal WA status callback — BLOCKED BY BOT SERVICE |
| `POST` | `/internal/v1/messages/incoming` | Incoming WA message — BLOCKED BY BOT SERVICE |
| `POST` | `/internal/v1/agents/:agentID/respond` | Internal agent respond (HMAC) |
| `GET` | `/internal/v1/agents/:agentID/health` | Internal agent health check (HMAC) |

Base URL development:

```text
http://localhost:3000
```

## Instalasi dari awal

### 1. Prasyarat

Install software berikut:

- Git
- Go 1.26 atau lebih baru
- GNU Make
- Docker Engine dengan Docker Compose
- OpenSSL
- Python 3
- `golang-migrate`

Pastikan tool utama tersedia:

```bash
go version
docker --version
docker compose version
make --version
openssl version
python3 --version
```

Jika Docker menghasilkan error permission untuk `/var/run/docker.sock`, tambahkan user saat ini ke group Docker:

```bash
sudo usermod -aG docker "$USER"
```

Setelah itu logout lalu login kembali. Alternatif sementara adalah menjalankan perintah Docker dengan `sudo`, tetapi konfigurasi group lebih nyaman untuk development.

### 2. Ambil source code

Clone repository ini, lalu masuk ke direktori project:

```bash
git clone <repository-url> backend
cd backend
```

Jika source code sudah tersedia, cukup masuk ke root project yang berisi `go.mod`, `Makefile`, dan `docker-compose.yml`.

### 3. Install Go dependencies

```bash
go mod download
```

### 4. Install migration CLI

```bash
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Pastikan direktori binary Go ada di `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
migrate -version
```

Agar permanen, tambahkan baris `export PATH=...` tersebut ke `~/.bashrc`, `~/.zshrc`, atau file konfigurasi shell yang digunakan.

### 5. Buat konfigurasi development

Jalankan:

```bash
make jwt
```

Perintah ini akan:

1. Menyalin `.env.example` menjadi `.env` jika `.env` belum ada.
2. Membuat `JWT_SECRET` acak menggunakan OpenSSL.
3. Menulis secret tersebut ke `.env`.

> Menjalankan `make jwt` lagi akan merotasi JWT secret dan membuat access token lama tidak valid.

Konfigurasi default memakai:

- PostgreSQL: `localhost:5432`
- Database: `chatsolv`
- User/password development: `postgres` / `postgres`
- Redis: `localhost:6379`
- API: `localhost:3000`

`Makefile` otomatis membaca dan mengekspor isi `.env` untuk target seperti `make run` dan `make migrate-up`.

### 6. Jalankan PostgreSQL dan Redis

```bash
make docker-up
```

Periksa status container:

```bash
docker compose ps
```

Tunggu sampai PostgreSQL dan Redis berstatus healthy.

### 7. Jalankan database migration

```bash
make migrate-up
```

Rollback satu migration terakhir jika diperlukan:

```bash
make migrate-down
```

### 8. Jalankan API

```bash
make run
```

Server berjalan di:

```text
http://localhost:3000
```

Biarkan terminal ini tetap berjalan. Gunakan terminal lain untuk mencoba API.

### 9. Verifikasi service

```bash
curl http://localhost:3000/health
curl http://localhost:3000/ready
```

Expected response:

```json
{"status":"ok"}
```

```json
{"status":"ready"}
```

Jika `/health` berhasil tetapi `/ready` mengembalikan `503`, proses API berjalan namun PostgreSQL atau Redis belum siap.

## Alternatif: menjalankan seluruh stack dengan Docker

Setelah `.env` dibuat dengan `make jwt`, jalankan:

```bash
docker compose up --build
```

Container backend menggunakan hostname internal `postgres` dan `redis`. Migration tetap perlu dijalankan dari host:

```bash
make migrate-up
```

Hentikan seluruh stack dengan:

```bash
docker compose down
```

Tambahkan `-v` hanya jika ingin sekaligus menghapus volume PostgreSQL beserta seluruh data development:

```bash
docker compose down -v
```

## Format response

Endpoint API menggunakan envelope PRD ChatSolv yang konsisten dan menyertakan `request_id` jika middleware request ID aktif.

Success:

```json
{"data": {}, "meta": {"message": "Request successful"}, "request_id": "..."}
```

Error:

```json
{"error": {"code": "INTERNAL_ERROR", "message": "Something went wrong"}, "request_id": "..."}
```

Semua request `POST` wajib menggunakan header:

```http
Content-Type: application/json
```

## How to Use / Menggunakan API

Contoh berikut menggunakan variable agar mudah disalin:

```bash
BASE_URL=http://localhost:3000
```

### Authentication untuk protected router

Router `/health`, `/ready`, dan `/api/v1/auth/*` tidak membutuhkan access token. Router workspace, agent, settings, dan knowledge membutuhkan JWT dari response login:

```bash
ACCESS_TOKEN='<access_token-dari-response-login>'

curl "$BASE_URL/api/v1/workspaces/<workspace-id>" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Untuk request JSON, selalu kirim kedua header berikut:

```bash
-H "Authorization: Bearer $ACCESS_TOKEN" \
-H "Content-Type: application/json"
```

### Router yang aktif

#### System dan authentication

```text
GET  /health
GET  /ready

POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
```

#### Workspace dan subscription

Semua router berikut membutuhkan `Authorization: Bearer <access_token>`:

```text
POST  /api/v1/workspaces
GET   /api/v1/workspaces/:workspaceID
PATCH /api/v1/workspaces/:workspaceID
GET   /api/v1/workspaces/:workspaceID/subscription
```

Create workspace:

```bash
curl -X POST "$BASE_URL/api/v1/workspaces" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Toko Contoh",
    "slug": "toko-contoh",
    "timezone": "Asia/Jakarta"
  }'
```

Response sukses menggunakan status `202 Accepted` karena Hermes Agent dan Obsidian Vault diprovisikan secara asynchronous.

Update workspace, hanya untuk role `owner` atau `admin`:

```bash
curl -X PATCH "$BASE_URL/api/v1/workspaces/$WORKSPACE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Bisnis Baru","timezone":"Asia/Jakarta"}'
```

#### Agent personality dan profile

```text
GET   /api/v1/agents/:agentID/personality
PATCH /api/v1/agents/:agentID/personality
GET   /api/v1/agents/:agentID/profile
PATCH /api/v1/agents/:agentID/profile
```

Update personality, hanya untuk role `owner` atau `admin`:

```bash
curl -X PATCH "$BASE_URL/api/v1/agents/$AGENT_ID/personality" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bot_name": "Naya",
    "role": "Customer Service",
    "tone": "friendly",
    "communication_style": "casual_professional",
    "primary_language": "id",
    "response_length": "medium",
    "emoji_usage": "moderate",
    "greeting_style": "warm",
    "closing_style": "helpful",
    "custom_instructions": "Jawab dengan jelas dan sopan.",
    "behavior_rules": [],
    "escalation_rules": [],
    "forbidden_topics": [],
    "fallback_behavior": "Tawarkan bantuan admin."
  }'
```

Update agent profile:

```bash
curl -X PATCH "$BASE_URL/api/v1/agents/$AGENT_ID/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Naya",
    "description": "Customer service Toko Contoh",
    "greeting_message": "Halo, ada yang bisa Naya bantu?",
    "away_message": "Kami sedang di luar jam operasional.",
    "fallback_message": "Naya akan menghubungkan kamu ke admin.",
    "language": "id"
  }'
```

Mutation agent configuration mengembalikan `config_version` dan status `syncing`. Worker harus berjalan agar perubahan ditulis ke vault dan disinkronkan ke Hermes:

```bash
make worker
```

#### Business profile dan policies

```text
GET   /api/v1/settings/workspaces/:workspaceID/business
PATCH /api/v1/settings/workspaces/:workspaceID/business
GET   /api/v1/settings/workspaces/:workspaceID/policies
PATCH /api/v1/settings/workspaces/:workspaceID/policies
```

Update business profile:

```bash
curl -X PATCH "$BASE_URL/api/v1/settings/workspaces/$WORKSPACE_ID/business" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Toko Contoh",
    "industry": "retail",
    "business_description": "Toko kebutuhan harian.",
    "website": "https://example.com",
    "address": "Jakarta",
    "business_hours": {"monday":"09:00-17:00"},
    "timezone": "Asia/Jakarta",
    "brand_voice": "ramah dan profesional",
    "company_values": ["jujur", "cepat"]
  }'
```

Update policies:

```bash
curl -X PATCH "$BASE_URL/api/v1/settings/workspaces/$WORKSPACE_ID/policies" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipping_policy": "Pesanan dikirim maksimal dua hari kerja.",
    "refund_policy": "Refund dapat diajukan dalam tujuh hari.",
    "return_policy": "Barang harus belum digunakan.",
    "warranty_policy": "Garansi berlaku 30 hari.",
    "payment_policy": "Pembayaran melalui metode yang tersedia.",
    "complaint_policy": "Keluhan diproses maksimal 1x24 jam."
  }'
```

#### Knowledge

Semua router knowledge membutuhkan access token dan memiliki rate limit terpisah, saat ini 20 request per menit:

```text
GET  /api/v1/knowledge?workspace_id=:workspaceID&limit=50
GET  /api/v1/knowledge/:id
PATCH /api/v1/knowledge/:id
DELETE /api/v1/knowledge/:id
POST /api/v1/knowledge/:id/retry
POST /api/v1/knowledge/documents
POST /api/v1/knowledge/text
POST /api/v1/knowledge/faqs
```

Upload document menggunakan `multipart/form-data`, bukan JSON:

```bash
curl -X POST "$BASE_URL/api/v1/knowledge/documents" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "workspace_id=$WORKSPACE_ID" \
  -F "title=Kebijakan Refund" \
  -F "file=@./refund-policy.pdf"
```

Format yang saat ini diterima: PDF, DOCX, TXT, CSV, dan Markdown. Response sukses adalah `202 Accepted`; jalankan worker untuk memproses ingestion.

Create text knowledge:

```bash
curl -X POST "$BASE_URL/api/v1/knowledge/text" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"workspace_id\":\"$WORKSPACE_ID\",\"title\":\"Jam Operasional\",\"content\":\"Buka Senin-Sabtu pukul 09:00-17:00.\"}"
```

Create FAQ:

```bash
curl -X POST "$BASE_URL/api/v1/knowledge/faqs" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"workspace_id\":\"$WORKSPACE_ID\",\"title\":\"FAQ Refund\",\"question\":\"Apakah barang bisa direfund?\",\"answer\":\"Bisa dalam tujuh hari sesuai ketentuan.\",\"category\":\"refund\"}"
```

> Router conversation dashboard, webhook delivery worker, dan internal WhatsApp message belum didaftarkan. Website session API, API key management, dan webhook endpoint CRUD sudah aktif. Router channel sudah terdaftar, tetapi connect/disconnect belum operasional end-to-end sampai WhatsApp Bot Service dibuat; lihat [`docs/WHATSAPP_INTEGRATION.md`](docs/WHATSAPP_INTEGRATION.md).

### Health check

```bash
curl "$BASE_URL/health"
```

Response `200 OK`:

```json
{
  "status": "ok"
}
```

### Readiness check

```bash
curl "$BASE_URL/ready"
```

Response `200 OK` ketika PostgreSQL dan Redis siap:

```json
{
  "status": "ready"
}
```

Response `503 Service Unavailable` jika dependency belum siap:

```json
{
  "error": {"code": "NOT_READY", "message": "Service unavailable"},
  "request_id": "..."
}
```

### Register

Endpoint:

```http
POST /api/v1/auth/register
```

Body:

| Field | Tipe | Aturan |
|---|---|---|
| `name` | string | Wajib, 2-100 karakter |
| `email` | string | Wajib, format email valid, maksimal 254 karakter |
| `password` | string | Wajib, 8-128 karakter |

Request:

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "password": "rahasia123"
  }'
```

Response `201 Created`:

```json
{
  "data": {
    "id": "79d4a9e7-9fba-4b66-a7ed-8db99952f9c8",
    "name": "Budi Santoso",
    "email": "budi@example.com"
  },
  "meta": {"message": "Registration successful"},
  "request_id": "..."
}
```

Jika email sudah terdaftar, response `409 Conflict`:

```json
{
  "error": {"code": "USER_ALREADY_EXISTS", "message": "Email already registered"},
  "request_id": "..."
}
```

### Login

Endpoint:

```http
POST /api/v1/auth/login
```

Body:

| Field | Tipe | Aturan |
|---|---|---|
| `email` | string | Wajib, format email valid |
| `password` | string | Wajib, maksimal 128 karakter |

Request:

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "budi@example.com",
    "password": "rahasia123"
  }'
```

Response `200 OK`:

```json
{
  "data": {
    "access_token": "<jwt-access-token>",
    "refresh_token": "<opaque-refresh-token>",
    "token_type": "Bearer",
    "expires_in": 900
  },
  "meta": {"message": "Login successful"},
  "request_id": "..."
}
```

`expires_in` dinyatakan dalam detik. Dengan konfigurasi default `JWT_ACCESS_TTL=15m`, nilainya adalah `900`.

Simpan token dari response jika ingin digunakan oleh client:

```bash
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}')

printf '%s\n' "$LOGIN_RESPONSE"
```

Jika credential salah atau user tidak ditemukan, response tetap disamakan menjadi `401 Unauthorized`:

```json
{
  "error": {"code": "INVALID_CREDENTIALS", "message": "Invalid credentials"},
  "request_id": "..."
}
```

### Forgot password

Endpoint:

```http
POST /api/v1/auth/forgot-password
```

Body:

| Field | Tipe | Aturan |
|---|---|---|
| `email` | string | Wajib, format email valid |

Request:

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "budi@example.com"
  }'
```

Response `200 OK` selalu dibuat generik, baik email terdaftar maupun tidak:

```json
{
  "data": {},
  "meta": {"message": "If the account exists, password reset instructions have been sent"},
  "request_id": "..."
}
```

Dalam mode development, belum ada email provider nyata. Reset token dicetak pada log server dengan message:

```text
development password reset delivery
```

Cari field `reset_token` pada baris log tersebut. Jangan menggunakan development sender atau menampilkan reset token di log production.

### Reset password

Endpoint:

```http
POST /api/v1/auth/reset-password
```

Body:

| Field | Tipe | Aturan |
|---|---|---|
| `token` | string | Wajib, ambil dari log development |
| `new_password` | string | Wajib, 8-128 karakter |

Request:

```bash
RESET_TOKEN='<token-dari-log-development>'

curl -i -X POST "$BASE_URL/api/v1/auth/reset-password" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$RESET_TOKEN\",\"new_password\":\"passwordBaru123\"}"
```

Response `200 OK`:

```json
{
  "data": {},
  "meta": {"message": "Password reset successful"},
  "request_id": "..."
}
```

Reset token bersifat single-use dan memiliki TTL. Password reset juga mencabut refresh session user yang sudah ada.

Token invalid, expired, atau sudah pernah digunakan menghasilkan response `400 Bad Request` yang sama:

```json
{
  "error": {"code": "INVALID_RESET_TOKEN", "message": "Invalid or expired reset token"},
  "request_id": "..."
}
```

Setelah reset berhasil, login menggunakan password baru:

```bash
curl -i -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "budi@example.com",
    "password": "passwordBaru123"
  }'
```

## Validation dan error umum

### Content-Type salah atau tidak ada

Status `400`:

```json
{
  "error": {"code": "INVALID_CONTENT_TYPE", "message": "Content-Type must be application/json"},
  "request_id": "..."
}
```

### JSON malformed

Status `400`:

```json
{
  "error": {"code": "INVALID_JSON", "message": "Invalid JSON body"},
  "request_id": "..."
}
```

### Field tidak valid

Status `400`:

```json
{
  "error": {"code": "VALIDATION_ERROR", "message": "Validation failed"},
  "request_id": "..."
}
```

### Rate limit

Secara default, setiap endpoint auth menerima maksimal 10 request per IP dalam window 1 menit. Jika terlampaui, status `429`:

```json
{
  "error": {"code": "RATE_LIMITED", "message": "Too many requests"},
  "request_id": "..."
}
```

## HTTP status dan error code

| Status | Error code | Kondisi |
|---|---|---|
| `200` | - | Login, forgot password, atau reset password berhasil |
| `201` | - | Registrasi berhasil |
| `400` | `INVALID_CONTENT_TYPE` | Request auth bukan JSON |
| `400` | `INVALID_JSON` | JSON tidak dapat diparse |
| `400` | `VALIDATION_ERROR` | Field request tidak memenuhi aturan |
| `400` | `INVALID_RESET_TOKEN` | Reset token invalid, expired, atau reused |
| `401` | `INVALID_CREDENTIALS` | Email/password login salah |
| `409` | `USER_ALREADY_EXISTS` | Email sudah terdaftar |
| `429` | `RATE_LIMITED` | Batas request terlampaui |
| `503` | `NOT_READY` | PostgreSQL atau Redis belum siap |
| `500` | `INTERNAL_ERROR` | Error internal yang tidak diekspos ke client |

Response memiliki header `X-Request-ID`. Client boleh mengirim UUID melalui header yang sama; jika nilainya tidak valid atau tidak dikirim, server membuat UUID baru.

## Konfigurasi environment

| Variable | Default/example | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | Saat ini hanya `development` dan `test` yang diizinkan karena production email sender belum tersedia |
| `APP_PORT` | `3000` | Port HTTP API |
| `DATABASE_URL` | `postgres://postgres:***@localhost:5432/chatsolv?sslmode=disable` | PostgreSQL connection URL |
| `DATABASE_MAX_CONNS` | `20` | Maksimum koneksi pool |
| `DATABASE_MIN_CONNS` | `5` | Minimum koneksi pool |
| `DATABASE_MAX_CONN_LIFETIME` | `1h` | Maksimum lifetime koneksi |
| `DATABASE_MAX_CONN_IDLE_TIME` | `30m` | Maksimum idle time koneksi |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `JWT_SECRET` | generated by `make jwt` | Secret HS256, minimal 32 byte |
| `JWT_ACCESS_TTL` | `15m` | Masa berlaku access token |
| `JWT_REFRESH_TTL` | `720h` | Masa berlaku refresh session |
| `PASSWORD_RESET_TTL` | `15m` | Masa berlaku reset token |
| `CORS_ORIGINS` | `http://localhost:3000` | Origin yang diizinkan, pisahkan banyak origin dengan koma |
| `REQUEST_BODY_LIMIT` | `16384` | Maksimum request body dalam byte |
| `RATE_LIMIT_MAX` | `10` | Maksimum request per endpoint/IP/window |
| `RATE_LIMIT_WINDOW` | `1m` | Durasi rate-limit window |
| `SHUTDOWN_TIMEOUT` | `10s` | Batas waktu graceful shutdown |
| `LOG_LEVEL` | `debug` | Gunakan `debug` atau level info default |

Semua duration memakai format Go, misalnya `30s`, `15m`, atau `1h`.

## Development commands

```bash
make run          # menjalankan API
make build        # build seluruh package
make test         # menjalankan test
make lint         # menjalankan golangci-lint
make fmt          # format file Go
make sqlc         # generate kode sqlc
make jwt          # membuat/merotasi JWT secret development
make migrate-up   # apply seluruh pending migration
make migrate-down # rollback satu migration
make docker-up    # menjalankan PostgreSQL dan Redis
make docker-down  # menghentikan Docker Compose stack
```

## Benchmark dengan k6

Benchmark register dipisahkan agar hasil `201` tidak tercampur dengan response `429`:

```bash
# Performance: jalankan backend dengan limit tinggi
make RATE_LIMIT_MAX=10000 run
k6 run bench/register.js

# Rate limit: gunakan limit default 10 pada backend
make run
k6 run bench/register-rate-limit.js
```

Jalankan perintah backend dan k6 di terminal terpisah. Petunjuk, konfigurasi, dan cara membaca hasil selengkapnya tersedia di [`bench/README.md`](bench/README.md).

## Testing dan quality checks

Jalankan pemeriksaan utama sebelum mengirim perubahan:

```bash
gofmt -w cmd internal generated pkg
go test ./... -count=1
go test -race ./internal/auth ./internal/config ./internal/workspace -count=1
go build ./...
go vet ./...
git diff --check
```

Jika tersedia:

```bash
golangci-lint run
govulncheck ./...
```

PostgreSQL integration test menggunakan `TEST_DATABASE_URL`. Test tersebut dapat skip jika database test atau container runtime tidak tersedia.

Contoh:

```bash
TEST_DATABASE_URL='postgres://postgres:***@localhost:5432/chatsolv_test?sslmode=disable' \
  go test ./internal/auth -count=1
```

Gunakan database khusus test jika data development perlu dipertahankan.

## Troubleshooting

### `failed to parse scheme from database URL: URL cannot be empty`

Pastikan `.env` tersedia dan `DATABASE_URL` terisi:

```bash
make jwt
```

Kemudian jalankan kembali:

```bash
make migrate-up
```

### `migrate: command not found`

Install CLI dan tambahkan Go binary directory ke `PATH`:

```bash
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Docker socket permission denied

```bash
sudo usermod -aG docker "$USER"
```

Logout/login, lalu cek:

```bash
docker ps
```

### Port `5432`, `6379`, atau `3000` sudah dipakai

Cari proses/container yang menggunakan port:

```bash
docker compose ps
sudo ss -ltnp | grep -E ':(3000|5432|6379)\b'
```

Hentikan service yang bentrok atau sesuaikan port development secara konsisten.

### API gagal start karena konfigurasi

Pastikan:

- `DATABASE_URL` tidak kosong.
- `JWT_SECRET` minimal 32 byte; gunakan `make jwt`.
- PostgreSQL dan Redis sedang berjalan.
- Semua TTL/duration bernilai positif.
- `DATABASE_MIN_CONNS` tidak lebih besar dari `DATABASE_MAX_CONNS`.
- `APP_ENV` tetap `development` atau `test` sampai production email sender tersedia.

## Catatan keamanan

- Jangan commit file `.env`.
- Jangan menggunakan credential development untuk production.
- Jangan log password, password hash, access token, refresh token, JWT secret, atau credential database.
- Development sender mencetak reset token hanya untuk local development.
- Gunakan secret kuat, TLS, CORS terbatas, PostgreSQL/Redis terproteksi, dan email provider nyata sebelum production deployment.
- PostgreSQL adalah source of truth untuk user; Redis hanya menyimpan state temporary yang memiliki TTL.

