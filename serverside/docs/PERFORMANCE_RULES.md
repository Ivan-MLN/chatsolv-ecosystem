# PERFORMANCE & LOW LATENCY ENGINEERING RULES

## 1. PRIMARY PERFORMANCE OBJECTIVE

Backend harus dirancang untuk menghasilkan **response time serendah mungkin** pada seluruh request.

Performance bukan fitur tambahan.

Performance adalah bagian dari architecture.

Target utama:

```text
Low Latency
+
High Throughput
+
Efficient Resource Usage
+
Predictable Performance
```

Namun:

> Performance tidak boleh dicapai dengan mengorbankan security, correctness, reliability, atau maintainability.

---

# 2. EVERY REQUEST MATTERS

AI harus memperhatikan performance pada **setiap request**, bukan hanya endpoint tertentu.

Endpoint v1:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password

GET /health
GET /ready
```

Setiap endpoint harus dianalisis:

```text
Request
 ↓
Middleware
 ↓
Handler
 ↓
Validation
 ↓
Service
 ↓
Database / Redis
 ↓
Serialization
 ↓
Response
```

Cari unnecessary latency pada setiap tahap.

---

# 3. LATENCY BUDGET

Setiap request harus memiliki mindset:

```text
Total Response Time
=
Middleware
+
Validation
+
Business Logic
+
Database
+
Redis
+
External Service
+
Serialization
```

AI harus menghindari operasi yang tidak diperlukan.

Contoh buruk:

```text
Request
 ↓
DB query
 ↓
Redis query
 ↓
DB query
 ↓
Redis query
 ↓
DB query
 ↓
Response
```

Jika business logic dapat dilakukan dengan:

```text
Request
 ↓
1 DB query
 ↓
1 Redis operation
 ↓
Response
```

maka gunakan desain kedua.

---

# 4. DATABASE IS OFTEN THE BOTTLENECK

AI harus menganggap database sebagai salah satu sumber latency terbesar.

Sebelum menambahkan query baru, tanyakan:

```text
Apakah query ini benar-benar diperlukan?
```

Kemudian:

```text
Apakah query dapat digabung?
Apakah query dapat dibuat lebih sederhana?
Apakah column yang diambil terlalu banyak?
Apakah index tersedia?
Apakah query menyebabkan sequential scan?
Apakah ada N+1 query?
```

---

# 5. AVOID N+1 QUERY

Jangan melakukan:

```text
Get User
 ↓
Get something
 ↓
Get something
 ↓
Get something
```

jika dapat dilakukan secara efisien dengan query yang sesuai.

Untuk authentication v1, jumlah database query harus dibuat seminimal mungkin.

---

# 6. SELECT ONLY REQUIRED COLUMNS

Hindari:

```sql
SELECT *
FROM users
WHERE email = $1;
```

Gunakan hanya field yang dibutuhkan.

Contoh:

```sql
SELECT
    id,
    email,
    password_hash
FROM users
WHERE email = $1;
```

Jika login hanya membutuhkan:

```text
id
email
password_hash
```

jangan mengambil:

```text
name
created_at
updated_at
```

jika tidak digunakan.

---

# 7. DATABASE INDEX

Setiap query yang digunakan pada hot path harus diperiksa index-nya.

Contoh:

```sql
WHERE email = $1
```

harus memiliki index/unique constraint yang sesuai.

Jangan membuat index secara berlebihan.

Index memiliki cost:

```text
Read Performance
       ↑
       |
Index
       ↓
Write Cost + Storage
```

Gunakan index berdasarkan query pattern nyata.

---

# 8. CONNECTION POOL

PostgreSQL harus menggunakan connection pool.

Gunakan:

```text
pgxpool
```

Jangan membuat koneksi database baru pada setiap request.

Connection pool harus dikonfigurasi melalui environment.

Perhatikan:

```text
MaxConns
MinConns
MaxConnLifetime
MaxConnIdleTime
HealthCheckPeriod
```

Jangan menetapkan angka secara asal.

Nilai harus masuk akal untuk deployment environment.

---

# 9. CONNECTION POOL EXHAUSTION

AI harus menghindari:

* query yang terlalu lama
* transaction yang terlalu panjang
* connection leak
* query tanpa timeout
* operasi network di dalam transaction

Jika connection pool penuh, latency seluruh application dapat meningkat drastis.

---

# 10. DATABASE TIMEOUT

Database operation harus memiliki context.

Contoh konsep:

```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```

Tujuannya mencegah request menggantung terlalu lama.

Timeout harus configurable.

---

# 11. TRANSACTION PERFORMANCE

Transaction harus sesingkat mungkin.

Jangan melakukan:

```text
BEGIN
 ↓
Database query
 ↓
HTTP request
 ↓
Redis
 ↓
Email API
 ↓
Database query
 ↓
COMMIT
```

Jangan melakukan external network request di dalam transaction.

Gunakan:

```text
BEGIN
 ↓
Required DB operations
 ↓
COMMIT
```

secepat mungkin.

---

# 12. REDIS PERFORMANCE

Redis digunakan karena latency-nya rendah.

Namun jangan melakukan Redis operation jika tidak diperlukan.

Hindari:

```text
Redis GET
 ↓
Redis GET
 ↓
Redis GET
```

jika satu operation dapat memenuhi kebutuhan.

Gunakan atomic operation atau pipeline jika memang sesuai.

Namun jangan menggunakan pipeline hanya untuk terlihat "optimized".

---

# 13. REDIS KEY DESIGN

Redis key harus predictable dan efficient.

Contoh:

```text
auth:refresh:<id>
auth:reset:<hash>
rate:login:<identifier>
```

Key jangan terlalu panjang jika tidak diperlukan.

Temporary data harus menggunakan TTL.

---

# 14. PASSWORD HASHING PERFORMANCE

Password hashing memang sengaja membutuhkan computational cost.

Jangan menurunkan password hashing cost hanya demi benchmark response time.

Security requirement lebih penting.

Gunakan password hashing configuration yang aman dan reasonable.

Targetnya:

```text
Secure enough
+
Fast enough
```

bukan:

```text
As fast as possible
```

---

# 15. JWT PERFORMANCE

JWT generation dan verification harus dilakukan secara efficient.

Jangan memasukkan payload yang tidak diperlukan.

JWT harus tetap kecil.

Hindari menyimpan data besar di token.

Contoh payload yang reasonable:

```text
sub
iat
exp
jti
```

Jangan memasukkan:

```text
large profile data
permissions array besar
database records
```

jika tidak diperlukan.

Token kecil berarti:

* less serialization
* less network transfer
* less parsing
* less memory usage

---

# 16. MIDDLEWARE PERFORMANCE

Middleware berjalan pada hampir semua request.

Karena itu middleware harus ringan.

Perhatikan:

```text
Recover
Request ID
Logger
CORS
Security Headers
Compression
Rate Limiter
```

Jangan melakukan database query pada middleware kecuali benar-benar diperlukan.

Contoh buruk:

```text
Every request
 ↓
Middleware
 ↓
SELECT user FROM database
```

---

# 17. LOGGING PERFORMANCE

Logging harus structured tetapi tidak boleh terlalu berat.

Gunakan:

```text
log/slog
```

Hindari logging berlebihan pada hot path.

Jangan melakukan expensive serialization hanya untuk log yang tidak diperlukan.

Jangan log request body authentication.

Terutama:

```text
password
token
secret
```

---

# 18. REQUEST BODY SIZE

Authentication request memiliki payload kecil.

Gunakan request body limit.

Tujuannya:

```text
Prevent Abuse
+
Reduce Memory Usage
+
Predictable Latency
```

Jangan menerima payload berukuran megabyte untuk endpoint login.

---

# 19. JSON SERIALIZATION

Gunakan serializer yang efficient.

Namun:

**Jangan mengganti serializer hanya berdasarkan benchmark internet.**

Default serializer boleh digunakan jika performance sudah memadai.

Serializer alternatif hanya dipertimbangkan jika:

```text
Measured Bottleneck
+
Compatible
+
Stable
+
Real Performance Improvement
```

---

# 20. ALLOCATION AWARENESS

AI harus menghindari allocation yang tidak diperlukan pada hot path.

Perhatikan:

* unnecessary string conversion
* unnecessary byte conversion
* unnecessary JSON marshal/unmarshal
* unnecessary copying
* unnecessary temporary objects

Namun:

> Jangan melakukan unsafe optimization hanya untuk mengurangi allocation.

Code clarity tetap penting.

---

# 21. AVOID PREMATURE OPTIMIZATION

Dilarang melakukan optimization hanya berdasarkan asumsi.

Contoh:

```text
"Ini pasti lebih cepat."
```

Tidak cukup.

Gunakan:

```text
Measure
 ↓
Identify bottleneck
 ↓
Optimize
 ↓
Benchmark
 ↓
Compare
```

---

# 22. BENCHMARK

Untuk bagian yang performance-sensitive, gunakan benchmark k6. Folder benchmark ada di bench/



Gunakan benchmark jika ada alasan nyata untuk membandingkan implementation.

Benchmark harus merepresentasikan workload yang masuk akal.

---

# 23. HOT PATH

AI harus mengidentifikasi **hot path**.

Contoh:

### Login

```text
Request
 ↓
Validation
 ↓
User lookup
 ↓
Password verification
 ↓
Token generation
 ↓
Redis
 ↓
Response
```

### Register

```text
Request
 ↓
Validation
 ↓
User insert
 ↓
Password hashing
 ↓
Response
```

### Forgot Password

```text
Request
 ↓
Validation
 ↓
User lookup
 ↓
Token generation
 ↓
Redis
 ↓
Email abstraction
 ↓
Response
```

### Reset Password

```text
Request
 ↓
Token validation
 ↓
Redis
 ↓
Password hashing
 ↓
Database update
 ↓
Response
```

AI harus memberikan perhatian khusus pada hot path tersebut.

---

# 24. PASSWORD HASHING VS RESPONSE TIME

Login dan reset password membutuhkan password hashing verification/generation.

Jangan menganggap semua latency adalah bad latency.

Password hashing memiliki intentional computational cost untuk security.

Jadi:

```text
Database latency
    → minimize

Redis latency
    → minimize

Network latency
    → minimize

Serialization
    → minimize

Password hashing
    → maintain secure work factor
```

Jangan mengurangi security hanya untuk mendapatkan benchmark yang lebih rendah.

---

# 25. EXTERNAL SERVICE

External service seperti email provider dapat menjadi sumber latency.

Jangan membuat authentication architecture bergantung pada external service secara synchronous jika tidak diperlukan.

Untuk production:

```text
Request
 ↓
Persist required state
 ↓
Queue / async processing
 ↓
Email Provider
```

Jika async email belum diperlukan untuk v1, gunakan abstraction yang sederhana.

Jangan membuat message broker hanya demi terlihat scalable.

---

# 26. HEALTH CHECK PERFORMANCE

`/health` harus sangat ringan.

Jangan:

```text
/health
 ↓
Complex database query
 ↓
Redis query
 ↓
External API
```

Health endpoint harus cepat.

`/ready` dapat melakukan dependency checks yang diperlukan, tetapi tetap harus memiliki timeout.

---

# 27. RESPONSE PAYLOAD

Response harus berisi hanya data yang dibutuhkan client.

Jangan mengirim:

```text
unused fields
debug data
internal database information
stack traces
```

Payload kecil membantu:

```text
Serialization
+
Network
+
Client parsing
```

---

# 28. COMPRESSION

Compression dapat membantu response payload yang besar.

Namun authentication response biasanya kecil.

Jangan menganggap compression selalu membuat response lebih cepat.

Untuk payload kecil:

```text
Compression overhead
>
Network savings
```

AI harus mempertimbangkan ukuran response dan workload nyata.

---

# 29. RATE LIMITER PERFORMANCE

Rate limiter harus cepat.

Jangan menggunakan database sebagai storage utama rate limit jika Redis tersedia untuk distributed rate limiting.

Rate limiting operation harus memiliki latency rendah.

---

# 30. CONTEXT PROPAGATION

Setiap dependency call harus membawa context.

Contoh:

```text
HTTP Request Context
       ↓
Service
       ↓
PostgreSQL
       ↓
Redis
```

Jika client disconnect atau timeout terjadi, pekerjaan yang tidak lagi diperlukan harus dapat dihentikan.

---

# 31. GOROUTINE RULES

Jangan membuat goroutine tanpa lifecycle yang jelas.

Dilarang:

```go
go func() {
    doSomething()
}()
```

jika operation tersebut penting tetapi tidak memiliki cancellation atau lifecycle management.

Perhatikan:

```text
goroutine leak
unbounded concurrency
race condition
```

---

# 32. MEMORY USAGE

AI harus memperhatikan memory usage.

Hindari:

* membaca data besar ke memory jika tidak diperlukan
* duplicate buffer
* unnecessary copies
* unbounded cache
* Redis data tanpa TTL
* goroutine leak

Memory pressure dapat menyebabkan:

```text
GC pressure
↓
CPU usage
↓
Latency meningkat
```

---

# 33. GARBAGE COLLECTION AWARENESS

Go garbage collector tidak boleh dilawan secara berlebihan.

Jangan menggunakan:

```text
sync.Pool
unsafe
manual memory tricks
```

kecuali terdapat benchmark yang menunjukkan masalah nyata.

Idiomatic Go lebih diutamakan.

---

# 34. DEPENDENCY PERFORMANCE

Setiap dependency baru harus dipertimbangkan dari sisi:

```text
CPU
Memory
Allocation
Latency
Dependency size
Maintenance
```

Jangan menambahkan library besar untuk pekerjaan yang dapat dilakukan standard library dengan baik.

---

# 35. RESPONSE TIME OBSERVABILITY

Application harus dapat mengukur latency.

Minimal log:

```text
request_id
method
path
status
duration
```

Contoh:

```text
method=POST
path=/api/v1/auth/login
status=200
duration=42ms
```

Tujuannya agar performance dapat diukur.

---

# 36. PERFORMANCE REGRESSION

Setiap perubahan besar harus mempertimbangkan kemungkinan performance regression.

Jika perubahan:

* menambah query
* menambah middleware
* menambah network call
* menambah serialization
* menambah dependency

AI harus mengevaluasi impact-nya.

---

# 38 USE AVAILABLE CPU EFFECTIVELY

Backend harus memanfaatkan CPU yang tersedia secara optimal.

Go runtime harus diberikan kesempatan untuk menggunakan CPU secara concurrent.

Jangan membuat artificial worker count hanya untuk "menggunakan semua CPU thread".

Do not assume:

    CPU threads = worker count

HTTP request concurrency harus dikelola oleh Go runtime dan Fiber.

---

# 38. PERFORMANCE PRIORITY

Jika AI menemukan dua solusi:

### Solution A

Lebih cepat sedikit tetapi:

* lebih kompleks
* lebih sulit dipahami
* lebih sulit ditest
* lebih banyak dependency

### Solution B

Sedikit lebih sederhana tetapi performance sudah sangat baik.

Pilih **Solution B** kecuali ada bukti bahwa Solution A memberikan keuntungan yang benar-benar signifikan.

---

# 39. GOLDEN PERFORMANCE RULE

> **Make the common path fast.**

Optimalkan hal yang paling sering terjadi:

```text
Login
Register
Password Reset
```

Jangan menghabiskan waktu mengoptimalkan code yang hampir tidak pernah dijalankan.

---

# 40. FINAL PERFORMANCE PRINCIPLE

Target project bukan:

> "Secepat mungkin tanpa peduli apa pun."

Target project adalah:

> **"Serendah mungkin latency-nya dengan architecture yang sederhana, aman, predictable, dan maintainable."**

Gunakan prinsip:

```text
Measure
 ↓
Understand
 ↓
Optimize
 ↓
Benchmark
 ↓
Verify
```

Bukan:

```text
Guess
 ↓
Micro-optimize
 ↓
Add complexity
 ↓
Hope it's faster
```

# 41. GOMAXPROCS

Jangan hardcode:

    runtime.GOMAXPROCS(1)

atau nilai kecil lainnya tanpa alasan yang jelas.

Secara default, gunakan konfigurasi GOMAXPROCS yang sesuai dengan CPU resources yang tersedia.

Jika deployment menggunakan CPU limit/container quota, pastikan konfigurasi runtime mempertimbangkan CPU yang benar-benar tersedia untuk container.

Jangan mengubah GOMAXPROCS hanya berdasarkan jumlah logical CPU host jika application berjalan di container dengan CPU limit yang lebih kecil.

---

# 42. DO NOT CREATE ARTIFICIAL WORKERS

Jangan membuat worker pool global hanya untuk memproses HTTP request.

Hindari architecture seperti:

    HTTP
     ↓
    Worker Queue
     ↓
    Fixed 8 Workers
     ↓
    Handler

jika pekerjaan tersebut sebenarnya dapat ditangani langsung oleh goroutine/request lifecycle Go.

Fiber + Go runtime sudah menangani concurrent request processing.

---

# 43. GOROUTINE USAGE

Gunakan goroutine jika pekerjaan memang dapat dilakukan secara concurrent.

Contoh yang masuk akal:

    Request
     ├── Redis operation
     └── independent operation

Namun jangan membuat goroutine hanya untuk:

    "memakai semua CPU".

Setiap goroutine harus memiliki lifecycle yang jelas.

Perhatikan:

- goroutine leak
- unbounded concurrency
- race condition
- synchronization overhead
- unnecessary context switching

---

# 44. CPU-BOUND WORK

Untuk pekerjaan CPU-intensive seperti password hashing:

Jangan membuat jumlah worker berdasarkan asumsi.

Jika CPU-bound operation menjadi bottleneck:

    Measure
      ↓
    Profile
      ↓
    Benchmark
      ↓
    Optimize

Jangan menurunkan security cost password hashing hanya untuk meningkatkan throughput.

---

# 45. I/O-BOUND WORK

Authentication request sebagian besar merupakan I/O-bound workload.

Prioritaskan:

- efficient PostgreSQL connection pool
- efficient Redis connection pool
- minimal database queries
- query indexes
- short transactions
- context timeout
- connection reuse
- minimal serialization
- minimal network round trips

daripada sekadar meningkatkan jumlah goroutine.

---

# 46. CONCURRENCY LIMIT

Jangan membatasi concurrency secara arbitrary.

Contoh buruk:

    MAX_CONCURRENT_REQUESTS = CPU_THREADS

CPU thread bukan merupakan jumlah maksimum HTTP request.

Aplikasi harus mampu menangani banyak request concurrent selama dependency seperti:

- PostgreSQL
- Redis
- memory
- CPU

masih mampu menanganinya.

Jika perlu membatasi concurrency, gunakan alasan berdasarkan resource capacity dan measurement.

---

# 47. BACKPRESSURE

Jika dependency mulai mengalami saturation:

    HTTP
      ↓
    Service
      ↓
    PostgreSQL
      ↓
    Connection Pool Saturated

jangan menyelesaikan masalah dengan membuat lebih banyak goroutine.

Gunakan:

- timeout
- rate limiting
- bounded concurrency jika diperlukan
- connection pool tuning
- backpressure

---

# 48. CPU SATURATION

AI harus membedakan:

    CPU-bound

dan:

    I/O-bound

Jika CPU utilization rendah tetapi latency tinggi:

Jangan langsung menambah CPU worker.

Cari:

- slow query
- connection pool waiting
- Redis latency
- network latency
- lock contention
- password hashing
- external service latency

Jika CPU utilization tinggi:

Lakukan profiling terlebih dahulu sebelum melakukan optimization.

---

# 49. PERFORMANCE MEASUREMENT

Performance harus diukur menggunakan:

- request latency
- throughput
- CPU utilization
- memory usage
- database latency
- Redis latency
- connection pool utilization
- goroutine count

Jangan menggunakan:

    "CPU belum 100%, berarti backend belum optimal."

CPU 100% bukan target.

Targetnya adalah:

    Lowest practical latency
    +
    High throughput
    +
    Efficient resource usage
    +
    Stable behavior under load

---

# 50. NO CPU 100% REQUIREMENT

Jangan berusaha membuat CPU selalu 100%.

CPU utilization 100% dapat meningkatkan queueing dan latency.

Backend yang sehat tidak harus menggunakan seluruh CPU pada setiap request.

Unused CPU capacity dapat menjadi headroom untuk traffic spikes.

---

# 51. LOAD TESTING

Jika performance menjadi concern, lakukan load testing.

Perhatikan:

    p50
    p95
    p99

Jangan hanya melihat:

    average response time

Contoh:

    p50 = 8ms
    p95 = 20ms
    p99 = 45ms

lebih informatif daripada hanya:

    average = 12ms

---

## GOLDEN RULE

> Let Go manage concurrency unless measurement proves that a custom concurrency model is necessary.

> Do not confuse CPU threads with HTTP worker count.

> Optimize the bottleneck, not the CPU utilization percentage.