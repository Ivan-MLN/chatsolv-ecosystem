# ChatSolv Backend Benchmark Suite (k6 & Go Tests)

Direktori ini berisi kumpulan script benchmark performa backend ChatSolv menggunakan **Grafana k6** dan **Go Micro-benchmarks (`testing.B`)**.

---

## 1. Daftar Script Benchmark

| File | Tool | Deskripsi | Target Metrik |
|---|---|---|---|
| `bench/k6_all_routes.js` | k6 | Comprehensive benchmark seluruh API routes fungsional (Health, Me, Workspace, Agent, Business Settings, Policies, Channels, Knowledge, Conversations, API Keys, Webhooks) | Latency p95 < 50ms, Error rate < 1% |
| `bench/k6_stress_test.js` | k6 | Stress / Load test bertahap (Ramp up -> 25 VUs steady -> Ramp down) pada Liveness & Readiness endpoints | Latency p95 < 50ms, 0% failure rate |
| `bench/k6_auth_flow.js` | k6 | Benchmark siklus Register -> Login -> Me -> Token Rotation Refresh | Latency p95 < 400ms (Argon2id CPU-bound) |
| `bench/router_bench_test.go` | Go `testing.B` | Micro-benchmark Go internal: Routing Fiber, Serialisasi Response JSON, JWT Gen/Verify, Argon2id, HMAC SHA-256 Signature, AES-GCM Webhook Crypto, dan UUID Parsing | Allocs/op & ns/op memory efficiency |

---

## 2. Cara Menjalankan Backend untuk Benchmark

Saat menjalankan benchmark dengan konkurensi tinggi, Anda dapat meng-override batas rate limit menggunakan perintah `make RATE_LIMIT_MAX=<N> run`:

```bash
# 1. Menjalankan backend dengan kuota rate-limit tinggi khusus benchmark (misal: 10000 req/min)
make RATE_LIMIT_MAX=10000 run

# 2. Menjalankan backend dengan kuota rate-limit standar produksi/pengembangan
make run
```

---

## 3. Cara Menjalankan Benchmark k6

### Prasyarat:
- Server ChatSolv Backend berjalan (`http://127.0.0.1:3000`).
- Database PostgreSQL & Redis aktif.
- Tool `k6` terpasang (`k6 version`).

---

### A. Menjalankan Comprehensive Route Benchmark (k6)
Menguji performa seluruh endpoint fungsional dengan 10-20 Virtual Users:

```bash
# Default: 10 VUs, 50 Iterasi
VUS=10 ITERATIONS=50 k6 run bench/k6_all_routes.js

# Kustomisasi Concurrency Tinggi: 20 VUs, 100 Iterasi
VUS=20 ITERATIONS=100 k6 run bench/k6_all_routes.js
```

---

### B. Menjalankan Stress & Load Test (k6)
Menguji ketahanan throughput backend di bawah beban konkurensi bertahap hingga 25 virtual users:

```bash
k6 run bench/k6_stress_test.js
```

---

### C. Menjalankan Auth Flow Lifecycle Benchmark (k6)
Menguji siklus autentikasi lengkap (Register -> Login -> Me -> Token Rotation Refresh):

```bash
ITERATIONS=10 k6 run bench/k6_auth_flow.js
```

---

### D. Menjalankan Micro-Benchmark Internal Go
Mengukur efisiensi alokasi memori dan micro-latency komponen kriptografi dan router:

```bash
go test -bench=. -benchmem ./bench
```

---

## 4. Hasil Pengujian Benchmark Baseline (`RATE_LIMIT_MAX=10000`)

### Hasil k6 Comprehensive All-Routes (20 Concurrent VUs, 100 Iterations):
- **Total Checks**: **`2.400 / 2.400 SUKSES (100.00%)`**
- **HTTP Failure Rate**: **`0.00%`** (0 error dari 2.403 request)
- **Throughput**: **`881.6 requests / detik`**
- **Median Latency**: **`4.2 ms`**
- **Average Latency**: **`15.03 ms`**
- **p95 Latency**: **`70.36 ms`**

### Hasil k6 Stress Test (25 Concurrent VUs):
- **Total Requests**: **`14.370 requests`**
- **Throughput**: **`717.7 requests / detik`**
- **Average Latency**: **`1.10 ms`**
- **p95 Latency**: **`3.01 ms`**
- **p99 Latency**: **`4.47 ms`**
- **HTTP Failure Rate**: **`0.00%`** (14.370 / 14.370 requests sukses)

### Hasil k6 Auth Lifecycle (Register -> Login -> Me -> Refresh):
- **Total Checks**: **`40 / 40 SUKSES (100.00%)`**
- **HTTP Failure Rate**: **`0.00%`**
- **Average Latency**: **`104.75 ms`** *(dipengaruhi Argon2id password hashing)*
- **p95 Latency**: **`232.91 ms`**

### Hasil Micro-Benchmark Go Internal:
- **UUID Parsing**: `48.07 ns/op` (0 allocs/op)
- **AES-GCM Webhook Decryption**: `213.1 ns/op` (1 alloc/op)
- **AES-GCM Webhook Encryption**: `228.3 ns/op` (1 alloc/op)
- **HMAC-SHA256 Microservice Signature**: `1.59 µs/op` (10 allocs/op)
- **JWT Generation**: `13.9 µs/op` (38 allocs/op)
- **Fiber Route & JWT Verify**: `41.0 µs/op` (61 allocs/op)
- **Envelope JSON Serialization**: `67.8 µs/op` (82 allocs/op)
- **Argon2id Hash/Verify**: `~132 ms/op` (OWASP anti brute-force standard)
