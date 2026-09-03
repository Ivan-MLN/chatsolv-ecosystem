# ChatSolv Ecosystem 🌿

Monorepo arsitektur backend, frontend next-gen, gateway WhatsApp bot, dan runtime persistent knowledge AI untuk platform Customer Service Otomatis ChatSolv.

---

## 🏗️ Struktur Arsitektur Monorepo

```text
chatsolv-ecosystem/
├── clientside/          # Next.js 16 (React 19 + Turbopack + Tailwind v4) Landing & Interactive Demo
├── serverside/          # Go Fiber v2 Control Plane, Auth, Postgres, Redis, Hermes Ingestion
├── whatsapp/            # Go whatsmeow WhatsApp Gateway Service (HMAC authenticated)
├── data/                # Persistent volume untuk Hermes Profile, Vaults Markdown, & MinIO S3
│   ├── hermes/          # Hermes agent profiles, SOUL.md, memories, skills, toolsets
│   ├── vaults/          # RAG Knowledge base markdown per workspace
│   └── minio/           # Object storage files & avatars
└── README.md            # Dokumentasi root ekosistem (DX Guide)
```

---

## 🚀 Peta Port & Layanan Lokal

| Layanan | Direktori | Port Default | Deskripsi |
| :--- | :--- | :--- | :--- |
| **Clientside** | `./clientside` | `3333` | Web landing page, presentation, & interactive live demo |
| **Serverside API** | `./serverside` | `3050` / `8080` | Backend API, Auth, Database Control Plane |
| **WhatsApp Gateway** | `./whatsapp` | `4010` | Whatsmeow WhatsApp Web Socket & HMAC webhook dispatcher |
| **MinIO Storage** | `./data/minio` | `9000` (API) / `9001` (Console) | S3 Object storage untuk media & dokumen |
| **Redis** | Server-wide | `6379` | Session caching, distributed rate limiting, and queues |
| **PostgreSQL** | Server-wide | `5433` / `5432` | Primary transactional relational database |

---

## 🛠️ Panduan Developer (DX Quickstart)

### 1. Clientside (Next.js 16)
```bash
cd clientside
npm install
npm run dev        # Dev server pada port 3333 (next dev -p 3333)
npm run build      # Production Turbopack build
npm run start -- -p 3333 # Production server
```

### 2. Serverside (Go Fiber Backend)
```bash
cd serverside
cp .env.example .env
make migrate-up    # Eksekusi migrasi PostgreSQL
make run           # Menjalankan backend server
make test          # Unit & integration testing
```

### 3. WhatsApp Gateway (Go whatsmeow)
```bash
cd whatsapp
cp .env.example .env
make run           # Menjalankan WhatsApp gateway
```

---

## 🔒 Keamanan & Inter-Service Communication

1. **HMAC-SHA256 Authentication**: Komunikasi antara `serverside` dan `whatsapp` diproteksi tanda tangan HMAC dengan timestamp replay tolerance 5 menit.
2. **Hermes Security Boundary**: Agent profile di `./data/hermes` menggunakan `SOUL.md` terisolasi (bisnis CS-only, prompt-injection defense, anti privilege escalation).
3. **Data Isolation**: Database tenant dan knowledge vaults dipartisi per `workspace_id`.
