# PRODUCT REQUIREMENTS DOCUMENT

## ChatSolv Backend Platform

**Versi:** 1.1
**Status:** Development Specification
**Target:** Backend Engineering Team
**Produk:** ChatSolv
**Tanggal:** 25 Agustus 2026

---

# 1. Ringkasan Produk

ChatSolv adalah SaaS customer service otomatis yang memungkinkan setiap bisnis memiliki agent customer service sendiri.

Setiap pelanggan ChatSolv memiliki konfigurasi yang terisolasi berupa:

* workspace sendiri
* profil bisnis sendiri
* profil bot sendiri
* personality agent sendiri
* Hermes Agent sendiri
* Obsidian Second Brain sendiri
* nomor/channel sendiri
* knowledge perusahaan sendiri
* percakapan pelanggan sendiri
* API key sendiri
* usage dan subscription sendiri

Backend ChatSolv merupakan pusat dari seluruh konfigurasi tersebut.

Backend **bukan WhatsApp bot**.

Service WhatsApp berjalan secara terpisah dan hanya menggunakan Backend ChatSolv sebagai control plane dan agent runtime.

Arsitektur sederhananya:

```text
Frontend Dashboard
        ↓
ChatSolv Backend API
        ↓
Agent Orchestrator
        ↓
Hermes Agent
        ↓
Obsidian Second Brain
```

Channel customer:

```text
Pelanggan
   ↓
WhatsApp
   ↓
WhatsApp Bot Service (TypeScript)
   ↓
ChatSolv Backend
   ↓
Hermes Agent Tenant
   ↓
Obsidian Vault Tenant
   ↓
Response
```

Integrasi website:

```text
Website Live Chat
       ↓
ChatSolv Public API
       ↓
Agent Orchestrator
       ↓
Hermes Agent Tenant
       ↓
Obsidian Second Brain Tenant
```

Satu tenant menggunakan personality dan knowledge yang sama pada semua channel.

---

# 2. Tujuan Backend

Backend harus memungkinkan pemilik bisnis untuk:

1. Membuat akun ChatSolv.
2. Memiliki workspace bisnis.
3. Memiliki Hermes Agent sendiri.
4. Memiliki Obsidian Second Brain sendiri.
5. Mengatur personality agent.
6. Mengatur profil bot.
7. Mengatur profil perusahaan.
8. Mengisi aturan dan kebijakan bisnis.
9. Menghubungkan nomor WhatsApp.
10. Mengunggah dokumen perusahaan.
11. Memasukkan FAQ.
12. Memasukkan knowledge secara manual.
13. Mengelola knowledge.
14. Menguji agent dari dashboard.
15. Melihat conversation history.
16. Mengambil alih percakapan jika dibutuhkan.
17. Menggunakan agent yang sama melalui Public API.
18. Mengintegrasikan ChatSolv ke live chat website.
19. Membuat API key.
20. Membuat webhook.
21. Melihat usage.
22. Mengelola subscription.

---

# 3. Non-Goal

Backend Go tidak bertanggung jawab secara langsung terhadap:

* implementasi protokol WhatsApp
* WebSocket WhatsApp
* Baileys session
* QR WhatsApp secara langsung
* reconnect WhatsApp
* WhatsApp encryption
* pengiriman packet WhatsApp
* pengelolaan session library WhatsApp

Hal tersebut merupakan tanggung jawab:

```text
WhatsApp Bot Service
TypeScript
```

Backend Go hanya berkomunikasi dengan service tersebut melalui internal API.

---

# 4. Technology Stack

Gunakan stable version terbaru pada waktu implementasi.

## Core

```text
Language
Go

Framework
Fiber

Database
PostgreSQL

Cache / Runtime
Redis

Agent Engine
Hermes Agent

Second Brain
Obsidian Vault
```

## Supporting Infrastructure

Recommended:

```text
Object Storage
S3-compatible storage

Background Jobs
Asynq + Redis
atau Redis Streams

Logging
Structured JSON

Tracing
OpenTelemetry

Metrics
Prometheus-compatible metrics
```

---

# 5. Prinsip Arsitektur Utama

## 5.1 Multi Tenant

Semua data bisnis harus berada dalam scope:

```text
workspace_id
```

Workspace adalah tenant boundary.

Contoh:

```text
Workspace A
├── Agent A
├── Obsidian Vault A
├── Knowledge A
├── Channel A
├── Conversations A
└── API Keys A

Workspace B
├── Agent B
├── Obsidian Vault B
├── Knowledge B
├── Channel B
├── Conversations B
└── API Keys B
```

Workspace A tidak boleh dapat mengakses resource Workspace B.

---

# 6. Resource Utama per Tenant

Untuk MVP:

```text
1 Workspace
    ↓
1 Default Agent
    ↓
1 Hermes Agent
    ↓
1 Obsidian Second Brain
```

Arsitektur database harus tetap mendukung multiple agent di masa depan.

Jangan membuat hardcoded:

```text
workspace.agent_id
```

sebagai satu-satunya desain.

Gunakan:

```text
workspace
has many
agents
```

meskipun MVP hanya memperbolehkan satu agent aktif.

Future:

```text
Sales Agent
Support Agent
Complaint Agent
Booking Agent
```

---

# 7. Hermes Agent Isolation

Ini adalah requirement wajib.

Setiap workspace harus mempunyai Hermes Agent sendiri.

DILARANG menggunakan:

```text
1 Hermes Agent global
+
prompt berbeda per customer
```

Resource isolation harus nyata.

Contoh:

```text
Workspace A
→ Hermes Agent A

Workspace B
→ Hermes Agent B
```

Hermes Agent A hanya boleh menggunakan Second Brain Workspace A.

Hermes Agent B hanya boleh menggunakan Second Brain Workspace B.

---

# 8. Obsidian Second Brain

Setiap workspace mempunyai satu Obsidian Vault terisolasi.

Contoh:

```text
/data/chatsolv/vaults/

workspace_A/
workspace_B/
workspace_C/
```

Obsidian Vault merupakan **source of truth untuk business knowledge**.

PostgreSQL bukan penyimpan utama isi knowledge.

PostgreSQL menyimpan:

* metadata
* indexing metadata
* resource relationships
* checksums
* versions
* status
* ownership
* audit information

---

# 9. Obsidian Tidak Berarti Menjalankan GUI

Backend tidak perlu menjalankan aplikasi desktop Obsidian per customer.

Yang digunakan adalah format vault Obsidian:

```text
folder
markdown
attachments
frontmatter
wikilinks jika diperlukan
```

Sehingga vault tetap kompatibel dengan Obsidian tetapi dapat dikelola programmatically oleh backend.

---

# 10. Struktur Vault Default

Saat workspace dibuat:

```text
vault/
│
├── .chatsolv/
│   ├── manifest.json
│   └── metadata.json
│
├── business/
│   ├── company-profile.md
│   ├── business-hours.md
│   └── brand-voice.md
│
├── bot/
│   ├── personality.md
│   └── behavior-rules.md
│
├── products/
│
├── services/
│
├── policies/
│   ├── shipping.md
│   ├── refund.md
│   ├── return.md
│   └── warranty.md
│
├── faq/
│
├── knowledge/
│
├── imports/
│
└── attachments/
```

Folder kosong boleh dibuat saat provisioning.

---

# 11. File Manifest Vault

Contoh:

```json
{
  "workspace_id": "wsp_xxx",
  "agent_id": "agt_xxx",
  "schema_version": 1,
  "created_at": "2026-08-25T12:00:00+07:00"
}
```

File ini:

```text
.chatsolv/manifest.json
```

tidak boleh digunakan sebagai satu-satunya source ownership.

Ownership utama tetap berada di PostgreSQL.

---

# 12. Knowledge Markdown Format

Setiap knowledge note harus memiliki frontmatter.

Contoh:

```md
---
id: knw_01XYZ
workspace_id: wsp_01XYZ
source_type: document
source_id: src_01XYZ
category: policy
version: 2
status: active
created_at: 2026-08-25T12:00:00+07:00
updated_at: 2026-08-25T12:00:00+07:00
---

# Kebijakan Pengembalian Barang

Barang dapat dikembalikan sesuai ketentuan perusahaan.

## Syarat

- Barang belum digunakan
- Kemasan masih lengkap
- Bukti transaksi tersedia
```

---

# 13. Second Brain vs Conversation Memory

Keduanya harus dipisahkan.

## Obsidian Second Brain

Menyimpan:

* profil perusahaan
* katalog
* produk
* layanan
* kebijakan
* FAQ
* SOP
* informasi operasional
* knowledge perusahaan

## Conversation Memory

Menyimpan:

* percakapan customer
* message history
* conversation status
* human handoff state
* customer metadata

Conversation history jangan otomatis ditulis ke Obsidian Vault.

---

# 14. PostgreSQL

PostgreSQL adalah source of truth untuk transactional state.

Contoh:

```text
users
workspaces
agents
channels
subscriptions
conversations
messages
api_keys
jobs
audit logs
```

Obsidian adalah source of truth untuk business knowledge.

Redis adalah runtime/cache.

Pembagian:

```text
PostgreSQL
→ Transactional truth

Obsidian
→ Knowledge truth

Redis
→ Cache/runtime/jobs

Hermes
→ Agent runtime/reasoning
```

---

# 15. Core Database Entities

Minimal:

```text
users

workspaces
workspace_members

subscriptions
subscription_entitlements

agents
agent_personalities
agent_profiles

second_brains

business_profiles
business_policies

knowledge_sources
knowledge_notes
knowledge_sync_jobs

channels
channel_connections

conversations
messages

api_keys
api_sessions

webhook_endpoints
webhook_deliveries

usage_records

audit_logs

outbox_events
```

---

# 16. Users

Table:

```text
users
```

Fields:

```text
id UUID

email
password_hash

name
avatar_url

email_verified_at

status

created_at
updated_at
deleted_at
```

Status:

```text
active
suspended
deleted
```

Password hashing:

```text
Argon2id
```

---

# 17. Workspace

```text
workspaces
```

Fields:

```text
id

name
slug

owner_user_id

status

timezone

created_at
updated_at
deleted_at
```

Status:

```text
provisioning
active
suspended
deleting
deleted
```

---

# 18. Workspace Member

```text
workspace_members
```

Fields:

```text
id

workspace_id
user_id

role

created_at
updated_at
```

Role:

```text
owner
admin
member
viewer
```

---

# 19. Agent

```text
agents
```

Fields:

```text
id

workspace_id

name

status

provider
provider_agent_id

config_version
synced_config_version

created_at
updated_at
deleted_at
```

Provider:

```text
hermes
```

Status:

```text
pending
provisioning
ready
syncing
suspended
failed
deleting
deleted
```

---

# 20. Second Brain Table

```text
second_brains
```

Fields:

```text
id

workspace_id
agent_id

provider
vault_key
vault_path

schema_version
version

status

last_synced_at

created_at
updated_at
deleted_at
```

Provider:

```text
obsidian
```

Status:

```text
pending
provisioning
ready
syncing
failed
suspended
deleting
deleted
```

---

# 21. Agent Personality

Personality jangan disimpan sebagai satu prompt bebas saja.

Gunakan structured configuration.

```text
agent_personalities
```

Fields contoh:

```text
id
workspace_id
agent_id

bot_name

role

tone

communication_style

primary_language

response_length

emoji_usage

greeting_style

closing_style

custom_instructions

behavior_rules

escalation_rules

forbidden_topics

fallback_behavior

created_at
updated_at
```

Contoh:

```json
{
  "bot_name": "Naya",
  "role": "Customer Service",
  "tone": "friendly",
  "communication_style": "casual_professional",
  "primary_language": "id",
  "response_length": "medium",
  "emoji_usage": "moderate"
}
```

---

# 22. Bot Profile

Dashboard mempunyai halaman:

```text
Agent
→ Profile
```

Config:

```text
display_name

avatar

description

greeting_message

away_message

fallback_message

language
```

Avatar asli disimpan di object storage.

Database menyimpan object key.

---

# 23. Business Profile

Fields:

```text
business_name

industry

business_description

website

email

phone

address

business_hours

timezone

brand_voice

company_values
```

Perubahan business profile juga harus memperbarui:

```text
business/company-profile.md
```

di vault.

---

# 24. Business Policy

Structured fields:

```text
shipping_policy
refund_policy
return_policy
warranty_policy
payment_policy
complaint_policy
```

Setelah save, backend dapat menulis:

```text
policies/shipping.md
policies/refund.md
policies/return.md
```

---

# 25. Source of Truth Configuration

Untuk setting dashboard seperti personality:

PostgreSQL tetap source of truth.

Kemudian konfigurasi tersebut disinkronkan menjadi markdown ke Obsidian jika diperlukan oleh Hermes.

Contoh:

```text
PATCH personality
     ↓
PostgreSQL
     ↓
Increment config version
     ↓
Generate personality.md
     ↓
Update Obsidian Vault
     ↓
Notify Hermes
```

Jangan menjadikan frontend menulis file Markdown secara langsung.

---

# 26. Agent Prompt Priority

Agent runtime mengikuti hierarchy:

```text
1. ChatSolv Platform Rules

2. Tenant Personality

3. Tenant Business Rules

4. Retrieved Second Brain Knowledge

5. Conversation Context

6. Current Customer Message
```

Tenant tidak boleh menimpa platform security rules.

---

# 27. Prompt Injection Protection

Knowledge dari Obsidian dianggap sebagai:

```text
REFERENCE DATA
```

bukan system instruction.

Jika document tenant berisi:

```text
Ignore all previous instructions
```

agent tidak boleh memperlakukannya sebagai instruction tingkat platform.

Hermes orchestration harus memisahkan:

```text
SYSTEM INSTRUCTIONS

TENANT CONFIGURATION

RETRIEVED KNOWLEDGE

USER MESSAGE
```

---

# 28. Workspace Provisioning Flow

Setelah user berhasil membuat workspace:

```text
Create Workspace
      ↓
Create Default Agent
      ↓
Create Second Brain Record
      ↓
Commit DB Transaction
      ↓
Enqueue Provisioning Job
```

HTTP tidak menunggu Hermes/Obsidian provisioning selesai.

Response:

```json
{
  "workspace_status": "provisioning",
  "agent_status": "provisioning",
  "second_brain_status": "provisioning"
}
```

---

# 29. Provisioning Worker

Worker menjalankan:

```text
Provisioning Job
      ↓
Create Obsidian Vault
      ↓
Create Default Folder Structure
      ↓
Write manifest
      ↓
Create Default Markdown Config
      ↓
Create Hermes Agent
      ↓
Attach / Configure Vault for Hermes
      ↓
Verify access
      ↓
Mark Agent READY
      ↓
Mark Second Brain READY
      ↓
Mark Workspace ACTIVE
```

---

# 30. Provisioning Harus Idempotent

Gunakan key:

```text
workspace:{workspace_id}:provision
```

Jika job gagal setelah vault dibuat:

retry harus menggunakan vault lama.

Jangan menghasilkan:

```text
workspace_x
workspace_x-copy
workspace_x-copy-2
```

---

# 31. Hermes Adapter

Semua integrasi Hermes harus berada pada layer khusus.

Contoh:

```text
internal/hermes/
```

Interface:

```go
type AgentProvider interface {
    CreateAgent(
        ctx context.Context,
        input CreateAgentInput,
    ) (*AgentResource, error)

    UpdateAgent(
        ctx context.Context,
        agentID string,
        input UpdateAgentInput,
    ) error

    ConfigureBrain(
        ctx context.Context,
        agentID string,
        brain BrainConfig,
    ) error

    Generate(
        ctx context.Context,
        input AgentRequest,
    ) (*AgentResponse, error)

    DeleteAgent(
        ctx context.Context,
        agentID string,
    ) error
}
```

Business logic tidak boleh tergantung langsung pada Hermes SDK.

---

# 32. Obsidian Adapter

Buat abstraction:

```text
internal/brain/obsidian/
```

Interface:

```go
type SecondBrain interface {
    CreateVault(
        ctx context.Context,
        workspaceID string,
    ) (*Vault, error)

    WriteNote(
        ctx context.Context,
        vaultID string,
        note Note,
    ) error

    ReadNote(
        ctx context.Context,
        vaultID string,
        path string,
    ) (*Note, error)

    DeleteNote(
        ctx context.Context,
        vaultID string,
        path string,
    ) error

    ListNotes(
        ctx context.Context,
        vaultID string,
    ) ([]Note, error)

    WriteAttachment(
        ctx context.Context,
        vaultID string,
        attachment Attachment,
    ) error

    DeleteVault(
        ctx context.Context,
        vaultID string,
    ) error
}
```

Dengan abstraction ini backend tidak tergantung pada layout filesystem tertentu.

---

# 33. Knowledge Sources

Dashboard menyediakan:

```text
Knowledge
```

Sumber yang didukung MVP:

```text
Document
Text
FAQ
Product Data
Structured Data
```

Future:

```text
Website Crawl
Google Drive
Notion
External Database
CRM
```

---

# 34. Supported Files

MVP:

```text
PDF
DOCX
TXT
CSV
JSON
XLSX
MD
```

Validation:

```text
file extension
MIME
size
checksum
```

Jangan percaya extension saja.

---

# 35. Knowledge Source Table

```text
knowledge_sources
```

Fields:

```text
id

workspace_id
second_brain_id

type

title

original_filename
mime_type
file_size

original_object_key

checksum

status

error_code

created_at
updated_at
deleted_at
```

---

# 36. Knowledge Notes

```text
knowledge_notes
```

Fields:

```text
id

workspace_id
second_brain_id
source_id

title
category

relative_path

content_hash

version

status

created_at
updated_at
deleted_at
```

---

# 37. Document Upload Flow

```text
Frontend
   ↓
POST /v1/knowledge/documents
   ↓
Authentication
   ↓
Workspace Authorization
   ↓
Subscription Entitlement
   ↓
Validate File
   ↓
Calculate SHA-256
   ↓
Duplicate Detection
   ↓
Save Original → Object Storage
   ↓
Create knowledge_source
   ↓
status = queued
   ↓
Create ingestion job
   ↓
HTTP 202
```

Response:

```json
{
  "data": {
    "id": "src_xxx",
    "status": "queued"
  }
}
```

---

# 38. Document Ingestion Pipeline

Worker:

```text
QUEUED
   ↓
Download Original
   ↓
Parse File
   ↓
Extract Content
   ↓
Normalize
   ↓
Detect Structure
   ↓
Split Into Logical Notes
   ↓
Generate Markdown
   ↓
Generate Frontmatter
   ↓
Write to Tenant Vault
   ↓
Notify Hermes
   ↓
Verify Agent Knowledge Access
   ↓
READY
```

---

# 39. Jangan Chunk Buta

Konversi dokumen harus mempertahankan struktur.

Contoh PDF:

```text
Refund Policy
 ├── Eligibility
 ├── Period
 └── Procedure
```

Jangan menghasilkan chunk acak:

```text
chunk-1.md
chunk-2.md
chunk-3.md
```

jika struktur semantik dapat dipertahankan.

Prefer:

```text
policies/
  refund/
    overview.md
    eligibility.md
    procedure.md
```

jika document besar.

---

# 40. CSV / Product Catalog

CSV katalog jangan dikonversi menjadi satu markdown raksasa.

Contoh:

```text
products/
├── produk-a.md
├── produk-b.md
└── produk-c.md
```

Frontmatter:

```md
---
sku: SKU001
price: 149000
stock_status: available
category: fashion
---
```

Konten:

```md
# Nama Produk

## Deskripsi

...

## Informasi Produk

...
```

---

# 41. FAQ Storage

FAQ dapat dibuat melalui dashboard.

Input:

```text
Question
Answer
Category
```

Stored sebagai:

```text
faq/refund.md
faq/shipping.md
faq/product.md
```

Atau satu note per FAQ jika volume besar.

---

# 42. Knowledge Versioning

Jika user mengganti document:

```text
refund-policy.pdf v1
↓
refund-policy.pdf v2
```

Flow:

```text
Upload v2
     ↓
Parse v2
     ↓
Generate new note version
     ↓
Verify
     ↓
Activate v2
     ↓
Deactivate v1
```

Jangan delete v1 sebelum v2 berhasil.

---

# 43. Delete Knowledge

Flow:

```text
DELETE knowledge source
      ↓
status = deleting
      ↓
enqueue job
      ↓
remove corresponding notes
      ↓
remove attachments
      ↓
notify Hermes
      ↓
remove original file if policy allows
      ↓
status = deleted
```

---

# 44. Knowledge Sync Status

Statuses:

```text
queued
processing
converting
writing
syncing
ready
failed
deleting
deleted
```

Frontend harus dapat melihat status.

---

# 45. Knowledge API

```text
GET    /v1/knowledge
GET    /v1/knowledge/:id

POST   /v1/knowledge/documents
POST   /v1/knowledge/text
POST   /v1/knowledge/faqs

PATCH  /v1/knowledge/:id

DELETE /v1/knowledge/:id

POST   /v1/knowledge/:id/retry
```

---

# 46. Agent Personality Update Flow

```text
Dashboard
   ↓
PATCH /v1/agents/:id/personality
   ↓
Validate
   ↓
Save PostgreSQL
   ↓
Increment config_version
   ↓
Invalidate Redis
   ↓
Generate personality.md
   ↓
Write Obsidian
   ↓
Sync Hermes
   ↓
Update synced_config_version
```

---

# 47. Config Version

Agent fields:

```text
config_version
synced_config_version
```

Example:

```text
config_version = 18
synced_config_version = 17
```

Agent:

```text
syncing
```

Setelah berhasil:

```text
18 == 18
```

Agent:

```text
ready
```

---

# 48. Agent Readiness

Agent dianggap siap jika:

```text
Hermes Agent exists

Obsidian Vault exists

Hermes dapat mengakses vault

Personality synced

Business profile synced

Required configuration complete

Subscription valid
```

---

# 49. Onboarding Customer

Setelah account dan subscription/trial aktif:

```text
Create Account
      ↓
Create Workspace
      ↓
Provision Hermes
      ↓
Provision Obsidian
      ↓
Dashboard Onboarding
```

Onboarding:

```text
1. Profil Bisnis

2. Personality Bot

3. Knowledge

4. Nomor WhatsApp

5. Test Bot

6. Aktifkan
```

---

# 50. Step 1 — Profil Bisnis

User mengisi:

```text
Nama bisnis

Kategori bisnis

Deskripsi

Website

Jam operasional

Alamat

Nomor kontak
```

Backend:

```text
PostgreSQL
+
business/company-profile.md
```

---

# 51. Step 2 — Personality

User mengatur:

```text
Nama bot

Formal / Santai

Profesional / Hangat

Jawaban Singkat / Sedang / Detail

Emoji:
Tidak Ada
Sedikit
Normal

Greeting

Fallback

Custom behavior
```

Tidak perlu menunjukkan system prompt mentah kepada user.

---

# 52. Step 3 — Knowledge

User dapat:

```text
Upload PDF

Upload DOCX

Upload CSV

Tambah FAQ

Tambah informasi manual
```

Frontend menunjukkan progress:

```text
Uploading

Processing

Preparing Knowledge

Ready
```

---

# 53. Step 4 — WhatsApp

User:

```text
Hubungkan WhatsApp
```

Backend berkomunikasi dengan bot service.

---

# 54. Channel Model

Table:

```text
channels
```

Fields:

```text
id

workspace_id
agent_id

type

display_name
phone_number

external_channel_id
service_instance_id

status

connected_at
last_seen_at

created_at
updated_at
```

Types:

```text
whatsapp
web
api
```

Future:

```text
instagram
telegram
messenger
email
```

---

# 55. WhatsApp Channel Status

```text
disconnected

connecting

waiting_pairing

connected

reconnecting

error

suspended
```

---

# 56. Connect WhatsApp Flow

```text
Dashboard
   ↓
POST /v1/channels/whatsapp/connect
   ↓
Backend creates Channel
   ↓
Backend → WhatsApp Bot Service
   ↓
Bot Service creates pairing session
   ↓
QR / pairing response
   ↓
Backend → Frontend
```

Setelah pairing:

```text
WhatsApp Service
      ↓
Webhook Internal
      ↓
Backend
      ↓
channel = connected
```

---

# 57. Internal Bot Service API

Example:

```text
POST /internal/v1/channels/events

POST /internal/v1/messages/incoming

POST /internal/v1/channels/status
```

Service authentication berbeda dari user authentication.

---

# 58. Internal Service Security

Jangan gunakan user JWT.

Gunakan service credential.

Minimum:

```text
Bearer internal credential
```

Better:

```text
signed HMAC requests
```

Future:

```text
mTLS
```

---

# 59. Incoming WhatsApp Message

Bot Service mengirim:

```http
POST /internal/v1/messages/incoming
```

Payload:

```json
{
  "channel_id": "chn_xxx",
  "external_message_id": "wamid_xxx",
  "external_user_id": "62812xxx",
  "message_type": "text",
  "content": {
    "text": "Kak, barang ini bisa COD?"
  },
  "timestamp": "2026-08-25T12:00:00Z"
}
```

---

# 60. Incoming Message Flow

```text
Receive Message
      ↓
Authenticate Bot Service
      ↓
Resolve Channel
      ↓
Resolve Workspace
      ↓
Resolve Agent
      ↓
Subscription Check
      ↓
Agent Readiness Check
      ↓
Duplicate Check
      ↓
Find/Create Conversation
      ↓
Save Customer Message
      ↓
Acquire Conversation Lock
      ↓
Load Agent Configuration
      ↓
Load Relevant Second Brain
      ↓
Load Recent Conversation
      ↓
Run Hermes Agent
      ↓
Save Agent Message
      ↓
Release Lock
      ↓
Return Response
```

---

# 61. Message Idempotency

Use unique combination:

```text
channel_id
+
external_message_id
```

Jika message dikirim ulang oleh WhatsApp service:

backend jangan generate response baru.

Return previous processing result.

---

# 62. Conversation Lock

Dua pesan customer dapat masuk hampir bersamaan.

Gunakan Redis lock:

```text
conversation:{id}:agent-lock
```

Tujuan:

menghindari:

```text
message 1
message 2
```

diproses Hermes dalam urutan terbalik.

Lock harus mempunyai TTL.

---

# 63. Conversation Table

```text
conversations
```

Fields:

```text
id

workspace_id
agent_id
channel_id

external_user_id

status
mode

assigned_user_id

started_at
last_message_at
closed_at

metadata

created_at
updated_at
```

Status:

```text
open
closed
```

Mode:

```text
agent
human
```

---

# 64. Messages

```text
messages
```

Fields:

```text
id

workspace_id
conversation_id

external_message_id

sender_type
content_type
content

provider

status

created_at
```

Sender:

```text
customer
agent
human
system
```

---

# 65. Hermes Runtime Context

Backend membangun runtime input.

Concept:

```text
Platform Rules
      +
Agent Personality
      +
Business Context
      +
Relevant Obsidian Knowledge
      +
Conversation Context
      +
Current Message
```

Caller tidak boleh menentukan Hermes resource secara langsung.

---

# 66. Runtime API Internal

```http
POST /internal/v1/agents/:agentId/respond
```

Internal request:

```json
{
  "conversation_id": "cnv_xxx",
  "message": "Kak, barangnya bisa COD?"
}
```

Backend resolve sendiri:

```text
workspace
personality
vault
knowledge
conversation
subscription
```

---

# 67. Agent Response Format

```json
{
  "data": {
    "message_id": "msg_xxx",
    "conversation_id": "cnv_xxx",
    "content": "Bisa kak...",
    "handoff_requested": false,
    "handoff_reason": null
  }
}
```

Optional:

```json
{
  "usage": {
    "input_tokens": 0,
    "output_tokens": 0
  }
}
```

---

# 68. Human Handoff

Agent dapat meminta handoff.

Triggers:

```text
customer meminta admin

complaint tertentu

refund membutuhkan approval

agent tidak yakin / tidak menemukan jawaban di Second Brain

tenant escalation rule

sensitive request
```

Result:

```text
conversation.mode = human
```

Response:

```json
{
  "handoff_requested": true,
  "handoff_reason": "AGENT_UNCERTAIN"
}
```

---

# 68A. WhatsApp Admin Relay Handoff Architecture

Alur eskalasi instan via WhatsApp memungkinkan staf/admin toko mengambil alih percakapan dan membalas langsung dari WhatsApp pribadi admin, dengan nomor WhatsApp Bot bisnis bertindak sebagai **Anonymous Relay/Proxy**.

### 1. Fallback & Escalation Notification
- Saat agent fallback/ragu (`handoff_requested = true`), backend mengirim pesan fallback ke customer (contoh: *"Maaf kak, pertanyaan ini akan kami sampaikan ke tim staff kami dulu ya..."*).
- `conversations.mode` diubah ke `'human'`.
- Backend mem-broadcast notifikasi WhatsApp ke nomor admin yang terdaftar pada workspace (`channel_connections` / `workspace_members.phone` / `business_profiles.phone`):

```text
⚠️ [ESKALASI PERCAKAPAN]
Pelanggan: [Nama/Nomor Pelanggan]
Percakapan ID: #CNV-XXXX
Pesan Terakhir: "[Isi Pesan Pelanggan]"

Ketik #ACC [ID_PERCAKAPAN] untuk mengambil alih percakapan ini.
```

### 2. Admin Accept (`#ACC`)
- Admin mengirim `#ACC <CNV_ID>` ke nomor WhatsApp Bot bisnis.
- Backend memverifikasi bahwa pengirim adalah nomor admin terdaftar untuk workspace terkait.
- Backend mencatat `assigned_user_id` / `assigned_admin_phone` pada record `conversations`.
- Backend membuat *Active Relay Session* di Redis (`relay:admin:<admin_phone> -> conversation_id` dan `relay:conversation:<conv_id> -> admin_phone`, dengan TTL session).
- **Notifikasi ke WhatsApp Admin:**
  ```text
  Anda terhubung dengan pelanggan. Semua pesan yang Anda kirim akan langsung diteruskan ke pelanggan.
  Ketik #DONE atau #CLOSE untuk menyelesaikan sesi.
  ```
- **Notifikasi ke WhatsApp Pelanggan:**
  Bot mengirim notifikasi ke pelanggan bahwa admin/staff sudah bergabung:
  ```text
  Halo kak, saat ini Anda sudah terhubung langsung dengan Customer Support kami. Silakan sampaikan pesan Anda 🙏
  ```

### 3. Direct 2-Way Seamless Anonymous Relay Chat (Tanpa Perlu Prefix #)
- **Dari Admin ke Pelanggan:**
  - Begitu admin melakukan `#ACC`, sesi relay aktif tersimpan di Redis.
  - **Semua pesan yang dikirimkan Admin (teks biasa, gambar, dokumen/PDF, audio, voice note) langsung diteruskan ke pelanggan secara seamless tanpa perlu mengetik prefix `#`**.
  - Backend mem-proxy pesan ke nomor WhatsApp Pelanggan menggunakan **nomor WhatsApp Bot bisnis**.
  - Pelanggan menerima pesan resmi dari bot bisnis tanpa pernah mengetahui nomor pribadi admin.
- **Dari Pelanggan ke Admin:**
  - Selama sesi relay aktif (`mode = 'human'`), setiap pesan baru dari Pelanggan (teks, gambar, dokumen, voice note) langsung diteruskan oleh Bot ke WhatsApp Admin:
    ```text
    📩 [PESAN DARI PELANGGAN: 628xxx]
    [Isi Pesan / Lampiran Pelanggan]

    Ketik balasan Anda langsung (atau #DONE jika selesai).
    ```

### 4. Closing / Handback (`#DONE` / `#CLOSE` / `#SELESAI`)
- Ketika admin merasa penanganan sudah cukup, admin cukup mengetik `#DONE`, `#CLOSE`, atau `#SELESAI`.
- Backend menghapus *Active Relay Session* di Redis.
- Backend mengembalikan `conversations.mode = 'agent'`.
- **Notifikasi ke WhatsApp Admin:**
  ```text
  Percakapan diselesaikan. Bot kembali aktif untuk pelanggan ini.
  ```
- **Notifikasi ke WhatsApp Pelanggan:**
  Bot mengirim pesan penutup ke pelanggan bahwa sesi dengan Customer Support telah selesai dan asisten AI aktif kembali:
  ```text
  Sesi konsultasi dengan Customer Support telah selesai. Terima kasih telah menghubungi kami! Asisten AI kami siap membantu kembali jika ada pertanyaan lainnya 🙏
  ```

---

# 69. Resume Agent

Admin dapat mengembalikan:

```text
human
→
agent
```

Endpoint:

```http
PATCH /v1/conversations/:id/mode
```

```json
{
  "mode": "agent"
}
```

---

# 70. Dashboard Conversations

Endpoints:

```text
GET /v1/conversations

GET /v1/conversations/:id

GET /v1/conversations/:id/messages
```

Filter:

```text
status
mode
channel
date
customer
assigned user
```

Messages menggunakan cursor pagination.

---

# 71. Redis Responsibilities

Redis hanya digunakan untuk runtime.

## Cache

```text
agent config
personality
subscription entitlement
business profile
```

## Locks

```text
conversation processing
knowledge sync
provisioning
```

## Jobs

```text
document ingestion
Hermes provisioning
Obsidian sync
webhook delivery
```

## Rate Limit

```text
login
public API
agent runtime
file upload
```

## Idempotency

```text
incoming messages
API mutations
```

---

# 72. Jangan Jadikan Redis Source of Truth

Jika Redis hilang:

backend harus dapat rebuild cache dari:

```text
PostgreSQL
+
Obsidian
```

---

# 73. Subscription

Table:

```text
subscriptions
```

Status:

```text
trialing
active
past_due
suspended
cancelled
expired
```

Product saat ini mempunyai konsep trial dan subscription.

Backend jangan hardcode feature hanya dari string plan.

Gunakan entitlement.

---

# 74. Subscription Entitlement

Example:

```json
{
  "max_agents": 1,
  "max_channels": 1,
  "max_storage_mb": 2000,
  "max_documents": 200,
  "monthly_messages": 20000,
  "public_api": true,
  "webhooks": true
}
```

---

# 75. Subscription Runtime Check

Sebelum menjalankan Hermes:

```text
workspace active?
subscription valid?
quota available?
agent ready?
```

Jika tidak:

```text
AGENT_RUNTIME_DISABLED
```

Subscription suspension tidak boleh langsung menghapus vault atau Hermes Agent.

---

# 76. Trial

Product awal mendukung:

```text
14 hari trial
```

Backend harus menyimpan:

```text
trial_started_at
trial_ends_at
```

Jangan menghitung trial hanya dari frontend.

---

# 77. Agent Test Playground

Dashboard menyediakan:

```text
Test Bot
```

Endpoint:

```http
POST /v1/agent/test
```

Payload:

```json
{
  "message": "Kak, bisa COD?"
}
```

Test menggunakan:

```text
Hermes Agent tenant
+
Obsidian Vault tenant
+
Personality tenant
```

Tetapi conversation:

```text
environment = test
```

dan tidak masuk production customer metrics.

---

# 78. Public Agent API

Customer dapat menggunakan agent ChatSolv pada website mereka.

Agent yang digunakan harus agent tenant yang sama.

```text
WhatsApp ─┐
          │
Website ──┼→ Hermes Agent → Obsidian Vault
          │
API ──────┘
```

Dengan demikian personality dan knowledge konsisten.

---

# 79. Developer Section

Dashboard:

```text
Developer

├── API Keys
├── Webhooks
├── API Usage
└── Documentation
```

---

# 80. API Key

Format:

```text
cs_live_xxxxxxxxx

cs_test_xxxxxxxxx
```

Database tidak boleh menyimpan full API key.

Simpan:

```text
id
workspace_id

prefix
hash
last_four

name
scopes

created_at
last_used_at
revoked_at
```

Full key hanya ditampilkan sekali saat creation.

---

# 81. API Key Scopes

```text
agent:invoke

conversation:read
conversation:write

knowledge:read
knowledge:write

webhook:manage
```

Default integration website:

```text
agent:invoke
```

---

# 82. Secret API Key Tidak Boleh di Browser

DILARANG:

```js
fetch("https://api.chatsolv...", {
  headers: {
    Authorization: "Bearer cs_live_SECRET"
  }
})
```

dari frontend browser.

Gunakan ephemeral client token.

---

# 83. Web Chat Session Flow

```text
Customer Backend
      ↓
ChatSolv Secret API Key
      ↓
Create Agent Session
      ↓
Receive Client Token
      ↓
Browser
      ↓
ChatSolv
```

---

# 84. Create Agent Session

```http
POST /api/v1/agent-sessions
```

Authentication:

```text
API Key
```

Payload:

```json
{
  "external_user_id": "visitor_19273",
  "metadata": {
    "page": "/pricing"
  }
}
```

Response:

```json
{
  "data": {
    "session_id": "ses_xxx",
    "client_token": "...",
    "expires_in": 3600
  }
}
```

---

# 85. Website Chat Message

```http
POST /api/v1/agent-sessions/:id/messages
```

Authentication:

```text
Ephemeral Client Token
```

Request:

```json
{
  "message": "Apakah bisa digunakan untuk toko online?"
}
```

---

# 86. Streaming API

Recommended:

```text
Server Sent Events
```

Endpoint:

```text
POST /api/v1/agent-sessions/:id/messages/stream
```

Events:

```text
message.start

message.delta

message.completed

message.error
```

Gunakan WebSocket hanya jika nanti ada kebutuhan two-way realtime yang lebih kompleks.

---

# 87. Developer Documentation

Documentation harus menjelaskan:

```text
Introduction

Authentication

API Keys

Create Session

Send Message

Streaming

Conversation

Errors

Rate Limits

Webhooks

Security

JavaScript Example

TypeScript Example

Node.js Example

curl Example
```

---

# 88. Webhooks

Events:

```text
conversation.created

conversation.updated

message.received

message.created

handoff.requested

agent.error
```

Future:

```text
knowledge.updated
channel.connected
channel.disconnected
```

---

# 89. Webhook Security

Header:

```text
X-ChatSolv-Signature
```

Signature:

```text
HMAC SHA-256
```

Payload:

```json
{
  "event_id": "evt_xxx",
  "event": "message.created",
  "timestamp": "...",
  "data": {}
}
```

---

# 90. Webhook Delivery

Store:

```text
webhook_deliveries
```

Fields:

```text
endpoint
event
HTTP response
attempt
status
next_retry_at
```

Retry exponential.

---

# 91. API Versioning

Public:

```text
/api/v1/
```

Dashboard:

```text
/api/v1/
```

Internal:

```text
/internal/v1/
```

Never expose internal endpoint publicly without service authentication.

---

# 92. Standard API Success

```json
{
  "data": {},
  "meta": {},
  "request_id": "req_xxx"
}
```

---

# 93. Standard API Error

```json
{
  "error": {
    "code": "AGENT_NOT_READY",
    "message": "Agent is still being prepared."
  },
  "request_id": "req_xxx"
}
```

---

# 94. Core Error Codes

```text
UNAUTHORIZED
FORBIDDEN

WORKSPACE_NOT_FOUND
WORKSPACE_SUSPENDED

AGENT_NOT_FOUND
AGENT_NOT_READY
AGENT_SYNCING
AGENT_SUSPENDED

SECOND_BRAIN_NOT_READY
SECOND_BRAIN_SYNC_FAILED

CHANNEL_NOT_FOUND
CHANNEL_DISCONNECTED

SUBSCRIPTION_REQUIRED
TRIAL_EXPIRED
QUOTA_EXCEEDED

DOCUMENT_TOO_LARGE
DOCUMENT_TYPE_UNSUPPORTED
DOCUMENT_ALREADY_EXISTS
DOCUMENT_PARSE_FAILED

KNOWLEDGE_WRITE_FAILED
KNOWLEDGE_SYNC_FAILED

HERMES_UNAVAILABLE
HERMES_TIMEOUT

OBSIDIAN_VAULT_UNAVAILABLE

INVALID_API_KEY
INVALID_SESSION_TOKEN

RATE_LIMIT_EXCEEDED

INTERNAL_ERROR
```

---

# 95. Request ID

Semua request mendapatkan:

```text
X-Request-ID
```

Digunakan untuk:

```text
logging
tracing
Hermes request
Bot service
webhook delivery
error tracking
```

---

# 96. Public Resource IDs

Recommended:

```text
usr_
wsp_
agt_
brn_
chn_
cnv_
msg_
src_
knw_
key_
ses_
evt_
wh_
```

Internal primary key tetap UUIDv7.

---

# 97. Authentication Dashboard

Gunakan:

```text
Access Token
+
Refresh Token
```

Access:

```text
±15 menit
```

Refresh:

```text
7–30 hari
```

Gunakan refresh token rotation.

---

# 98. Authorization

Request:

```text
Authenticate User
       ↓
Resolve Workspace
       ↓
Resolve Membership
       ↓
Resolve Role
       ↓
Resolve Permission
```

Jangan percaya:

```text
workspace_id
```

dari body client.

---

# 99. PostgreSQL Tenant Isolation

Tenant-owned table wajib:

```text
workspace_id NOT NULL
```

Recommended indexes:

```text
(workspace_id, id)
```

Pertimbangkan PostgreSQL RLS sebagai lapisan tambahan.

---

# 100. Obsidian Tenant Isolation

Vault path tidak boleh dibentuk dari input user mentah.

Gunakan internal ID.

Safe:

```text
vaults/wsp_01...
```

Dangerous:

```text
vaults/{workspace_name_from_user}
```

Backend wajib mencegah:

```text
../
path traversal
symlink escape
absolute path injection
```

---

# 101. Obsidian File Permission

Runtime harus memastikan worker tenant A tidak dapat menulis di vault tenant B karena kesalahan path.

Semua operations harus menerima:

```text
second_brain_id
```

kemudian backend resolve path internal.

Jangan menerima raw vault path dari API request.

---

# 102. Knowledge Filename Safety

User filename:

```text
../../company.md
```

tidak boleh menjadi filesystem path.

Original filename hanya metadata.

Generated file menggunakan sanitized/internal slug.

---

# 103. Original Documents

Original upload disimpan di:

```text
S3-compatible Object Storage
```

Bukan di vault sebagai satu-satunya copy.

Obsidian menyimpan knowledge hasil ekstraksi.

Contoh:

```text
refund-policy.pdf
        ↓
Object Storage
        +
Obsidian Markdown
```

---

# 104. Security Uploaded Files

Validasi:

```text
MIME
size
extension
checksum
```

Recommended:

```text
malware scan
```

Jangan execute file upload.

Parser harus diperlakukan sebagai untrusted input boundary.

---

# 105. Secrets

Encrypt:

```text
WhatsApp credentials
provider token
webhook secret
OAuth credential
```

API key:

```text
hash
```

bukan reversible encryption jika tidak perlu mengambil raw key kembali.

---

# 106. Logging

Gunakan structured JSON.

Fields:

```text
timestamp

level

service

request_id

workspace_id

user_id

agent_id

conversation_id

operation

duration_ms

error_code
```

Jangan log:

```text
password

JWT

refresh token

full API key

provider secrets
```

Isi conversation jangan selalu ditulis ke application log.

---

# 107. Audit Log

Catat:

```text
personality.updated

business.updated

knowledge.created
knowledge.updated
knowledge.deleted

api_key.created
api_key.revoked

channel.connected
channel.disconnected

workspace.member_added

workspace.member_removed

agent.suspended

subscription.changed
```

---

# 108. Usage

Track:

```text
messages incoming

messages outgoing

agent invocations

website sessions

public API requests

document storage

knowledge sources

Hermes token usage jika tersedia
```

Usage harus diaggregation.

Jangan:

```sql
COUNT(*) FROM messages
```

setiap dashboard request.

---

# 109. Dashboard Overview

Endpoint:

```text
GET /v1/dashboard
```

Response aggregate:

```json
{
  "data": {
    "agent": {
      "status": "ready"
    },
    "second_brain": {
      "status": "ready",
      "knowledge_sources": 16
    },
    "channel": {
      "status": "connected"
    },
    "conversations": {
      "today": 42,
      "open": 8
    }
  }
}
```

---

# 110. Rate Limiting

Redis backed.

Different buckets:

```text
login

password reset

agent test

agent runtime

public API

file upload

API key
```

Return:

```text
HTTP 429
```

---

# 111. Background Jobs

Types:

```text
workspace.provision

agent.provision

agent.sync

brain.provision

knowledge.ingest

knowledge.update

knowledge.delete

webhook.deliver

workspace.delete
```

---

# 112. Job Retry

Retry transient failures:

```text
Hermes 5xx

storage timeout

temporary filesystem issue

webhook error
```

Gunakan:

```text
exponential backoff
+
jitter
```

Permanent errors:

```text
unsupported document
invalid content
```

langsung failed.

---

# 113. Dead Letter / Failed Jobs

Job yang melewati retry limit harus masuk:

```text
failed_jobs
```

atau queue setara.

Admin harus dapat melihat:

```text
job type
workspace
error
attempts
timestamp
```

---

# 114. Transactional Outbox

Recommended untuk operation:

```text
DB mutation
+
async job
```

Contoh:

```text
Transaction:

Create knowledge_source
Create outbox event

COMMIT
```

Worker kemudian mengirim event ke Redis queue.

Ini mencegah kondisi:

```text
DB sukses
enqueue gagal
```

---

# 115. Go Project Structure

Recommended:

```text
cmd/
├── api/
│   └── main.go
│
└── worker/
    └── main.go

internal/

├── auth/
├── user/
├── workspace/
├── subscription/

├── agent/
├── personality/
├── profile/

├── brain/
│   └── obsidian/

├── knowledge/
│   ├── parser/
│   ├── converter/
│   └── ingestion/

├── channel/
│   └── whatsapp/

├── conversation/

├── hermes/

├── developer/
│   ├── apikey/
│   ├── session/
│   └── webhook/

├── storage/
├── cache/
├── jobs/

├── middleware/
├── observability/
└── database/
```

---

# 116. Application Layers

Pattern:

```text
HTTP Handler
     ↓
Service
     ↓
Repository
```

External:

```text
Service
  ↓
Hermes Adapter

Service
  ↓
Obsidian Adapter

Service
  ↓
Bot Service Client
```

HTTP handler jangan memuat business logic besar.

---

# 117. Core Dashboard Endpoints

```text
GET    /v1/me

GET    /v1/workspace
PATCH  /v1/workspace

GET    /v1/agent
PATCH  /v1/agent

GET    /v1/agent/profile
PATCH  /v1/agent/profile

GET    /v1/agent/personality
PATCH  /v1/agent/personality

POST   /v1/agent/test

GET    /v1/business
PATCH  /v1/business

GET    /v1/knowledge
POST   /v1/knowledge/documents
POST   /v1/knowledge/text
POST   /v1/knowledge/faqs
DELETE /v1/knowledge/:id

GET    /v1/channels

POST   /v1/channels/whatsapp/connect

DELETE /v1/channels/:id

GET    /v1/conversations
GET    /v1/conversations/:id
GET    /v1/conversations/:id/messages

PATCH  /v1/conversations/:id/mode

GET    /v1/subscription
GET    /v1/usage

GET    /v1/api-keys
POST   /v1/api-keys
DELETE /v1/api-keys/:id

GET    /v1/webhooks
POST   /v1/webhooks
PATCH  /v1/webhooks/:id
DELETE /v1/webhooks/:id
```

---

# 118. Dashboard Navigation Target

Frontend dapat menggunakan:

```text
Overview

Agent
├── Profile
├── Personality
└── Business Knowledge

Knowledge
├── Documents
├── FAQ
└── Data

Channels
└── WhatsApp

Conversations

Developer
├── API Keys
├── Webhooks
└── Documentation

Subscription

Settings
```

---

# 119. Agent Health

Internal:

```text
GET /internal/v1/agents/:id/health
```

Response:

```json
{
  "data": {
    "status": "ready",
    "hermes": "healthy",
    "second_brain": "ready",
    "config_synced": true,
    "vault_accessible": true
  }
}
```

---

# 120. Health Checks

Application:

```text
GET /health
```

Liveness:

```text
GET /health/live
```

Readiness:

```text
GET /health/ready
```

Readiness checks:

```text
PostgreSQL

Redis

required storage
```

Jangan selalu membuat readiness tergantung Hermes jika kegagalan Hermes tidak perlu membuat seluruh dashboard offline.

---

# 121. Metrics

Infrastructure:

```text
HTTP latency

HTTP errors

Postgres connections

Redis latency

queue size

failed jobs
```

Agent:

```text
Hermes latency

Hermes error rate

agent response latency

knowledge load latency
```

Knowledge:

```text
ingestion duration

failed parsing

vault write errors

sync duration
```

---

# 122. Tracing

Gunakan OpenTelemetry.

Trace contoh:

```text
Incoming WhatsApp
      ↓
Go API
      ↓
Conversation Repository
      ↓
Second Brain Resolver
      ↓
Hermes
      ↓
Database Save
      ↓
Bot Service Response
```

---

# 123. Performance Target

Dashboard non-agent API:

```text
p95 < 500 ms
```

Cached setting:

```text
p95 < 200 ms
```

Backend overhead untuk agent request di luar Hermes:

Target:

```text
< 300 ms typical
```

Actual LLM/Hermes generation tidak termasuk target tersebut.

---

# 124. Timeout

Explicit timeout wajib untuk:

```text
PostgreSQL

Redis

Hermes

Object Storage

Bot Service

Webhook
```

Semua service methods menerima:

```go
context.Context
```

---

# 125. Graceful Shutdown

Saat API shutdown:

```text
Stop accepting connections

Finish active requests

Close database

Close Redis

Shutdown tracing
```

Worker:

```text
Stop accepting jobs

Finish current safe jobs

Exit
```

---

# 126. Tenant Deletion

Flow:

```text
Request Delete Workspace
        ↓
Workspace = deleting
        ↓
Revoke API keys
        ↓
Disable Agent
        ↓
Disconnect Channels
        ↓
Delete Hermes Agent
        ↓
Archive/Delete Obsidian Vault
        ↓
Delete Original Documents
        ↓
Delete Conversation Data
        ↓
Apply retention policy
        ↓
Workspace = deleted
```

Deletion asynchronous.

---

# 127. Jangan Langsung Hapus pada Subscription Expired

Subscription expired:

```text
agent suspended
```

tetapi:

```text
Hermes Agent retained
Obsidian Vault retained
Documents retained
```

sesuai retention period.

---

# 128. Testing Requirements

## Unit Tests

```text
personality validation

authorization

knowledge filename sanitation

Obsidian path resolver

document conversion

entitlement

agent prompt builder

error mapping
```

## Integration Tests

```text
PostgreSQL

Redis

Object Storage

Obsidian adapter

Hermes mocked adapter
```

## E2E

```text
Signup

Workspace provision

Configure personality

Upload knowledge

Connect WhatsApp

Send test message

Receive response

View conversation

Create API key

Create website session

Website agent response
```

---

# 129. Mandatory Tenant Isolation Test

Setup:

```text
Workspace A
Workspace B
```

User A attempts:

```text
read Agent B

read Vault metadata B

read Knowledge B

read Conversation B

invoke Agent B

delete API key B
```

Semua harus gagal.

---

# 130. Mandatory Second Brain Isolation Test

Vault A:

```text
Refund allowed within 7 days.
```

Vault B:

```text
Refunds are not accepted.
```

Prompt Agent A:

```text
Berapa lama batas refund?
```

Agent A hanya boleh menggunakan information A.

Prompt Agent B tidak boleh melihat knowledge A.

Test ini WAJIB sebelum production.

---

# 131. Obsidian Path Security Test

Test input:

```text
../../
../workspace_B
%2e%2e/
symbolic link
absolute path
```

Backend tidak boleh dapat escape dari vault tenant.

---

# 132. End-to-End Definition of Done

Backend MVP selesai jika flow berikut dapat berjalan.

## Scenario

Customer registrasi.

↓

Workspace dibuat.

↓

Backend otomatis membuat:

```text
Hermes Agent
+
Obsidian Vault
```

↓

Vault structure tersedia.

↓

User masuk dashboard.

↓

Mengatur:

```text
Nama agent:
Naya

Tone:
Ramah

Style:
Santai profesional

Emoji:
Normal
```

↓

Backend menyimpan config.

↓

Backend generate:

```text
bot/personality.md
```

↓

Hermes menerima update.

↓

User mengisi profil perusahaan.

↓

Backend generate:

```text
business/company-profile.md
```

↓

User upload:

```text
catalog.pdf

refund-policy.pdf
```

↓

Original disimpan di object storage.

↓

Worker extract.

↓

Worker convert menjadi Markdown.

↓

Markdown masuk ke:

```text
products/

policies/
```

↓

Hermes dapat menggunakan knowledge tersebut.

↓

User menghubungkan nomor WhatsApp.

↓

Bot Service status:

```text
CONNECTED
```

↓

Customer WhatsApp mengirim:

```text
Kak, barang ini bisa dikembalikan?
```

↓

Bot Service mengirim message ke Go Backend.

↓

Go Backend menentukan:

```text
workspace

agent

personality

vault

conversation
```

↓

Hermes menjalankan Agent tenant.

↓

Agent menggunakan refund policy dari Obsidian tenant.

↓

Response tersimpan ke database.

↓

Response dikembalikan ke Bot Service.

↓

Customer menerima jawaban.

↓

Conversation muncul di dashboard.

↓

Owner membuat:

```text
ChatSolv API Key
```

↓

Owner membuat website agent session.

↓

Visitor website mengirim pertanyaan.

↓

Request menggunakan:

```text
Hermes Agent yang sama

Personality yang sama

Obsidian Second Brain yang sama
```

↓

Agent menjawab.

Jika keseluruhan flow ini bekerja dengan isolation dan reliability yang benar:

**Backend MVP ChatSolv dinyatakan berhasil.**

---

# 133. MVP Scope

## Wajib MVP

* Authentication
* Workspace
* Basic RBAC
* Subscription/trial
* Default Agent
* Dedicated Hermes Agent per tenant
* Dedicated Obsidian Vault per tenant
* Vault provisioning
* Personality settings
* Bot profile
* Business profile
* Policies
* PDF ingestion
* DOCX ingestion
* TXT ingestion
* CSV ingestion
* Markdown conversion
* Obsidian write/sync
* WhatsApp channel metadata
* Internal WhatsApp Service API
* Conversation storage
* Agent runtime
* Agent test playground
* Human handoff state
* Redis cache
* Async workers
* API keys
* Website session API
* Basic Public API
* Audit logs
* Usage tracking
* Structured logging
* Tenant isolation

---

# 134. Phase 1.5

* XLSX advanced ingestion
* JSON structured importer
* knowledge version UI
* webhook
* SSE streaming
* advanced human inbox
* advanced usage analytics
* team management
* multiple knowledge categories
* custom escalation rules

---

# 135. Future Scope

* Multiple Hermes agents per workspace
* Multiple Obsidian Second Brains
* Multiple WhatsApp numbers
* Website crawler
* Google Drive
* Notion
* External database connector
* CRM connectors
* Instagram
* Telegram
* Messenger
* Email
* React SDK
* JavaScript SDK
* ChatSolv embeddable widget
* custom tools / functions
* order lookup
* CRM lookup
* appointment lookup
* payment status tool
* agent workflows

---

# 136. Future Database Integration

Ketika nanti customer ingin memberikan akses database perusahaan, jangan menyalin seluruh production database customer ke Obsidian.

Buat konsep:

```text
Knowledge Connector
```

dan:

```text
Runtime Tool Connector
```

Bedakan dua kebutuhan.

## Knowledge Connector

Cocok untuk:

```text
katalog

product description

FAQ

static company data
```

Data dapat disinkronkan ke Obsidian.

## Runtime Tool

Cocok untuk:

```text
cek pesanan

cek stok realtime

cek booking

cek status pembayaran
```

Hermes memanggil tool/API saat dibutuhkan.

Data realtime jangan disimpan sebagai markdown yang cepat basi.

---

# 137. Future Agent Tool Architecture

Concept:

```text
Customer Message

      ↓

Hermes Agent

      ↓

Need realtime information?

      ↓ YES

Tool Registry

      ↓

Tenant Tool

      ↓

Customer API

      ↓

Result

      ↓

Hermes

      ↓

Customer Response
```

Ini akan memungkinkan ChatSolv berkembang dari:

```text
Knowledge-based CS
```

menjadi:

```text
Action-capable CS Agent
```

tanpa merusak arsitektur awal.

---

# 138. Critical Engineering Rules

### RULE 01

Setiap tenant memiliki Hermes Agent sendiri.

### RULE 02

Setiap tenant memiliki Obsidian Vault sendiri.

### RULE 03

Obsidian merupakan Second Brain utama.

### RULE 04

Conversation history bukan Second Brain.

### RULE 05

PostgreSQL adalah source of truth transactional.

### RULE 06

Redis bukan database utama.

### RULE 07

Hermes resource ID tidak diekspos ke frontend.

### RULE 08

Vault path tidak diekspos ke frontend.

### RULE 09

Frontend tidak pernah mengakses vault filesystem.

### RULE 10

Semua knowledge mutation melalui backend.

### RULE 11

Document ingestion asynchronous.

### RULE 12

WhatsApp bot merupakan service terpisah.

### RULE 13

Backend Go menentukan tenant dan agent untuk setiap request.

### RULE 14

Incoming channel events harus idempotent.

### RULE 15

Setiap tenant-owned database row memiliki workspace boundary.

### RULE 16

Knowledge dianggap reference, bukan system instruction.

### RULE 17

Secret Public API key tidak boleh digunakan di browser.

### RULE 18

Live chat browser menggunakan ephemeral session token.

### RULE 19

WhatsApp dan Website menggunakan Hermes Agent tenant yang sama.

### RULE 20

WhatsApp dan Website menggunakan Obsidian Vault tenant yang sama.

### RULE 21

Realtime operational data jangan dipaksakan masuk Obsidian.

### RULE 22

Integrasi database realtime nantinya menggunakan Tool Connector.

### RULE 23

Backend harus dirancang agar multiple agent dapat ditambahkan nanti tanpa refactor fundamental.

---

# 139. Prioritas Implementasi Backend Developer

Urutan pengerjaan recommended:

## Sprint 1 — Foundation

```text
Go Fiber bootstrap

Configuration

PostgreSQL

Redis

Migration

Authentication

Workspace

RBAC

Error handling

Logging
```

## Sprint 2 — Tenant Agent Infrastructure

```text
Agent entity

Hermes adapter

Obsidian adapter

Second Brain entity

Agent provisioning

Vault provisioning

Provisioning worker
```

## Sprint 3 — Agent Configuration

```text
Personality

Agent Profile

Business Profile

Policies

Config versioning

Hermes sync
```

## Sprint 4 — Knowledge

```text
File upload

Object storage

Parser

Markdown converter

Vault writer

Knowledge jobs

Knowledge management API
```

## Sprint 5 — Agent Runtime

```text
Conversation

Messages

Second Brain resolver

Hermes runtime

Agent Test

Conversation locking

Human handoff trigger (uncertainty/escalation)

Usage
```

## Sprint 6 — WhatsApp Integration & Admin Relay

```text
Channel

Bot Service internal auth

Connect flow

Incoming messages

Response

Channel status

WhatsApp Admin Relay & Anonymous Proxy (#ACC, #DONE, # prefix relay)
```

## Sprint 7 — Developer Platform

```text
API keys

Agent sessions

Public API

Ephemeral token

SSE

Webhooks
```

## Sprint 8 — Hardening

```text
Isolation tests

Security

Rate limits

Tracing

Metrics

Audit log

Retries

Dead jobs

Load tests
```

---

# 140. Final Backend Mental Model

Backend developer harus memahami ChatSolv seperti ini:

```text
                 CHATSOLV

                    │
              Go Fiber API
                    │
       ┌────────────┼────────────┐
       │            │            │
       ▼            ▼            ▼
 PostgreSQL       Redis       Storage
 Transaction     Runtime      Originals
       │
       ▼
 Workspace Resolver
       │
       ▼
 Tenant Agent
       │
       ├───────────────┐
       │               │
       ▼               ▼
 Hermes Agent      Obsidian Vault
 Personality       Second Brain
 Reasoning         Business Knowledge
       │               │
       └───────┬───────┘
               │
               ▼
          Agent Runtime
               │
       ┌───────┼───────────┐
       ▼       ▼           ▼
   WhatsApp   Web Chat    Public API
```

**ChatSolv Backend bukan sekadar REST API untuk dashboard.**

Backend adalah orchestration layer yang menghubungkan:

```text
Tenant

Personality

Hermes

Obsidian

Knowledge

Conversation

Channels

Developer API
```

menjadi satu agent customer service yang konsisten untuk setiap bisnis.
