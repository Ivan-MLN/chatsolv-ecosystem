# ChatSolv Ecosystem 🌿
*The Next-Generation AI-Driven Customer Service & WhatsApp Automation Monorepo*

---

## 🌟 1. Project Introduction (Apa itu ChatSolv?)

**ChatSolv** adalah platform otomasi Customer Service WhatsApp pintar dan infrastruktur komunikasi bisnis 24/7 generasi baru. 

Banyak bisnis kehilangan calon pembeli dan menghadapi churn pelanggan bukan karena produk mereka buruk, melainkan karena **lambatnya respons chat ketika pelanggan sedang memiliki niat beli (buying intent)**. ChatSolv memecahkan masalah ini dengan menggabungkan:
1. **AI Agent Berbasis Pengetahuan Nyata Bisnis**: Agent AI yang tidak halusinasi, dilatih dan di-grounding dengan Standard Operating Procedure (SOP), katalog produk, FAQ, dan kebijakan operasional masing-masing workspace/tenant.
2. **WhatsApp Multi-Session Gateway Tanpa API Resmi yang Kaku**: Menggunakan protokol soket modern (`whatsmeow`) yang memungkinkan bisnis menghubungkan nomor WhatsApp yang sudah ada dalam hitungan detik tanpa setup kompleks.
3. **Smart Human-Agent Handoff**: Memisahkan pesan otomatis berulang (stok, harga, FAQ) dari percakapan bernilai tinggi (negosiasi tender, komplain darurat) yang otomatis diteruskan ke admin manusia via live take-over.
4. **Keamanan Ketat & Tenant Isolation**: Arsitektur multi-tenant dengan partisi data terisolasi, enkripsi token, tanda tangan HMAC internal, dan proteksi prompt injection level enterprise.

Ekosistem repositori ini (`chatsolv-ecosystem`) merupakan monorepo lengkap yang menyatukan seluruh sub-sistem ChatSolv: frontend presentasi next-gen, backend control plane, WhatsApp bot socket service, serta persistent memory & storage AI.

---

## 🏗️ 2. Top-Level Monorepo Architecture

```text
chatsolv-ecosystem/
├── clientside/          # Next.js 16 (React 19 + Turbopack + Tailwind v4 + Framer Motion)
├── serverside/          # Go Fiber v2 Control Plane, Auth, PostgreSQL, Redis, RAG Pipeline
├── whatsapp/            # Go whatsmeow WhatsApp Gateway Service (HMAC Secure Dispatcher)
├── data/                # Persistent Local Storage Volume (Hermes AI Profiles, Vaults, MinIO)
└── README.md            # Dokumentasi Utama Ekosistem Monorepo
```

### Diagram Interaksi Antar Layanan

```text
                 [ Browser / Web User ]
                           │
                           ▼ :3333 (Cloudflare Tunnel: cs.naeladtya.my.id)
                  ┌─────────────────┐
                  │   clientside    │ (Next.js 16 Landing & Interactive Demo)
                  └─────────────────┘
                           │
                           ▼ :3050 (api-cs.naeladtya.my.id)
                  ┌─────────────────┐
                  │   serverside    │ ◄───► [ PostgreSQL :5433 ] & [ Redis :6379 ]
                  └─────────────────┘
                     │           ▲
        HMAC API     │           │ HMAC Webhooks
     :4010 Connect   ▼           │ :3050 Incoming Msgs
                  ┌─────────────────┐
                  │    whatsapp     │ ◄───► [ WhatsApp Web Protocol (whatsmeow) ]
                  └─────────────────┘
                           ▲
                           │
                    [ Customer WA ]
```

---

## 🗺️ 3. Port & Service Directory Map

| Service | Direktori | Port Default | Domain / Routing | Fungsi Utama |
| :--- | :--- | :--- | :--- | :--- |
| **Clientside** | `./clientside` | `3333` | `cs.naeladtya.my.id` | Single-Page Sage Green UI, Framer Motion Scrubber, & Live Interactive Chat Simulator |
| **Serverside API** | `./serverside` | `3050` / `8080` | `api-cs.naeladtya.my.id` | Auth (JWT/Argon2id), Workspace CRUD, Knowledge Ingestion, Handoff Management |
| **WhatsApp Gateway**| `./whatsapp` | `4010` | Internal Service | WhatsApp Web session lifecycle, QR code generator, event dispatcher |
| **MinIO S3 Storage**| `./data/minio` | `9000` / `9001` | Local S3 Store | Object storage untuk dokumen attachment, avatar, & invoice |
| **PostgreSQL** | System-wide | `5433` (atau `5432`)| Local DB | Source of truth untuk users, workspaces, membership, conversations |
| **Redis** | System-wide | `6379` | Local In-Memory | Refresh token rotation, temporary state, rate limiting, and queues |

---

## 📂 4. Detailed Component & Directory Breakdown

### 🎨 A. `clientside/` (Next-Generation Frontend)
Frontend modern dengan arsitektur **Pinned Single-Viewport Carousel**:
- **Framework**: Next.js 16 (Turbopack compiler), React 19.
- **Styling**: Tailwind CSS v4 dengan CSS Variable Tokens tema **Sage Green** (`#618264`, `#79AC78`, `#B0D9B1`, `#D0E7D2`, `#d6ebd8`).
- **Animasi**: GPU-accelerated animated fluid ambient background mesh & word-by-word blur writer typography reveal.
- **Fitur Utama**:
  - *Slide 1 (Welcome)*: Editorial 2-baris headline *"Pelanggan Tidak Menghilang Tiba-Tiba. Mereka Berhenti Menunggu."*
  - *Slide 2 (Demo Interaktif)*: Live chat window WhatsApp-style dengan glassmorphism panel, typing indicator, auto-scroll, dan smart simulated auto-replies.
  - *Slide 3 (Coming Soon)*: Pengumuman rilis fitur yang seamless dan terintegrasi.
- **DX Commands**:
  ```bash
  cd clientside
  npm install
  npm run dev -- -p 3333    # Local development server
  npm run lint              # ESLint checks
  npm run build             # Production Turbopack build
  npm run start -- -p 3333  # Run production build
  ```

### ⚙️ B. `serverside/` (Go Fiber Backend Platform)
Control plane backend berkinerja tinggi yang dibangun dengan prinsip Clean Architecture:
- **Core Stack**: Go (1.23+), GoFiber v2, PostgreSQL (pgx/v5 + sqlc), Redis.
- **Security**: Argon2id password hashing, JWT HS256 access token, Opaque Redis-backed refresh token rotation, Distributed rate limiting.
- **Arsitektur Folder `serverside/`**:
  - `cmd/server/`: Entrypoint bootstrapping API server, dependency injection, graceful shutdown.
  - `cmd/worker/`: Asynchronous background worker untuk knowledge document parsing dan sync jobs.
  - `internal/auth/`: Handler & service otentikasi (Register, Login, Password Reset, Refresh Token).
  - `internal/workspace/`: Multi-tenant workspace management, role permissions (Owner, Admin, Member, Developer).
  - `internal/knowledge/`: RAG document conversion (PDF/TXT/MD), chunking, dan storage ke markdown vaults.
  - `internal/conversation/`: Ingestion chat, tracking percakapan, dan automated agent runtime.
  - `internal/handoff/`: Mekanisme pergantian status percakapan dari AI ke Admin manusia (Takeover/Resume).
  - `internal/middleware/`: HMAC validation middleware, CORS, request logging (`log/slog`), rate limiter.
  - `db/migrations/`: SQL reversible migrations (up/down).
  - `db/queries/`: SQL queries yang di-compile menjadi type-safe Go code via `sqlc`.
- **DX Commands**:
  ```bash
  cd serverside
  cp .env.example .env
  make migrate-up          # Menjalankan migrasi database
  make run                 # Menjalankan server lokal
  make test                # Menjalankan unit test suite
  make sqlc                # Regenerate Go database code
  ```

### 📱 C. `whatsapp/` (Go whatsmeow Bot Service)
Service microservice mandiri yang bertugas mengelola koneksi soket WhatsApp Web:
- **Core Engine**: `whatsmeow` library.
- **Penyimpanan Sesi**: Modernc SQLite database terisolasi per channel di `./data/sessions/<channel_id>.db`.
- **Security Model**:
  - Seluruh endpoint `/internal/v1/*` diproteksi **HMAC-SHA256** menggunakan secret bersama (`INTERNAL_SERVICE_SECRET`).
  - Header: `X-ChatSolv-Timestamp` (toleransi replay 5 menit) & `X-ChatSolv-Signature`.
- **Fitur Utama**:
  - Endpoint `POST /internal/v1/channels/connect` untuk inisialisasi sesi & pairing QR code.
  - Endpoint `POST /internal/v1/channels/disconnect` untuk logout channel.
  - HMAC-signed Webhook callback ke backend saat ada pesan masuk atau perubahan status koneksi.
- **DX Commands**:
  ```bash
  cd whatsapp
  cp .env.example .env
  make run                 # Menjalankan WhatsApp gateway server
  make test                # Menjalankan unit test callback & HMAC
  ```

### 💾 D. `data/` (Persistent Storage Volume)
Pusat penyimpanan lokal persisten untuk file dan knowledge:
- `data/hermes/profiles/`: Profil runtime AI Hermes per workspace, berisi instruksi master `SOUL.md`, memori jangka panjang `memories/`, serta skill tools `skills/`.
- `data/vaults/`: Direktori markdown terstruktur berisi knowledge base bisnis (`products/`, `faq/`, `policies/`).
- `data/minio/`: Volume lokal MinIO S3 untuk penyimpanan file attachment dan media upload.

---

## 🔒 5. Enterprise Security & Isolation Rules

1. **Strict Business AI Scope**:
   Agent AI Hermes diikat dengan aturan `SOUL.md` yang melarang eksekusi instruksi di luar lingkup Customer Service (anti prompt-injection, jailbreak protection, melarang eksekusi coding/terminal).
2. **Tenant Data Partitioning**:
   Setiap data (conversations, messages, API keys, webhook endpoints, dan knowledge vaults) memiliki foreign key `workspace_id` dan di-resolve berdasarkan authenticated membership.
3. **Single-Use Cryptographic Tokens**:
   Password reset token dan email verification token bersifat cryptographically secure (`crypto/rand`), ber-TTL, dan hangus dalam sekali pakai (single-use).

---

## 🛠️ 6. Quick Start Guide untuk Developer Baru (DX)

1. **Clone Repositori**:
   ```bash
   git clone https://github.com/Ivan-MLN/chatsolv-ecosystem.git
   cd chatsolv-ecosystem
   ```
2. **Setup Environment**:
   - Salin `.env.example` di masing-masing subfolder (`serverside/.env.example` -> `serverside/.env`, `whatsapp/.env.example` -> `whatsapp/.env`).
   - Pastikan `INTERNAL_SERVICE_SECRET` sama antara `serverside` dan `whatsapp`.
3. **Jalankan Layanan**:
   - Jalankan `serverside` via `make run`
   - Jalankan `whatsapp` via `make run`
   - Jalankan `clientside` via `npm run dev -- -p 3333`
