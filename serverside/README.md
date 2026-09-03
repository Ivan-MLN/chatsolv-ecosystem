# ChatSolv Serverside Control Plane ⚙️
*High-Performance Enterprise Backend Architecture for Customer Service Automation*

---

## 🌟 1. Pengenalan Serverside

**Serverside** adalah control plane dan backend engine inti untuk ekosistem ChatSolv. Dibangun menggunakan bahasa **Go (Golang)** dan framework **GoFiber v2**, service ini bertanggung jawab atas seluruh logika bisnis kritis, mencakup:
- Autentikasi multi-tenant level enterprise (Argon2id, JWT, Redis-backed single-use refresh token rotation).
- Isolasi workspace, sistem keanggotaan (RBAC), dan langganan (entitlements).
- Ingestion data knowledge base (PDF, TXT, FAQ, Markdown) ke format RAG vaults.
- Orchestration AI agent Hermes runtime dan penegakan batas keamanan prompt (`SOUL.md`).
- Pengelolaan percakapan multi-channel, pesan masuk WhatsApp, dan smart human-agent handoff.
- Penyediaan Public API Key dan webhook event dispatcher untuk developer eksternal.

---

## 🏗️ 2. Arsitektur & Pola Desain (Clean Architecture)

Backend menerapkan arsitektur bersih yang terpisah secara modular:

```text
HTTP Request (Fiber / Middleware)
  │
  ▼
Handler Layer (HTTP boundary parsing, validator/v10, parameter sanitation)
  │
  ▼
Service Layer (Business rules, cryptographic operations, state transition)
  │
  ▼
Repository / Store Interface (SQL Queries via sqlc / Redis client / MinIO client)
  │
  ▼
PostgreSQL & Redis & MinIO S3
```

### Prinsip Utama:
1. **No ORM Overhead**: Menggunakan SQL murni yang di-compile menjadi kode Go type-safe via `sqlc`.
2. **Zero Leaked Internal Errors**: Seluruh respon error dibungkus dengan envelope standar tanpa membocorkan stack trace database/Redis ke client.
3. **Context Propagation**: `context.Context` diteruskan dari HTTP Fiber hingga database pooling `pgxpool`.
4. **Graceful Shutdown**: Menangani sinyal OS (`SIGINT`, `SIGTERM`) dengan aman untuk menyelesaikan koneksi database dan background jobs.

---

## 📂 3. Penjelasan Detail Struktur Folder & File

### A. Entrypoints (`cmd/`)
- `cmd/server/main.go`: Entrypoint bootstrapping API server HTTP, inisialisasi koneksi PostgreSQL (`pgxpool`), Redis, S3 MinIO, injeksi dependency ke seluruh service & handler, serta konfigurasi graceful shutdown.
- `cmd/worker/main.go`: Background asynchronous job worker yang memproses ingestion dokumen knowledge base dan webhook dispatching secara background.
- `cmd/e2e_tester/main.go`: Program CLI mandiri untuk menguji end-to-end flow autentikasi dan workspace.
- `cmd/diagnose_channel/`: Utility untuk mendiagnosis status koneksi channel WhatsApp.

### B. Database & Query Layer (`db/` & `generated/sqlc/`)
- `db/migrations/`: Skrip SQL migrasi database yang reversible (`up` & `down`):
  - `000001_create_users`: Tabel users, verifikasi email, dan password reset.
  - `000002_create_workspaces_and_subscriptions`: Tabel tenant workspace, workspace_members, dan subscriptions.
  - `000003_create_agent_infrastructure`: Tabel agent_instances dan status runtime.
  - `000004_create_agent_configuration`: Tabel agent_profiles dan business_profiles.
  - `000005_create_knowledge`: Tabel knowledge_sources dan vector/document chunk metadata.
  - `000006_create_channels_conversations_developer`: Tabel channel_connections, conversations, messages, api_keys, dan webhook_endpoints.
  - `000007_create_webhooks`: Tabel webhook event delivery logs.
  - `000008_subscription_first_developer_role`: Penambahan role developer dan penyesuaian tier.
  - `000009_product_movement_enhancements`: Penambahan onboarding_profiles dan konfigurasi business intelligence.
  - `000010_expand_agent_templates`: Template konfigurasi agent industri spesifik.
- `db/queries/`: File SQL murni (`users.sql`, `workspaces.sql`, `conversations.sql`, `knowledge.sql`, dll) yang menjadi sumber kompilasi sqlc.
- `generated/sqlc/`: Kode Go hasil kompilasi otomatis sqlc yang 100% type-safe.

### C. Domain Modules (`internal/`)
- `internal/auth/`:
  - `crypto.go`: Implementasi Argon2id password hashing dan verifikasi.
  - `model.go` & `service.go`: Logika registrasi, login JWT, pembuatan token refresh ber-TTL, dan password reset single-use.
  - `handler.go`: Endpoint HTTP auth (`/api/v1/auth/*`).
  - `middleware.go`: JWT verification middleware dan extraction claims user.
  - `redis.go`: Adapter penyimpanan session refresh token dan rate limiter.
- `internal/workspace/`:
  - Logika pembuatan workspace multi-tenant, manipulasi profil bisnis, subscription check, dan pembagian hak akses (Owner, Admin, Member, Developer).
- `internal/knowledge/`:
  - `converter.go`: Parser dokumen (PDF, Plain Text, Markdown) menjadi text chunks terstruktur.
  - `ingestion.go` & `service.go`: Pipeline ingestion dokumen dan penyimpanan metadata ke SQLite / RAG vaults.
- `internal/conversation/`:
  - `runtime.go`: Interaksi runtime dengan Hermes AI agent, pembuatan dynamic prompt ber-grounding data bisnis.
  - `service.go` & `dashboard.go`: Pengelolaan riwayat chat, pagination cursor pesan, dan agregasi analytics chat.
- `internal/handoff/`:
  - Logika pergantian mode percakapan (`AI` -> `HUMAN` -> `AI`) ketika ada komplain kompleks atau permintaan customer untuk berbicara dengan admin.
- `internal/channel/`:
  - Komunikasi control plane ke WhatsApp Gateway Service via internal HMAC API.
- `internal/hermes/`:
  - `cli.go` & `provider.go`: Adapter pemanggilan Hermes CLI runner dengan isolasi profil per workspace (`cs<workspace_uuid>`).
- `internal/developer/`:
  - Sub-modul `apikey/`, `webhook/`, dan `publicapi/` untuk integrasi developer eksternal.
- `internal/middleware/`:
  - `internal_hmac.go`: Validasi tanda tangan HMAC-SHA256 untuk route internal.
  - `middleware.go`: Security headers, CORS, request ID, dan panic recovery.

---

## 🔒 4. Model Keamanan & Autentikasi

1. **Password Security**: Menggunakan **Argon2id** (64 MiB Memory, 3 Iterations, 2 Parallelism) dengan salt acak kriptografis 16 byte per user.
2. **Access & Refresh Tokens**:
   - Access Token: JWT (HS256) dengan masa berlaku pendek (contoh: 15 menit), berisi `sub` (user ID) dan `jti`.
   - Refresh Token: Opaque random string (32 bytes), di-hash sebelum disimpan di Redis dengan TTL terkonfigurasi. Rotasi wajib dilakukan setiap token digunakan.
3. **Rate Limiting Terdistribusi**: Menggunakan Redis Sliding Window rate limiter untuk mencegah serangan brute-force pada endpoint otentikasi.
4. **Internal Service HMAC**: Seluruh endpoint antar microservice (`/internal/v1/*`) mewajibkan header `X-ChatSolv-Timestamp` dan `X-ChatSolv-Signature` (`hex(HMAC-SHA256(secret, timestamp + "." + body))`) dengan batas toleransi 5 menit.

---

## 🛠️ 5. Panduan Menjalankan & Perintah Makefile (DX)

```bash
# 1. Setup konfigurasi environment
cp .env.example .env

# 2. Menjalankan migrasi database PostgreSQL
make migrate-up

# 3. Rollback migrasi database (jika diperlukan)
make migrate-down

# 4. Regenerate kode database sqlc setelah mengubah query SQL
make sqlc

# 5. Menjalankan unit test suite
make test

# 6. Menjalankan backend server
make run

# 7. Melakukan build binary produksi
make build
```
