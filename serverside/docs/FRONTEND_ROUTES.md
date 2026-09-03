# ChatSolv Frontend Route Handoff

Dokumen ini adalah daftar route untuk integrasi frontend ChatSolv.

Last updated: 25 August 2026

## Implementation Progress

| Area | Active | Planned | Progress | Notes |
|---|---:|---:|---|---|
| System health | 4 | 0 | COMPLETE | `/health`, `/ready`, `/health/live`, `/health/ready` aktif |
| Authentication | 5 | 0 | COMPLETE | Register, login, refresh, forgot/reset password aktif |
| Current user and dashboard | 2 | 0 | COMPLETE | `/me` dan tenant-scoped dashboard aggregate aktif |
| Workspace and subscription | 6 | 2 | PARTIAL | Canonical workspace aktif; subscription alias dan usage belum aktif |
| Agent configuration | 11 | 0 | COMPLETE | Canonical agent, profile, personality, dan isolated test playground aktif |
| Business settings | 6 | 0 | COMPLETE | Canonical business aliases dan workspace-scoped settings aktif |
| Knowledge lifecycle | 8 | 0 | COMPLETE | Create, list, detail, update, delete, dan retry aktif |
| Channels | 3 | 0 | COMPLETE | List, connect (pairing QR), dan delete/disconnect aktif dengan WhatsApp service di `Projects/whatsapp` |
| Conversations and handoff | 4 | 0 | COMPLETE | Tenant list/detail/messages dan manager handoff/resume aktif |
| API keys | 3 | 0 | COMPLETE | Tenant-managed list/create/revoke aktif; raw secret hanya sekali |
| Webhooks | 4 | 0 | COMPLETE | Tenant-managed CRUD aktif; signing secret encrypted at rest |
| Public website agent API | 3 | 0 | COMPLETE | API-key session, client-token message, dan SSE stream aktif |
| Internal service API | 5 | 0 | COMPLETE | HMAC SHA-256; internal agent respond/health dan WhatsApp bot callbacks aktif |

Current totals:

- `63 READY` routes aktif dan siap digunakan (termasuk 5 internal service routes).
- `0 BLOCKED` routes (WhatsApp service sudah diimplementasikan di `Projects/whatsapp`).
- `2 PLANNED` subscription/usage routes masih perlu diimplementasikan.

Latest completed batch:

- `GET|PATCH /api/v1/business?workspace_id=:workspaceID`
- `GET /api/v1/channels?workspace_id=:workspaceID`
- `POST /api/v1/channels/whatsapp/connect?workspace_id=:workspaceID`
- `DELETE /api/v1/channels/:id`
- Channel mutation memverifikasi owner/admin sebelum memanggil WhatsApp Bot Service.
- Kedua route mutation tersebut belum operasional karena WhatsApp Bot Service belum dibuat. Integration map: [`docs/WHATSAPP_INTEGRATION.md`](WHATSAPP_INTEGRATION.md).

Previous completed batch — canonical agent:

- `GET /api/v1/agent?workspace_id=:workspaceID`
- `PATCH /api/v1/agent?workspace_id=:workspaceID`
- `GET|PATCH /api/v1/agent/profile?workspace_id=:workspaceID`
- `GET|PATCH /api/v1/agent/personality?workspace_id=:workspaceID`
- `POST /api/v1/agent/test?workspace_id=:workspaceID`
- Agent test memakai conversation `environment=test` dan tidak mencatat production usage.

Earlier completed batch — canonical workspace:

- `GET /api/v1/workspace?workspace_id=:workspaceID`
- `PATCH /api/v1/workspace?workspace_id=:workspaceID`
- Canonical workspace routes memakai service, authorization, dan tenant boundary yang sama dengan plural routes.

Earlier completed batch — current user/dashboard:

- `GET /api/v1/me`
- `GET /api/v1/dashboard?workspace_id=:workspaceID`
- Dashboard hanya mengagregasi workspace yang dapat diakses authenticated user.
- `/me` mengembalikan identity dan seluruh workspace membership milik user.

Earlier completed batch — knowledge lifecycle:

- `PATCH /api/v1/knowledge/:id`
- `DELETE /api/v1/knowledge/:id`
- `POST /api/v1/knowledge/:id/retry`
- Create/update/delete/retry knowledge dan durable outbox event dilakukan atomik dalam transaksi PostgreSQL.
- Tenant boundary dan duplicate checksum mapping telah diberi test.

Base URL development:

```text
http://localhost:3000
```

## Status

- `ACTIVE`: sudah terdaftar di `cmd/server/main.go` dan dapat diintegrasikan sekarang.
- `ACTIVE — BLOCKED BY BOT SERVICE`: sudah terdaftar dan backend-side contract tersedia, tetapi request eksternal akan gagal sampai service WhatsApp dibuat.
- `PLANNED`: contract target berdasarkan `docs/PRD.md`, tetapi backend belum mendaftarkan route tersebut.
- Jangan memanggil route `PLANNED` dari production frontend sampai statusnya berubah menjadi `ACTIVE`.

## Authentication

Route bertanda `JWT` membutuhkan:

```http
Authorization: Bearer <access_token>
```

Mutation JSON membutuhkan:

```http
Content-Type: application/json
```

Response API normal menggunakan envelope:

```json
{
  "data": {},
  "meta": {},
  "request_id": "..."
}
```

Response error:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Safe error message"
  },
  "request_id": "..."
}
```

---

# 1. Active Routes

Berikut route yang benar-benar aktif saat ini.

## System

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/health` | Public | ACTIVE |
| `GET` | `/ready` | Public | ACTIVE |

## Authentication

| Method | Path | Auth | Status |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | Public | ACTIVE |
| `POST` | `/api/v1/auth/login` | Public | ACTIVE |
| `POST` | `/api/v1/auth/refresh` | Public | ACTIVE |
| `POST` | `/api/v1/auth/forgot-password` | Public | ACTIVE |
| `POST` | `/api/v1/auth/reset-password` | Public | ACTIVE |

## Current user and dashboard

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/me` | JWT | ACTIVE |
| `GET` | `/api/v1/dashboard?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |

`GET /api/v1/me` response data:

```json
{
  "user": {
    "id": "user-uuid",
    "name": "Ayu",
    "email": "ayu@example.com",
    "created_at": "2026-08-25T12:00:00Z"
  },
  "workspaces": [
    {
      "workspace_id": "workspace-uuid",
      "name": "Toko A",
      "slug": "toko-a",
      "status": "active",
      "timezone": "Asia/Jakarta",
      "role": "owner"
    }
  ]
}
```

`workspace_id` wajib dikirim ke dashboard agar tenant context eksplisit. Response data mengikuti aggregate PRD:

```json
{
  "workspace_id": "workspace-uuid",
  "agent": {"status": "ready"},
  "second_brain": {"status": "ready", "knowledge_sources": 16},
  "channel": {"status": "connected"},
  "conversations": {"today": 42, "open": 8}
}
```

Jika user bukan member workspace tersebut, backend mengembalikan `404 WORKSPACE_NOT_FOUND` agar keberadaan tenant lain tidak dibocorkan.

## Workspace and subscription

| Method | Path | Auth | Status |
|---|---|---|---|
| `POST` | `/api/v1/workspaces` | JWT | ACTIVE |
| `GET` | `/api/v1/workspaces/:workspaceID` | JWT | ACTIVE |
| `PATCH` | `/api/v1/workspaces/:workspaceID` | JWT, owner/admin for mutation | ACTIVE |
| `GET` | `/api/v1/workspaces/:workspaceID/subscription` | JWT | ACTIVE |
| `GET` | `/api/v1/workspace?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/workspace?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |

Canonical workspace routes mewajibkan `workspace_id` agar tenant context eksplisit:

```bash
curl "$BASE_URL/api/v1/workspace?workspace_id=$WORKSPACE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -X PATCH "$BASE_URL/api/v1/workspace?workspace_id=$WORKSPACE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Toko Baru","timezone":"Asia/Jakarta"}'
```

## Agent configuration

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/agents/:agentID/personality` | JWT | ACTIVE |
| `PATCH` | `/api/v1/agents/:agentID/personality` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/agents/:agentID/profile` | JWT | ACTIVE |
| `PATCH` | `/api/v1/agents/:agentID/profile` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/agent?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/agent?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/agent/profile?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/agent/profile?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/agent/personality?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/agent/personality?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `POST` | `/api/v1/agent/test?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |

Canonical routes me-resolve default agent milik workspace. `workspace_id` wajib agar tenant context tidak ambigu. Test payload:

```json
{"message":"Kak, bisa COD?"}
```

Agent test membuat conversation terisolasi dengan `environment=test`; conversation ini tidak masuk production customer metrics.

## Business settings

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/settings/workspaces/:workspaceID/business` | JWT | ACTIVE |
| `PATCH` | `/api/v1/settings/workspaces/:workspaceID/business` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/settings/workspaces/:workspaceID/policies` | JWT | ACTIVE |
| `PATCH` | `/api/v1/settings/workspaces/:workspaceID/policies` | JWT, owner/admin | ACTIVE |
| `GET` | `/api/v1/business?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/business?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |

Canonical business route memakai `workspace_id` eksplisit dan contract body yang sama dengan workspace-scoped business settings.

## Channels

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/channels?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `POST` | `/api/v1/channels/whatsapp/connect?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `DELETE` | `/api/v1/channels/:id` | JWT, owner/admin | ACTIVE |

WhatsApp connect body dan response pairing:

```json
{"display_name":"WhatsApp Utama"}
```

Saat Bot Service tersedia, connect akan mengembalikan `202 Accepted` dengan `channel` dan `pairing.qr`. Raw internal `service_instance_id` tidak dikirim ke frontend. DELETE memverifikasi tenant dan role sebelum meminta Bot Service memutus session.

Untuk sekarang frontend boleh memakai GET list, tetapi tombol connect/disconnect harus ditandai unavailable/coming soon. Daftar file dan contract yang perlu disambungkan saat Bot Service dibuat ada di [`docs/WHATSAPP_INTEGRATION.md`](WHATSAPP_INTEGRATION.md).

## Knowledge

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/knowledge?workspace_id=:workspaceID&limit=50` | JWT | ACTIVE |
| `GET` | `/api/v1/knowledge/:id` | JWT | ACTIVE |
| `PATCH` | `/api/v1/knowledge/:id` | JWT, writable membership | ACTIVE |
| `DELETE` | `/api/v1/knowledge/:id` | JWT, writable membership | ACTIVE |
| `POST` | `/api/v1/knowledge/:id/retry` | JWT, writable membership | ACTIVE |
| `POST` | `/api/v1/knowledge/documents` | JWT, multipart | ACTIVE |
| `POST` | `/api/v1/knowledge/text` | JWT | ACTIVE |
| `POST` | `/api/v1/knowledge/faqs` | JWT | ACTIVE |

Total route terdaftar: 65 (`63 READY`, `2 PLANNED`).

---

# 2. Planned Frontend Routes from PRD

Semua path `/v1/...` di PRD dinormalisasi menjadi prefix public/dashboard `/api/v1/...` sesuai aturan API versioning PRD.

Progress bagian ini: current user, dashboard, canonical workspace/agent/business, channels, dan knowledge lifecycle selesai dan sudah dipindahkan ke `Active Routes`.

## Conversations and handoff

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/conversations?workspace_id=:workspaceID` | JWT, workspace member | ACTIVE |
| `GET` | `/api/v1/conversations/:id` | JWT, workspace member | ACTIVE |
| `GET` | `/api/v1/conversations/:id/messages` | JWT, workspace member | ACTIVE |
| `PATCH` | `/api/v1/conversations/:id/mode` | JWT, owner/admin | ACTIVE |

List mendukung filter `status`, `mode`, dan `limit` (maksimal 100). Messages mendukung cursor RFC3339/RFC3339Nano dan `limit` maksimal 100. Detail/messages memakai not-found semantics untuk cross-tenant access. Mode hanya menerima `agent` atau `human`; runtime handoff juga mempersist mode `human`.

## Subscription and usage

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/subscription` | JWT | PLANNED |
| `GET` | `/api/v1/usage` | JWT | PLANNED |

## API keys

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/api-keys?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `POST` | `/api/v1/api-keys?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `DELETE` | `/api/v1/api-keys/:id` | JWT, owner/admin | ACTIVE |

Raw API key hanya ditampilkan sekali pada response create. Create body: `{"name":"Production","scopes":["agent:invoke"]}`. Database hanya menyimpan prefix, SHA-256 hash, dan last four. Frontend tidak boleh menyimpan atau mencetak raw key ke log.

## Webhooks

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/api/v1/webhooks?workspace_id=:workspaceID` | JWT, owner/admin | ACTIVE |
| `POST` | `/api/v1/webhooks?workspace_id=:workspaceID` | JWT, owner/admin, webhook entitlement | ACTIVE |
| `PATCH` | `/api/v1/webhooks/:id` | JWT, owner/admin | ACTIVE |
| `DELETE` | `/api/v1/webhooks/:id` | JWT, owner/admin | ACTIVE |

Create body: `{"url":"https://example.com/webhooks/chatsolv","events":["message.created"]}`. URL wajib HTTPS. Signing secret `whsec_*` hanya ditampilkan sekali dan disimpan encrypted menggunakan AES-GCM. Delivery worker/retry belum termasuk CRUD batch ini.

---

# 3. Public Website Agent API

Route ini dipakai oleh customer backend atau website visitor, bukan dashboard utama.

| Method | Path | Auth | Status |
|---|---|---|---|
| `POST` | `/api/v1/agent-sessions` | Secret API key from customer backend | ACTIVE |
| `POST` | `/api/v1/agent-sessions/:id/messages` | Ephemeral client token | ACTIVE |
| `POST` | `/api/v1/agent-sessions/:id/messages/stream` | Ephemeral client token | ACTIVE |

Secret `cs_live_*`/`cs_test_*` dilarang ditanam di browser. Browser hanya menerima ephemeral client token dari customer backend. Create body: `{"external_user_id":"visitor_19273","metadata":{"page":"/pricing"}}`. Client token berlaku 3600 detik dan hanya hash-nya yang disimpan. SSE mengirim `message.start` dan `message.completed`; runtime saat ini belum mengirim token delta native.

---

# 4. Health Aliases

Aliases berikut aktif dan menggunakan implementation yang sama dengan `/health` dan `/ready`:

| Method | Path | Auth | Status |
|---|---|---|---|
| `GET` | `/health/live` | Public | ACTIVE |
| `GET` | `/health/ready` | Public | ACTIVE |

---

# 5. Internal Routes — Not for Frontend

Route berikut hanya untuk komunikasi service-to-service. Jangan dipanggil dari browser atau dashboard frontend.

| Method | Path | Authentication | Status |
|---|---|---|---|
| `POST` | `/internal/v1/channels/events` | Internal HMAC | ACTIVE |
| `POST` | `/internal/v1/channels/status` | Internal HMAC | ACTIVE |
| `POST` | `/internal/v1/messages/incoming` | Internal HMAC | ACTIVE |
| `POST` | `/internal/v1/agents/:agentID/respond` | Internal HMAC | ACTIVE |
| `GET` | `/internal/v1/agents/:agentID/health` | Internal HMAC | ACTIVE |

Semua route memakai timestamped HMAC SHA-256 dengan replay window lima menit; user JWT tidak berlaku. Internal credentials, signatures, provider agent IDs, dan vault paths tidak boleh tersedia di frontend. WhatsApp service di `Projects/whatsapp` sudah mengimplementasikan callback ini.

---

# 6. Frontend Integration Recommendation

1. Integrasikan hanya route berstatus `ACTIVE`.
2. Simpan path dalam satu typed API client agar migrasi dari active plural routes ke canonical PRD routes tidak tersebar di komponen UI.
3. Perlakukan `202 Accepted` sebagai asynchronous operation; tampilkan status `provisioning`, `queued`, `processing`, atau `syncing` sesuai response.
4. Jangan mengambil `workspace_id`, `agent_id`, provider ID, atau vault path dari input bebas user.
5. Selalu tampilkan error berdasarkan `error.code`, bukan membandingkan teks `error.message`.
6. Kirim dan simpan `X-Request-ID` untuk troubleshooting, tetapi jangan log access token, refresh token, reset token, API key, atau request body sensitif.
