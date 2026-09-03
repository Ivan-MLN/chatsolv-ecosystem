# ChatSolv Backend API Reference & Usage Guide

Dokumen ini adalah panduan lengkap cara menggunakan seluruh route ChatSolv Backend API, mencakup Method, Path, Query Parameters, Headers, Auth Requirement, Request Body, Response Schema (Sukses & Error), status code lengkap (200, 201, 202, 400, 401, 403, 404, 409, 500, 503), dan contoh `curl`.

---

## Standar Header & Format Response

### 1. Header Otentikasi
- **JWT Auth (Dashboard & User APIs):** `Authorization: Bearer <access_token>`
- **API Key Auth (Secret Key):** `Authorization: Bearer cs_live_<secret_key>`
- **Public Agent Sessions:** `Authorization: Bearer cs_pub_<client_token>`
- **Internal Microservices (HMAC SHA-256):**
  - `X-ChatSolv-Timestamp: <RFC3339 timestamp>`
  - `X-ChatSolv-Signature: <hex HMAC-SHA256(secret, timestamp + "." + body)>`
- **Content-Type:** `application/json` (kecuali upload multipart dokumen)

### 2. Envelope Response Sukses
```json
{
  "success": true,
  "message": "Deskripsi operasi sukses",
  "data": { ... }
}
```

### 3. Envelope Response Error
```json
{
  "error": {
    "code": "VALIDATION_ERROR | UNAUTHORIZED | FORBIDDEN | NOT_FOUND | CONFLICT | INTERNAL_ERROR | SERVICE_UNAVAILABLE",
    "message": "Pesan deskriptif error"
  },
  "request_id": "876a6a7e-10ff-4ca0-adb1-064cab84f3f1"
}
```

---

## 1. System Health & Readiness

### GET `/health` & `/health/live`
Pemeriksaan liveness cepat proses backend tanpa membebani database/cache.

- **Auth:** Public
- **Contoh Request:**
```bash
curl -s http://127.0.0.1:3000/health
curl -s http://127.0.0.1:3000/health/live
```
- **Response `200 OK`:**
```json
{
  "status": "ok"
}
```

---

### GET `/ready` & `/health/ready`
Pemeriksaan readiness menyeluruh ke PostgreSQL dan Redis dengan bounded timeout 1 detik.

- **Auth:** Public
- **Contoh Request:**
```bash
curl -s http://127.0.0.1:3000/ready
curl -s http://127.0.0.1:3000/health/ready
```
- **Response `200 OK`:**
```json
{
  "status": "ready"
}
```
- **Response `503 Service Unavailable` (Jika PostgreSQL / Redis mati):**
```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Database or cache unavailable"
  },
  "request_id": "a910bf21-..."
}
```

---

## 2. Authentication

### POST `/api/v1/auth/register`
Mendaftarkan akun pengguna baru.

- **Auth:** Public
- **Request Body:**
```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "Password123!"
}
```
- **Contoh Request:**
```bash
curl -s -X POST http://127.0.0.1:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "password": "Password123!"
  }'
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "Registration successful. Please verify your email.",
  "data": {
    "id": "c1f7a4e8-8b9a-4e2a-8c9a-9e2a8c9a9e2a",
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "created_at": "2026-08-25T12:00:00Z"
  }
}
```
- **Response `400 Bad Request` (Validasi Gagal):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid registration input (password minimum 8 chars with uppercase, lowercase, number, symbol)"
  },
  "request_id": "..."
}
```
- **Response `409 Conflict` (Email Sudah Terdaftar):**
```json
{
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email is already registered"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/auth/login`
Autentikasi user untuk mendapatkan JWT Access Token dan Refresh Token.

- **Auth:** Public
- **Request Body:**
```json
{
  "email": "budi@example.com",
  "password": "Password123!"
}
```
- **Contoh Request:**
```bash
curl -s -X POST http://127.0.0.1:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "budi@example.com",
    "password": "Password123!"
  }'
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "def50200...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "c1f7a4e8-8b9a-4e2a-8c9a-9e2a8c9a9e2a",
      "name": "Budi Santoso",
      "email": "budi@example.com"
    }
  }
}
```
- **Response `401 Unauthorized` (Email/Password Salah):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/auth/refresh`
Memperbarui access token yang telah kedaluwarsa menggunakan refresh token.

- **Auth:** Public
- **Request Body:**
```json
{
  "refresh_token": "def50200..."
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "def50200...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```
- **Response `401 Unauthorized` (Refresh Token Tidak Valid / Expired):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired refresh token"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/auth/forgot-password`
Mengirimkan instruksi reset password ke email user.

- **Auth:** Public
- **Request Body:**
```json
{
  "email": "budi@example.com"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "If the email is registered, password reset instructions have been sent."
}
```
- **Response `400 Bad Request` (Format Email Salah):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format"
  },
  "request_id": "..."
}
```

---

## 3. Current User & Dashboard

### GET `/api/v1/me`
Mengambil detail pengguna yang terotentikasi dan daftar workspace yang diikuti.

- **Auth:** JWT Bearer
- **Contoh Request:**
```bash
curl -s http://127.0.0.1:3000/api/v1/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Current user retrieved",
  "data": {
    "user": {
      "id": "c1f7a4e8-8b9a-4e2a-8c9a-9e2a8c9a9e2a",
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "created_at": "2026-08-25T12:00:00Z"
    },
    "workspaces": [
      {
        "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
        "name": "ChatSolv Indonesia",
        "slug": "chatsolv-id",
        "status": "active",
        "timezone": "Asia/Jakarta",
        "role": "owner"
      }
    ]
  }
}
```
- **Response `401 Unauthorized` (Token Hilang / Salah):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid authorization token"
  },
  "request_id": "..."
}
```

---

### GET `/api/v1/dashboard?workspace_id=:id`
Mengambil overview status operasional workspace (Agent, Second Brain, Channel, dan Conversation).

- **Auth:** JWT Bearer
- **Query Params:** `workspace_id` (wajib)
- **Contoh Request:**
```bash
curl -s "http://127.0.0.1:3000/api/v1/dashboard?workspace_id=$WORKSPACE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Dashboard retrieved",
  "data": {
    "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
    "agent": {
      "status": "ready"
    },
    "second_brain": {
      "status": "ready",
      "knowledge_sources": 5
    },
    "channel": {
      "status": "connected"
    },
    "conversations": {
      "today": 12,
      "open": 3
    }
  }
}
```
- **Response `400 Bad Request` (`workspace_id` tidak disertakan):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "workspace_id is required"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Workspace tidak ada / bukan member):**
```json
{
  "error": {
    "code": "WORKSPACE_NOT_FOUND",
    "message": "Dashboard workspace not found"
  },
  "request_id": "..."
}
```

---

## 4. Workspace Management

### POST `/api/v1/workspaces`
Membuat workspace baru dengan langganan awal berstatus `inactive`, agent default, second brain, dan antrean provisioning. Workspace developer mendapat akses komersial tanpa langganan, tetapi tetap tunduk pada batas tenant.

- **Auth:** JWT Bearer
- **Request Body:**
```json
{
  "name": "ChatSolv Store",
  "slug": "chatsolv-store",
  "timezone": "Asia/Jakarta"
}
```
- **Response `202 Accepted`:**
```json
{
  "success": true,
  "message": "Workspace provisioning started",
  "data": {
    "workspace": {
      "id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
      "name": "ChatSolv Store",
      "slug": "chatsolv-store",
      "status": "provisioning",
      "timezone": "Asia/Jakarta",
      "created_at": "2026-08-25T12:00:00Z"
    },
    "membership": {
      "id": "a910bf21-...",
      "role": "owner"
    }
  }
}
```
- **Response `400 Bad Request` (Format slug / timezone salah):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid workspace input"
  },
  "request_id": "..."
}
```
- **Response `409 Conflict` (Slug sudah digunakan):**
```json
{
  "error": {
    "code": "SLUG_EXISTS",
    "message": "Workspace slug already exists"
  },
  "request_id": "..."
}
```

---

### GET `/api/v1/workspace?workspace_id=:id`
Mengambil informasi detail workspace (Canonical).

- **Auth:** JWT Bearer
- **Query Params:** `workspace_id` (wajib)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Workspace retrieved",
  "data": {
    "workspace": {
      "id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
      "name": "ChatSolv Store",
      "slug": "chatsolv-store",
      "status": "active",
      "timezone": "Asia/Jakarta"
    },
    "membership": {
      "role": "owner"
    }
  }
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "WORKSPACE_NOT_FOUND",
    "message": "Workspace not found"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/workspace?workspace_id=:id`
Mengubah nama atau timezone workspace (Canonical).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "name": "ChatSolv Store Official",
  "timezone": "Asia/Jakarta"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Workspace updated",
  "data": {
    "id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
    "name": "ChatSolv Store Official",
    "slug": "chatsolv-store",
    "timezone": "Asia/Jakarta"
  }
}
```
- **Response `403 Forbidden` (User adalah role `member`):**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Workspace action forbidden"
  },
  "request_id": "..."
}
```

---

### GET `/api/v1/workspaces/:workspaceID/subscription`
Mengambil status langganan, mode akses (Developer Mode vs Normal Subscription), batasan kuota, dan ringkasan penggunaan workspace.

- **Auth:** JWT Bearer
- **Response `200 OK` (Developer Mode Account):**
```json
{
  "success": true,
  "message": "Subscription retrieved",
  "data": {
    "subscription": {
      "id": "sub_12345",
      "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
      "status": "active",
      "plan_id": "chatsolv_starter",
      "billing_cycle": "monthly",
      "currency": "IDR",
      "amount": 459000
    },
    "entitlement": {
      "max_agents": 1,
      "max_channels": 1,
      "max_storage_mb": 2048,
      "max_documents": 200,
      "monthly_messages": 20000,
      "public_api": true,
      "webhooks": true
    },
    "access": {
      "mode": "developer",
      "subscription_required": false,
      "unlimited": true
    },
    "limits": {
      "agents": "unlimited",
      "channels": "unlimited",
      "messages_monthly": "unlimited",
      "knowledge_documents": "unlimited",
      "knowledge_storage_bytes": "unlimited",
      "public_api": true,
      "webhooks": true
    }
  }
}
```

- **Response `200 OK` (Standard Paid Customer):**
```json
{
  "success": true,
  "message": "Subscription retrieved",
  "data": {
    "subscription": {
      "id": "sub_12345",
      "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
      "status": "active",
      "plan_id": "chatsolv_starter",
      "billing_cycle": "monthly",
      "currency": "IDR",
      "amount": 459000
    },
    "access": {
      "mode": "subscription",
      "subscription_required": true,
      "unlimited": false
    },
    "limits": {
      "agents": 1,
      "channels": 1,
      "messages_monthly": 20000,
      "knowledge_documents": 200,
      "knowledge_storage_bytes": 2147483648,
      "public_api": true,
      "webhooks": true
    }
  }
}
```

---

### POST `/api/v1/workspaces/:workspaceID/checkout`
Membuat transaksi checkout pembayaran langganan ChatSolv Starter (Rp459.000 / bulan).

- **Auth:** JWT Bearer
- **Request Body:**
```json
{
  "plan_id": "chatsolv_starter",
  "provider": "xendit"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Checkout created",
  "data": {
    "payment_reference": "CS-A1B2C3D4E5F6",
    "plan_id": "chatsolv_starter",
    "amount": 459000,
    "currency": "IDR",
    "status": "pending",
    "provider": "xendit",
    "checkout_url": "https://checkout.chatsolv.com/pay/CS-A1B2C3D4E5F6"
  }
}
```

---

### POST `/api/v1/payments/webhook`
Webhook penerimaan konfirmasi pembayaran server-to-server yang idempoten dari payment gateway.

- **Auth:** HMAC server-to-server. Wajib mengirim `X-ChatSolv-Timestamp` dan `X-ChatSolv-Signature` dengan skema HMAC yang dijelaskan pada bagian header. Callback tanpa tanda tangan yang valid ditolak.
- **Request Body:**
```json
{
  "payment_reference": "CS-A1B2C3D4E5F6",
  "provider_transaction_id": "xendit_txn_987654",
  "status": "paid",
  "metadata": {}
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Payment processed successfully",
  "data": {
    "payment_reference": "CS-A1B2C3D4E5F6",
    "status": "paid",
    "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d"
  }
}
```

---
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "SUBSCRIPTION_NOT_FOUND",
    "message": "Subscription not found"
  },
  "request_id": "..."
}
```

---

## 5. Agent Configuration & Personality

### GET `/api/v1/agent?workspace_id=:id`
Mengambil data agent utama di workspace (Canonical).

- **Auth:** JWT Bearer
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Agent retrieved",
  "data": {
    "id": "b1e3f890-...",
    "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
    "name": "Super Agent",
    "status": "ready",
    "provider": "hermes",
    "config_version": 2,
    "synced_config_version": 2
  }
}
```
- **Response `404 Not Found` (Agent belum dibuat / dihapus):**
```json
{
  "error": {
    "code": "AGENT_NOT_FOUND",
    "message": "Agent not found"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/agent/profile?workspace_id=:id`
Memperbarui profil publik agent dan template pesan (Canonical).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "display_name": "ChatSolv Assistant",
  "language": "id",
  "description": "Virtual Assistant ChatSolv Store",
  "greeting_message": "Halo! Ada yang bisa kami bantu?",
  "away_message": "Mohon maaf, saat ini di luar jam kerja.",
  "fallback_message": "Mohon tunggu, saya sambungkan dengan agen manusia."
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Agent profile sync queued",
  "data": {
    "config_version": 3,
    "status": "syncing"
  }
}
```
- **Response `400 Bad Request` (Display name kosong / field melebihi batas panjang):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid agent configuration"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/agent/personality?workspace_id=:id`
Mengatur tone, gaya komunikasi, batasan, aturan eskalasi, dan instruksi AI agent.

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "bot_name": "ChatSolv AI",
  "role": "Customer Support",
  "tone": "friendly",
  "communication_style": "conversational",
  "primary_language": "id",
  "response_length": "medium",
  "emoji_usage": "moderate",
  "greeting_style": "Halo! Ada yang bisa dibantu?",
  "closing_style": "Terima kasih telah menghubungi kami!",
  "custom_instructions": "Selalu bersikap sopan, berikan info akurat.",
  "behavior_rules": ["Jangan berikan diskon tanpa otorisasi"],
  "escalation_rules": ["Eskalasi ke human jika komplain pengiriman"],
  "forbidden_topics": ["Politik", "Kompetitor"],
  "fallback_behavior": "direct_to_human"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Personality sync queued",
  "data": {
    "config_version": 4,
    "status": "syncing"
  }
}
```
- **Response `400 Bad Request` (Tone / Style tidak sesuai enum yang diizinkan):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid personality configuration"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/agent/test?workspace_id=:id`
Playground testing agent tanpa mengirim ke channel produksi.

- **Auth:** JWT Bearer
- **Request Body:**
```json
{
  "message": "Halo, apakah melayani COD?"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Agent test completed",
  "data": {
    "conversation_id": "79bda5ac-...",
    "content": "Halo! Ya, kami melayani COD untuk area Jabodetabek melalui kurir J&T."
  }
}
```
- **Response `400 Bad Request` (Message kosong):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Message is required"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Agent / Second brain belum siap):**
```json
{
  "error": {
    "code": "AGENT_NOT_FOUND",
    "message": "Agent not found"
  },
  "request_id": "..."
}
```

---

## 6. Business Settings & Policies

### PATCH `/api/v1/business?workspace_id=:id`
Mengatur identitas bisnis, industri, jam operasional, dan brand voice workspace.

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "business_name": "ChatSolv Indonesia",
  "industry": "E-Commerce",
  "business_description": "Platform perpesanan AI untuk toko online.",
  "website": "https://chatsolv.com",
  "address": "Jakarta Selatan, Indonesia",
  "timezone": "Asia/Jakarta",
  "brand_voice": "Ramah dan profesional",
  "company_values": ["Inovasi", "Kecepatan", "Kepuasan Pelanggan"]
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Business profile updated",
  "data": {
    "config_version": 5,
    "status": "syncing"
  }
}
```
- **Response `400 Bad Request` (Format website salah / timezone tidak valid):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid business configuration"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/settings/workspaces/:id/policies`
Mengatur kebijakan bisnis toko (pengiriman, retur, pengembalian dana, garansi, pembayaran, komplain).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "shipping_policy": "Pengiriman Senin - Sabtu via JNE dan SiCepat.",
  "refund_policy": "Pengembalian dana 100% jika produk cacat pabrik.",
  "return_policy": "Retur maksimal 7 hari setelah barang sampai.",
  "warranty_policy": "Garansi resmi distributor 1 tahun.",
  "payment_policy": "Menerima QRIS, transfer bank BCA/Mandiri, dan kartu kredit.",
  "complaint_policy": "Pengaduan diproses maksimal 1x24 jam."
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Business policies updated",
  "data": {
    "config_version": 6,
    "status": "syncing"
  }
}
```

---

## 7. Channels & WhatsApp Integration

### GET `/api/v1/channels?workspace_id=:id`
Mengambil daftar seluruh channel komunikasi yang terhubung ke workspace.

- **Auth:** JWT Bearer
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Channels retrieved",
  "data": [
    {
      "id": "7326b803-fab7-4305-9070-149e0bdf69a4",
      "type": "whatsapp",
      "display_name": "Customer Service WA",
      "phone_number": "6281234567890",
      "status": "connected",
      "connected_at": "2026-08-25T12:30:00Z"
    }
  ]
}
```

---

### POST `/api/v1/channels/whatsapp/connect?workspace_id=:id`
Menginisialisasi koneksi pairing WhatsApp baru ke WhatsApp Bot Service (whatsmeow).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "display_name": "Official CS WhatsApp"
}
```
- **Contoh Request:**
```bash
curl -s -X POST "http://127.0.0.1:3000/api/v1/channels/whatsapp/connect?workspace_id=$WORKSPACE_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name": "Official CS WhatsApp"}'
```
- **Response `202 Accepted`:**
```json
{
  "success": true,
  "message": "WhatsApp pairing initiated",
  "data": {
    "channel": {
      "id": "7326b803-fab7-4305-9070-149e0bdf69a4",
      "type": "whatsapp",
      "display_name": "Official CS WhatsApp",
      "status": "waiting_pairing"
    },
    "pairing": {
      "session_id": "7326b803-fab7-4305-9070-149e0bdf69a4",
      "status": "waiting_pairing",
      "qr": "2@ABC123xyz... (raw whatsmeow QR code matrix)"
    }
  }
}
```
- **Response `409 Conflict` (Kuota channel workspace sudah penuh):**
```json
{
  "error": {
    "code": "CHANNEL_QUOTA_EXCEEDED",
    "message": "Channel quota exceeded"
  },
  "request_id": "..."
}
```
- **Response `500 Internal Server Error` (WhatsApp Bot Service down / timeout):**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "WhatsApp bot service connection failed"
  },
  "request_id": "..."
}
```

---

### DELETE `/api/v1/channels/:id`
Memutuskan sesi WhatsApp dan menghapus channel dari workspace.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Channel deleted"
}
```
- **Response `404 Not Found` (Channel ID tidak ditemukan / beda workspace):**
```json
{
  "error": {
    "code": "CHANNEL_NOT_FOUND",
    "message": "Channel or workspace not found"
  },
  "request_id": "..."
}
```

---

## 8. Knowledge Base Management

### POST `/api/v1/knowledge/faqs`
Menambahkan data tanya-jawab (FAQ) ke Second Brain.

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
  "title": "FAQ Pengiriman & Pembayaran",
  "faqs": [
    {
      "question": "Berapa lama pengiriman?",
      "answer": "Pengiriman reguler membutuhkan waktu 2-3 hari kerja."
    }
  ]
}
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "FAQ knowledge source created",
  "data": {
    "id": "k_faq_12345",
    "title": "FAQ Pengiriman & Pembayaran",
    "type": "faq",
    "status": "pending"
  }
}
```
- **Response `400 Bad Request` (Daftar FAQ kosong):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid FAQ knowledge input"
  },
  "request_id": "..."
}
```
- **Response `500 Internal Server Error` (MinIO Object Storage unreachable):**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Knowledge operation failed"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/knowledge/text`
Menambahkan data teks bebas / artikel SOP ke Second Brain.

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
  "title": "SOP Retur Barang",
  "content": "Syarat retur barang: 1. Video unboxing tanpa jeda. 2. Maksimal 3 hari."
}
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "Text knowledge source created",
  "data": {
    "id": "k_txt_67890",
    "title": "SOP Retur Barang",
    "type": "text",
    "status": "pending"
  }
}
```
- **Response `400 Bad Request` (Content teks kosong):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid text knowledge input"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/knowledge/documents`
Mengunggah file dokumen PDF / TXT / DOCX.

- **Auth:** JWT Bearer (Owner / Admin)
- **Content-Type:** `multipart/form-data`
- **Contoh Request:**
```bash
curl -s -X POST http://127.0.0.1:3000/api/v1/knowledge/documents \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "workspace_id=$WORKSPACE_ID" \
  -F "title=Katalog Produk 2026" \
  -F "file=@/path/to/katalog.pdf"
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "Document uploaded and queued for processing",
  "data": {
    "id": "k_doc_11223",
    "title": "Katalog Produk 2026",
    "type": "document",
    "status": "processing"
  }
}
```
- **Response `400 Bad Request` (Tipe file tidak didukung / ukuran file melebihi batas 10MB):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Unsupported file type. Only PDF, TXT, and DOCX are allowed."
  },
  "request_id": "..."
}
```

---

### GET `/api/v1/knowledge?workspace_id=:id`
Mengambil daftar knowledge source di workspace.

- **Auth:** JWT Bearer
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Knowledge sources retrieved",
  "data": [
    {
      "id": "k_faq_12345",
      "title": "FAQ Pengiriman & Pembayaran",
      "type": "faq",
      "status": "ready",
      "created_at": "2026-08-25T12:00:00Z"
    }
  ]
}
```

---

### PATCH `/api/v1/knowledge/:id`
Memperbarui judul knowledge source.

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "title": "FAQ Pengiriman & Pembayaran (Updated 2026)"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Knowledge source updated",
  "data": {
    "id": "k_faq_12345",
    "title": "FAQ Pengiriman & Pembayaran (Updated 2026)"
  }
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "KNOWLEDGE_NOT_FOUND",
    "message": "Knowledge source not found"
  },
  "request_id": "..."
}
```

---

### DELETE `/api/v1/knowledge/:id`
Menghapus knowledge source dari database dan object store.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Knowledge source deleted"
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "KNOWLEDGE_NOT_FOUND",
    "message": "Knowledge source not found"
  },
  "request_id": "..."
}
```

---

## 9. Conversations & Human Handoff

### GET `/api/v1/conversations?workspace_id=:id`
Mengambil daftar percakapan pelanggan.

- **Auth:** JWT Bearer
- **Query Params:**
  - `workspace_id`: ID workspace (wajib)
  - `status`: `"open" | "closed"` (opsional)
  - `mode`: `"agent" | "human"` (opsional)
  - `limit`: integer maks 100 (opsional)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Conversations retrieved",
  "data": [
    {
      "id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
      "workspace_id": "87687adc-28a9-4412-830b-f6c99e7a9e2d",
      "external_user_id": "6281234567890",
      "status": "open",
      "mode": "agent",
      "started_at": "2026-08-25T13:00:00Z",
      "last_message_at": "2026-08-25T13:05:00Z"
    }
  ]
}
```

---

### GET `/api/v1/conversations/:id/messages`
Mengambil daftar pesan dalam sebuah percakapan dengan cursor pagination RFC3339.

- **Auth:** JWT Bearer
- **Query Params:**
  - `cursor`: RFC3339 timestamp (opsional)
  - `limit`: jumlah pesan, maks 100 (opsional)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Messages retrieved",
  "data": [
    {
      "id": "m_111",
      "conversation_id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
      "sender_type": "customer",
      "content_type": "text",
      "content": "Halo, apakah pesanan saya sudah dikirim?",
      "created_at": "2026-08-25T13:00:00Z"
    },
    {
      "id": "m_112",
      "conversation_id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
      "sender_type": "agent",
      "content_type": "text",
      "content": "Halo! Boleh infokan nomor pesanan Anda agar saya bantu cek?",
      "created_at": "2026-08-25T13:00:02Z"
    }
  ]
}
```
- **Response `400 Bad Request` (Format cursor bukan RFC3339):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid cursor format"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Conversation tidak ada / beda tenant):**
```json
{
  "error": {
    "code": "CONVERSATION_NOT_FOUND",
    "message": "Conversation not found"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/conversations/:id/mode`
Mengubah mode percakapan: Takeover oleh CS Manusia (`human`) atau kembalikan ke AI (`agent`).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "mode": "human"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Conversation mode updated",
  "data": {
    "id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
    "mode": "human"
  }
}
```
- **Response `400 Bad Request` (Mode bukan `human` atau `agent`):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid conversation input"
  },
  "request_id": "..."
}
```
- **Response `403 Forbidden` (User adalah role `member`):**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "You cannot perform this action"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Conversation tidak ditemukan):**
```json
{
  "error": {
    "code": "CONVERSATION_NOT_FOUND",
    "message": "Conversation not found"
  },
  "request_id": "..."
}
```

---

## 10. Developer API Keys

### GET `/api/v1/api-keys?workspace_id=:id`
Mengambil daftar API Key rahasia workspace.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "API keys retrieved",
  "data": [
    {
      "id": "ak_123",
      "name": "Production Server Key",
      "prefix": "cs_live_prod_abc",
      "last_four": "9xyz",
      "scopes": ["agent:invoke", "knowledge:read"],
      "created_at": "2026-08-25T12:00:00Z"
    }
  ]
}
```

---

### POST `/api/v1/api-keys?workspace_id=:id`
Membuat Secret API Key baru. **Raw key `cs_live_*` hanya ditampilkan SATU KALI pada response ini.**

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "name": "Production Server Key",
  "scopes": ["agent:invoke", "knowledge:read"]
}
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "API key created. Save this secret now; it will not be shown again.",
  "data": {
    "key": "cs_live_a1b2c3d4e5f6...",
    "api_key": {
      "id": "ak_456",
      "name": "Production Server Key",
      "prefix": "cs_live_a1b2",
      "last_four": "e5f6",
      "scopes": ["agent:invoke", "knowledge:read"],
      "created_at": "2026-08-25T12:00:00Z"
    }
  }
}
```
- **Response `400 Bad Request` (Nama kosong / scopes tidak valid):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid API key input"
  },
  "request_id": "..."
}
```

---

### DELETE `/api/v1/api-keys/:id`
Mencabut (revoke) API Key.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "API key revoked"
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "API_KEY_NOT_FOUND",
    "message": "API key not found"
  },
  "request_id": "..."
}
```

---

## 11. Developer Webhooks

### GET `/api/v1/webhooks?workspace_id=:id`
Mengambil daftar endpoint webhook yang terdaftar di workspace.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Webhooks retrieved",
  "data": [
    {
      "id": "wh_001",
      "workspace_id": "87687adc-...",
      "url": "https://api.myapp.com/webhooks/chatsolv",
      "events": ["message.created", "conversation.created"],
      "status": "active",
      "created_at": "2026-08-25T12:00:00Z"
    }
  ]
}
```

---

### POST `/api/v1/webhooks?workspace_id=:id`
Mendaftarkan webhook baru. URL wajib menggunakan protokol HTTPS. Signing secret `whsec_*` dikembalikan pada response ini dan dienkripsi AES-GCM di database.

- **Auth:** JWT Bearer (Owner / Admin)
- **Prasyarat:** Subscription workspace memiliki entitlement `webhooks: true`
- **Request Body:**
```json
{
  "url": "https://api.myapp.com/webhooks/chatsolv",
  "events": ["message.created", "handoff.requested"]
}
```
- **Response `201 Created`:**
```json
{
  "success": true,
  "message": "Webhook created",
  "data": {
    "endpoint": {
      "id": "wh_002",
      "workspace_id": "87687adc-...",
      "url": "https://api.myapp.com/webhooks/chatsolv",
      "events": ["message.created", "handoff.requested"],
      "status": "active"
    },
    "secret": "whsec_9xyzABC123..."
  }
}
```
- **Response `400 Bad Request` (URL tidak menggunakan HTTPS / event tidak dikenal):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid webhook input (HTTPS URL required)"
  },
  "request_id": "..."
}
```
- **Response `403 Forbidden` (Paket subscription belum mengaktifkan fitur Webhook):**
```json
{
  "error": {
    "code": "SUBSCRIPTION_REQUIRED",
    "message": "Webhook entitlement is required"
  },
  "request_id": "..."
}
```

---

### PATCH `/api/v1/webhooks/:id`
Memperbarui URL, events, atau status webhook (`active` / `disabled`).

- **Auth:** JWT Bearer (Owner / Admin)
- **Request Body:**
```json
{
  "url": "https://api.myapp.com/webhooks/chatsolv-v2",
  "events": ["message.created"],
  "status": "active"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Webhook updated",
  "data": {
    "id": "wh_002",
    "url": "https://api.myapp.com/webhooks/chatsolv-v2",
    "status": "active"
  }
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "WEBHOOK_NOT_FOUND",
    "message": "Webhook not found"
  },
  "request_id": "..."
}
```

---

### DELETE `/api/v1/webhooks/:id`
Menghapus / menonaktifkan endpoint webhook.

- **Auth:** JWT Bearer (Owner / Admin)
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Webhook deleted"
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "WEBHOOK_NOT_FOUND",
    "message": "Webhook not found"
  },
  "request_id": "..."
}
```

---

## 12. Public Website Agent API

### POST `/api/v1/agent-sessions`
Membuat ephemeral session untuk widget website publik dari server customer.

- **Auth:** Secret API Key (`Authorization: Bearer cs_live_...`)
- **Request Body:**
```json
{
  "external_user_id": "visitor_session_12345",
  "metadata": {
    "current_page": "/pricing"
  }
}
```
- **Contoh Request:**
```bash
curl -s -X POST http://127.0.0.1:3000/api/v1/agent-sessions \
  -H "Authorization: Bearer cs_live_your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{
    "external_user_id": "visitor_session_12345",
    "metadata": { "current_page": "/pricing" }
  }'
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Agent session created",
  "data": {
    "session_id": "sess_998877",
    "client_token": "cs_pub_a8f9...",
    "expires_at": "2026-08-25T14:00:00Z"
  }
}
```
- **Response `401 Unauthorized` (Secret Key salah / dicabut):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid API key"
  },
  "request_id": "..."
}
```
- **Response `403 Forbidden` (Key tidak memiliki scope `agent:invoke` atau paket tidak mengaktifkan `public_api`):**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Scope agent:invoke is required"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/agent-sessions/:id/messages`
Mengirim pesan dari widget website ke AI Agent (Single JSON response).

- **Auth:** Ephemeral Client Token (`Authorization: Bearer cs_pub_...`)
- **Request Body:**
```json
{
  "message": "Apakah paket Pro mendukung integrasi WhatsApp?"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Message processed",
  "data": {
    "message_id": "msg_554433",
    "content": "Ya! Paket Pro sudah mencakup integrasi penuh WhatsApp bot hingga 5 channel."
  }
}
```
- **Response `401 Unauthorized` (Client token expired > 1 jam / invalid):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Session token expired or invalid"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Session ID tidak ditemukan):**
```json
{
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Agent session not found"
  },
  "request_id": "..."
}
```

---

### POST `/api/v1/agent-sessions/:id/messages/stream`
Mengirim pesan dari widget website dan menerima stream Server-Sent Events (SSE).

- **Auth:** Ephemeral Client Token (`Authorization: Bearer cs_pub_...`)
- **Request Body:**
```json
{
  "message": "Jelaskan langkah-langkah setup awal ChatSolv"
}
```
- **Response Stream (`200 OK`, `Content-Type: text/event-stream`):**
```
event: token
data: {"token":"Langkah "}

event: token
data: {"token":"pertama: "}

event: done
data: {"content":"Langkah pertama: Daftarkan akun dan hubungkan nomor WhatsApp Anda."}
```
- **Response `401 Unauthorized`:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Session token expired or invalid"
  },
  "request_id": "..."
}
```

---

## 13. Internal Service Routes (HMAC Authenticated)

Semua rute `/internal/v1/*` dilindungi dengan signature HMAC SHA-256 dan batas toleransi waktu replay 5 menit.

**Header Wajib:**
- `X-ChatSolv-Timestamp`: Timestamp UTC ISO8601/RFC3339 (misal `2026-08-25T12:00:00Z`).
- `X-ChatSolv-Signature`: `hex(HMAC-SHA256(INTERNAL_SERVICE_SECRET, timestamp + "." + request_body))`.

### POST `/internal/v1/messages/incoming`
Callback dari WhatsApp Bot Service ketika menerima pesan masuk dari WhatsApp user. Backend menjalankan AI runtime dan mengembalikan jawaban di JSON agar bot mengirimkannya ke chat WhatsApp.

- **Auth:** Internal HMAC
- **Request Body:**
```json
{
  "channel_id": "7326b803-fab7-4305-9070-149e0bdf69a4",
  "external_message_id": "wamid_ABC123xyz",
  "external_user_id": "6281234567890",
  "message_type": "text",
  "content": {
    "text": "Halo kak, apakah ada diskon untuk pembelian 5 unit?"
  }
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Message processed",
  "data": {
    "message_id": "m_9988",
    "conversation_id": "98bb6c6d-...",
    "content": "Halo! Untuk pembelian 5 unit ke atas, Anda mendapatkan diskon grosir 10%. Silakan lanjutkan pemesanan.",
    "handoff_requested": false,
    "handoff_reason": null
  }
}
```
- **Response `400 Bad Request` (Payload tidak lengkap / message_type bukan text):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid internal request"
  },
  "request_id": "..."
}
```
- **Response `401 Unauthorized` (Timestamp melebihi selisih 5 menit / signature HMAC salah):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid internal signature"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Channel ID tidak terdaftar di database):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Internal resource not found"
  },
  "request_id": "..."
}
```

---

### POST `/internal/v1/channels/status`
Update status koneksi channel dari WhatsApp Bot Service.

- **Auth:** Internal HMAC
- **Valid Status:** `"waiting_pairing" | "connecting" | "connected" | "reconnecting" | "disconnected" | "error" | "suspended"`
- **Request Body:**
```json
{
  "channel_id": "7326b803-fab7-4305-9070-149e0bdf69a4",
  "status": "connected",
  "phone_number": "6281234567890",
  "session_id": "7326b803-fab7-4305-9070-149e0bdf69a4"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Channel status updated"
}
```
- **Response `400 Bad Request` (Status bukan salah satu enum yang valid):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid channel status"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found`:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Channel not found"
  },
  "request_id": "..."
}
```

---

### POST `/internal/v1/channels/events`
Mencatat event lifecycle channel (misal `qr_refresh`, `pair_success`, `logged_out`).

- **Auth:** Internal HMAC
- **Request Body:**
```json
{
  "channel_id": "7326b803-fab7-4305-9070-149e0bdf69a4",
  "event": "qr_refresh",
  "phone_number": "",
  "session_id": "7326b803-fab7-4305-9070-149e0bdf69a4"
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Channel event processed"
}
```
- **Response `401 Unauthorized`:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid internal signature"
  },
  "request_id": "..."
}
```

---

### GET `/internal/v1/agents/:agentID/health`
Pemeriksaan kesehatan AI agent runtime internal.

- **Auth:** Internal HMAC
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Agent health retrieved",
  "data": {
    "agent_id": "b1e3f890-...",
    "status": "ready",
    "brain_status": "ready",
    "ready": true
  }
}
```
- **Response `404 Not Found` (Agent ID tidak ada di database):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Agent not found"
  },
  "request_id": "..."
}
```

---

### POST `/internal/v1/agents/:agentID/respond`
Trigger eksekusi runtime AI internal untuk percakapan tertentu.

- **Auth:** Internal HMAC
- **Request Body:**
```json
{
  "conversation_id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
  "message": "Lanjutkan percakapan sebelumnya."
}
```
- **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Response generated",
  "data": {
    "conversation_id": "98bb6c6d-455a-4191-9ffb-a1b268e15eeb",
    "content": "Tentu, saya siap membantu. Ada hal lain yang perlu diperjelas?"
  }
}
```
- **Response `400 Bad Request` (Message kosong):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid agent respond input"
  },
  "request_id": "..."
}
```
- **Response `404 Not Found` (Conversation ID atau Agent ID tidak cocok):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Conversation or agent not found"
  },
  "request_id": "..."
}
```
