# TESTING RULES

## ChatSolv Backend — MVP Authentication & System Standard

> Testing adalah WAJIB. Backend tidak dianggap selesai hanya karena code berhasil compile atau endpoint dapat dipanggil secara manual.

---

# 1. Tujuan Testing

Testing digunakan untuk memastikan:
* Behavior sesuai kebutuhan PRD.
* Authentication & Authorization aman dari vulnerability dasar.
* Tenant Boundary & Data Isolation tidak jebol.
* Error Mapping dan API Contract konsisten.
* Redis & PostgreSQL query berjalan secara benar dan aman.
* Performance & Latency target terpelihara tanpa regression.

---

# 2. Scope Test MVP Saat Ini

Scope awal wajib diuji pada endpoint Authentication dasar:
```text
POST /v1/auth/register
POST /v1/auth/login
POST /v1/auth/forgot-password
POST /v1/auth/reset-password
```
Setiap penambahan endpoint/fitur baru (Workspace, Ingestion, Channel, Runtime) WAJIB disertai dengan unit/integration test terkait.

---

# 3. Testing Pyramid & Test Types

```text
        E2E Tests       (Critical Flow Utama)
       /         \
  Integration Tests    (DB, Redis, Storage Boundary)
 /                 \
Unit Tests           (Domain Logic, Validation, Parsers)
```

**Unit Tests:**
Wajib untuk pure logic tanpa dependency eksternal (Validation, Error Mapping, Token Claims, Path Resolver, Password Policy).

**Integration Tests:**
Wajib menggunakan PostgreSQL & Redis instance nyata untuk menguji Query, Constraint, Transaction, Indexing, dan Connection Handling. Mocking database dilarang untuk SQL assertion.

**E2E Tests:**
Menguji alur lengkap via HTTP Layer dari Register → Login → Reset Password → Re-login.

---

# 4. Standard Convention & Naming

* **Testing Package:** Gunakan standard library `testing`. Tambahkan assertion helper hanya jika meningkatkan readability.
* **Test Naming:** Harus deskriptif menjelaskan behavior.
  * *Good:* `TestLogin_InvalidPassword_ReturnsUnauthorized(t *testing.T)`
  * *Bad:* `TestLogin1(t *testing.T)`
* **Data Isolation:** Setiap integration test wajib terisolasi (menggunakan transaction rollback atau DB cleanup) agar test data tidak saling mempengaruhi.

---

# 5. Security & Isolation Specific Tests

**Data Redaction Test:**
Test log/output DILARANG mencetak password, JWT, Refresh Token, Reset Token, maupun API Secret Keys.

**Tenant & Path Security Tests (Wajib Sebelum Release):**
* **Tenant Isolation Test:** Memastikan User Tenant A tidak bisa membaca/mengubah Agent, Vault, Knowledge, atau Key milik Tenant B.
* **Obsidian Path Traversal Test:** Input berbahaya seperti `../../`, `../workspace_B`, atau URL encoded path traversal harus ditolak oleh path resolver backend.
* **Hermes Second Brain Isolation Test:** Memastikan Hermes Agent Tenant A tidak mendapatkan konteks data dari Vault Tenant B.

---

# 6. Execution Command Standard

```bash
# Run unit & integration tests
go test ./...

# Run static analysis
go vet ./...

# Run race detector (Mandatory for CI)
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

---

# 7. Rules untuk AI Assistant / Developer

1. Wajib membaca `PRD.md`, `PERFORMANCE.md`, dan `TESTING_RULES.md` sebelum melakukan refactoring/coding.
2. Dilarang menghapus atau mentoleransi skip test (`t.Skip()`) hanya agar pipeline CI menjadi hijau.
3. Dilarang melakukan hardcode sleep (`time.Sleep()`) untuk menangani race condition pada test async. Gunakan channels atau sync primitives.
4. Setiap kegagalan test harus diperbaiki dari implementasi bisnis atau koreksi assertion spec.
