# LOGGING RULES

## ChatSolv Backend

> Logging must provide enough information to understand, debug, monitor, and trace the system without leaking sensitive information or unnecessarily increasing request latency.

---

# 1. PURPOSE

Logging digunakan untuk:

- debugging
- monitoring
- incident investigation
- request tracing
- performance analysis
- security investigation
- integration troubleshooting
- error tracking

Logging BUKAN tempat untuk menyimpan seluruh data application.

Jangan menggunakan log sebagai database.

Jangan menggunakan log sebagai audit log.

Jangan menggunakan log sebagai storage conversation.

---

# 2. SOURCE OF TRUTH

Logging harus mengikuti:

```text
PRD.md
LOGGING_RULES.md
SECURITY_RULES.md
PERFORMANCE_RULES.md
AI_RULES.md
```

Jika terjadi konflik:

1. Security requirement harus diprioritaskan.
2. Jangan diam-diam mengubah logging contract.
3. Dokumentasikan technical decision jika perubahan bersifat fundamental.

---

# 3. STRUCTURED LOGGING

Gunakan structured JSON logging.

Recommended:

```text
log/slog
```

Log harus berupa structured fields, bukan string concatenation.

GOOD:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "INFO",
  "service": "chatsolv-api",
  "request_id": "req_xxx",
  "operation": "auth.login",
  "duration_ms": 42
}
```

BAD:

```text
User logged in successfully in 42ms
```

Structured logging memudahkan:

- searching
- filtering
- aggregation
- alerting
- observability
- debugging

---

# 4. REQUIRED BASE FIELDS

Application log harus menggunakan field yang konsisten.

Base fields:

```text
timestamp
level
service
environment
request_id
operation
```

Jika tersedia dan relevan:

```text
workspace_id
user_id
agent_id
conversation_id
duration_ms
error_code
```

PRD secara eksplisit menetapkan field:

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

Jangan membuat nama field yang berbeda-beda untuk konsep yang sama.

GOOD:

```text
request_id
```

BAD:

```text
requestId
req_id
requestID
correlation
request_identifier
```

Pilih satu convention dan gunakan secara konsisten.

---

# 5. REQUEST ID

Semua HTTP request harus memiliki:

```text
X-Request-ID
```

Request ID digunakan untuk:

```text
logging
tracing
Hermes request
Bot Service
webhook delivery
error tracking
```

Sesuai PRD.

---

# 6. REQUEST ID BEHAVIOR

Jika client mengirim:

```http
X-Request-ID: req_abc
```

backend dapat menggunakan ID tersebut jika formatnya valid dan sesuai security policy.

Jika tidak tersedia:

backend harus membuat request ID baru.

Request ID harus:

- unique
- safe
- bounded length
- tidak mengandung secret
- tidak mengandung password
- tidak berasal dari user-controlled value secara mentah jika dapat menyebabkan log injection

Request ID harus tersedia di request context.

---

# 7. REQUEST ID PROPAGATION

Request ID harus dapat diteruskan ke downstream service.

Contoh:

```text
Frontend
   ↓
ChatSolv Backend
   ↓
Hermes
   ↓
Obsidian
```

atau:

```text
WhatsApp Bot Service
   ↓
ChatSolv Backend
   ↓
Hermes
```

Gunakan request ID yang sama jika protocol memungkinkan.

Tujuannya agar satu request dapat dicari dari seluruh service.

---

# 8. OPERATION

Setiap log penting harus memiliki operation yang jelas.

Gunakan format:

```text
<domain>.<action>
```

Contoh:

```text
auth.register
auth.login
auth.forgot_password
auth.reset_password
auth.refresh
```

Untuk future feature:

```text
workspace.create
workspace.update
agent.provision
knowledge.create
knowledge.update
knowledge.delete
channel.connect
message.receive
message.send
webhook.deliver
```

Jangan gunakan operation generik seperti:

```text
request
handler
process
doSomething
```

jika operation yang sebenarnya dapat diketahui.

---

# 9. LOG LEVELS

Gunakan level secara konsisten.

## DEBUG

Untuk informasi development/debugging yang detail.

Tetap:

**Jangan pernah log secret atau sensitive data.**

DEBUG tidak berarti security rules boleh dilewati.

## INFO

Untuk event normal yang penting.

Contoh:

```text
server.started
server.shutdown
auth.register.success
auth.login.success
auth.password_reset.requested
```

Jangan log setiap detail kecil sebagai INFO.

## WARN

Untuk kondisi abnormal tetapi application masih dapat berjalan.

Contoh:

```text
rate_limit.near_limit
external_service.slow
redis.degraded
database_pool_near_limit
```

## ERROR

Untuk operation yang gagal dan membutuhkan investigation.

ERROR harus memiliki:

```text
error_code
operation
request_id
```

dan context relevan lainnya.

## FATAL

Hindari penggunaan FATAL di application code.

Application sebaiknya menggunakan graceful shutdown daripada memanggil fatal logging pada request path.

---

# 10. CLIENT ERRORS VS SERVER ERRORS

Tidak semua HTTP error adalah ERROR log.

Contoh:

```text
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
429 Rate Limited
```

adalah expected client/application behavior dalam banyak kasus.

Jangan memenuhi production log dengan ERROR untuk setiap invalid request.

Contoh:

```text
POST /login
wrong password
→ 401
```

Tidak harus menjadi:

```text
ERROR password incorrect
```

Gunakan level yang sesuai dan jangan membocorkan authentication information.

---

# 11. AUTHENTICATION LOGGING

Authentication adalah security-sensitive area.

Log authentication event secara hati-hati.

Contoh event:

```text
auth.register.success
auth.login.success
auth.login.failed
auth.password_reset.requested
auth.password_reset.success
auth.password_reset.failed
auth.refresh.success
auth.refresh.failed
```

Field yang boleh digunakan jika relevan:

```text
request_id
operation
user_id
duration_ms
error_code
```

Untuk user yang belum berhasil authenticated:

Jangan memaksakan `user_id` jika belum diketahui.

---

# 12. NEVER LOG PASSWORD

DILARANG log:

```text
password
password_hash
current_password
new_password
confirm_password
```

Dalam kondisi apa pun pada production.

Contoh BAD:

```go
logger.Info(
    "login",
    "email", email,
    "password", password,
)
```

Jangan dilakukan.

---

# 13. NEVER LOG TOKENS

DILARANG log:

```text
access token
JWT
refresh token
reset password token
session token
API key
service credential
internal credential
```

Jangan log token:

```text
full
partial
decoded
base64
authorization header
```

Contoh BAD:

```text
Authorization: Bearer eyJ...
```

Jangan pernah ditulis ke application log.

---

# 14. NEVER LOG SECRETS

DILARANG log:

```text
database password
Redis password
JWT secret
API secret
provider token
webhook secret
OAuth secret
encryption key
private key
service credential
```

Jika sebuah object memiliki field secret:

Jangan log object tersebut secara langsung.

Contoh BAD:

```go
logger.Info("config", "config", config)
```

jika `config` mengandung secret.

---

# 15. API KEY LOGGING

PRD menetapkan:

```text
API key
→ hash
```

bukan reversible encryption jika raw key tidak perlu diambil kembali.

Application log:

**Jangan log full API key.**

Jika diperlukan untuk debugging, gunakan identifier non-secret seperti:

```text
api_key_id
```

bukan raw API key.

---

# 16. EMAIL / PII LOGGING

Email merupakan data user.

Jangan memasukkan email ke setiap log hanya karena tersedia.

Jika email memang dibutuhkan untuk debugging authentication:

gunakan redaction atau controlled representation.

Contoh:

```text
u***@example.com
```

atau gunakan:

```text
user_id
```

setelah user berhasil di-resolve.

Prefer:

```text
user_id
```

daripada:

```text
email
```

untuk correlation internal.

---

# 17. CONVERSATION CONTENT

Jangan memasukkan full conversation content ke application log.

DILARANG secara default:

```text
customer message
agent response
full conversation
prompt
agent context
knowledge content
```

PRD secara eksplisit menyatakan:

```text
Isi conversation jangan selalu ditulis ke application log.
```

Conversation merupakan application data, bukan debugging log.

Jika debugging membutuhkan content tertentu:

gunakan controlled, redacted, temporary debugging mechanism.

Jangan menjadikan full message logging sebagai default.

---

# 18. REQUEST BODY

Jangan log request body secara default.

Terutama authentication endpoint:

```text
POST /auth/register
POST /auth/login
POST /auth/forgot-password
POST /auth/reset-password
```

Request body dapat mengandung:

```text
password
token
email
personal information
```

Jika request body memang dibutuhkan untuk debugging:

gunakan explicit field allowlist dan redaction.

Jangan melakukan:

```go
logger.Info("request", "body", c.Body())
```

secara global.

---

# 19. RESPONSE BODY

Jangan log full response body secara default.

Response dapat mengandung:

```text
tokens
user data
personal information
internal error details
```

Log metadata saja:

```text
status
duration_ms
operation
request_id
```

---

# 20. HTTP REQUEST LOGGING

Request log minimal harus dapat menjawab:

```text
Apa request-nya?
Kapan terjadi?
Berapa lama?
Apa hasilnya?
Request ID-nya apa?
Operation apa?
```

Recommended fields:

```text
timestamp
level
service
request_id
method
path
status
operation
duration_ms
```

Jika relevan:

```text
workspace_id
user_id
```

---

# 21. HTTP RESPONSE LOGGING

Gunakan satu consistent request completion log.

Contoh:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "INFO",
  "service": "chatsolv-api",
  "request_id": "req_123",
  "operation": "auth.login",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "duration_ms": 42
}
```

Jangan menghasilkan banyak log redundant untuk satu request.

---

# 22. PERFORMANCE LOGGING

Logging harus mendukung performance investigation.

Request log harus mencatat:

```text
duration_ms
```

Jika relevan, gunakan additional timing fields:

```text
db_duration_ms
redis_duration_ms
external_service_duration_ms
```

Tetapi jangan menambahkan field tersebut ke setiap log jika tidak diperlukan.

Contoh:

```text
auth.login
duration_ms = 48

db_duration_ms = 4

redis_duration_ms = 1

password_verify_duration_ms = 41
```

Ini dapat membantu mengidentifikasi bottleneck.

---

# 23. DO NOT LOG EVERYTHING

Logging lebih banyak bukan berarti observability lebih baik.

Hindari:

```text
INFO: entering handler
INFO: validating request
INFO: validation complete
INFO: calling service
INFO: service called
INFO: repository called
INFO: query started
INFO: query finished
INFO: response started
INFO: response finished
```

untuk setiap request.

Prefer:

```text
auth.login started
auth.login completed
```

ditambah ERROR/WARN ketika ada kondisi abnormal.

---

# 24. HOT PATH LOGGING

Authentication adalah hot path.

Logging pada:

```text
register
login
forgot password
reset password
```

harus ringan.

Jangan:

- serialize object besar
- serialize request body
- serialize response body
- stack trace pada setiap request
- melakukan database query hanya untuk logging
- melakukan Redis operation hanya untuk logging

Logging tidak boleh menjadi bottleneck.

---

# 25. NO DATABASE LOGGING

Jangan membuat database query khusus hanya untuk menulis application log.

BAD:

```text
HTTP request
 ↓
PostgreSQL
 ↓
Insert application_logs
```

untuk setiap request.

Gunakan structured application logging ke stdout/stderr atau logging backend.

Audit log memiliki mekanisme berbeda.

---

# 26. NO REDIS LOGGING

Jangan menggunakan Redis sebagai application log storage hanya agar log mudah diakses.

Redis adalah runtime/cache layer sesuai PRD.

Application logs harus memiliki lifecycle yang berbeda dari Redis cache.

---

# 27. LOGGING MUST NOT BLOCK REQUESTS UNNECESSARILY

Logging tidak boleh menambahkan network/database dependency ke request path hanya untuk mencatat log.

Prefer:

```text
Application
    ↓
stdout / stderr
    ↓
Log collector
    ↓
Storage / observability platform
```

daripada:

```text
Application
    ↓
HTTP logging service
    ↓
Database
    ↓
Response
```

---

# 28. ERROR LOGGING

Setiap unexpected error harus memiliki context.

Minimal:

```text
request_id
operation
error_code
```

Jika relevan:

```text
workspace_id
user_id
agent_id
conversation_id
```

Gunakan error wrapping di Go agar root cause tetap dapat ditelusuri.

Contoh:

```go
return fmt.Errorf("create user: %w", err)
```

Jangan kehilangan root error.

---

# 29. ERROR CODE

Gunakan error code yang stabil.

PRD memiliki contoh:

```text
UNAUTHORIZED
FORBIDDEN
INVALID_SESSION_TOKEN
RATE_LIMIT_EXCEEDED
INTERNAL_ERROR
```

Untuk authentication:

gunakan code yang konsisten dengan API error contract.

Jangan menggunakan error message sebagai identifier.

BAD:

```text
"password salah"
```

sebagai machine-readable identifier.

GOOD:

```text
AUTHENTICATION_FAILED
```

---

# 30. INTERNAL ERROR DETAILS

Jangan expose internal error detail ke client.

Log internal error:

```text
database connection refused
```

Client menerima:

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Internal server error."
  },
  "request_id": "req_xxx"
}
```

Request ID membantu support/debugging tanpa membocorkan internal implementation.

---

# 31. STACK TRACE

Stack trace tidak perlu dicetak pada setiap error.

Gunakan stack trace terutama untuk:

```text
unexpected panic
unexpected internal failure
critical error
```

Jangan menghasilkan stack trace untuk expected validation errors.

---

# 32. PANIC / RECOVER

Application harus memiliki recover middleware.

Jika panic terjadi:

1. Recover panic.
2. Generate/retain request ID.
3. Log panic sebagai ERROR.
4. Sertakan operation jika diketahui.
5. Sertakan stack trace.
6. Return safe `500 Internal Server Error`.

Jangan mengembalikan stack trace kepada client.

---

# 33. PANIC LOGGING

Contoh informasi yang relevan:

```text
request_id
operation
method
path
status
panic
stack
```

Jangan memasukkan request body atau authentication headers.

---

# 34. LOG INJECTION PROTECTION

User-controlled strings harus diperlakukan sebagai untrusted data.

Jangan membuat plain text log dengan concatenation:

```go
logger.Info("user=" + userInput)
```

Gunakan structured logging:

```go
logger.Info("user operation", "value", userInput)
```

dan gunakan logger/output yang menangani encoding dengan benar.

Jangan mengizinkan newline/control characters dari input user merusak struktur log.

---

# 35. LOG FIELD CARDINALITY

Hindari field dengan cardinality yang tidak terkendali jika logging backend akan melakukan indexing.

Contoh yang harus dipertimbangkan:

```text
full message
full URL with query parameters
arbitrary user input
large JSON
```

Gunakan identifier:

```text
request_id
user_id
workspace_id
conversation_id
```

sesuai kebutuhan.

---

# 36. URL QUERY PARAMETERS

Jangan log query parameters secara mentah.

Query parameter dapat mengandung:

```text
token
api_key
email
search query
personal data
```

Jika diperlukan:

gunakan allowlist parameter yang aman.

---

# 37. AUTHORIZATION HEADER

DILARANG log:

```http
Authorization
Cookie
Set-Cookie
```

secara mentah.

Terutama:

```text
Bearer JWT
refresh token
session token
```

---

# 38. COOKIE LOGGING

Jangan log cookie value.

Jika debugging session membutuhkan informasi:

gunakan identifier non-secret.

Contoh:

```text
session_id
```

jika memang aman dan bukan raw session secret.

---

# 39. DATABASE LOGGING

Application log boleh mencatat:

```text
database operation
duration
error code
```

tetapi jangan mencatat:

```text
database password
connection string
credentials
full SQL parameters
sensitive query values
```

SQL query logging harus digunakan secara selective.

Jangan mengaktifkan verbose SQL logging di production tanpa alasan.

---

# 40. REDIS LOGGING

Boleh mencatat:

```text
operation
duration
key category
error
```

Jangan mencatat:

```text
secret value
refresh token
reset token
full sensitive payload
```

Jika Redis key mengandung sensitive identifier:

gunakan redaction atau safe identifier.

---

# 41. EXTERNAL SERVICE LOGGING

Untuk external service:

log metadata seperti:

```text
service
operation
duration_ms
status
error_code
request_id
```

Contoh:

```text
hermes.generate_response
duration_ms=842
status=success
```

Jangan log:

```text
provider credential
full authorization header
secret
unnecessary full payload
```

---

# 42. INTERNAL SERVICE LOGGING

Internal services seperti WhatsApp Bot Service harus dapat di-correlate menggunakan:

```text
request_id
```

Internal authentication:

```text
service credential
HMAC
mTLS
```

credential tersebut tidak boleh masuk application log.

---

# 43. AUDIT LOG ≠ APPLICATION LOG

Jangan mencampur:

```text
Application Log
```

dengan:

```text
Audit Log
```

Application log digunakan untuk:

```text
debugging
monitoring
performance
errors
operations
```

Audit log digunakan untuk:

```text
security-sensitive business actions
accountability
compliance
user activity history
```

---

# 44. AUDIT LOG

PRD menetapkan audit event seperti:

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

Future implementation harus menggunakan audit log terpisah.

Jangan menganggap:

```go
logger.Info("api key revoked")
```

berarti audit trail sudah selesai.

---

# 45. AUDIT LOG SHOULD BE STRUCTURED

Audit event minimal harus memiliki konsep:

```text
event
timestamp
actor
workspace
resource
action
request_id
```

Contoh konsep:

```json
{
  "event": "api_key.revoked",
  "request_id": "req_xxx",
  "workspace_id": "wsp_xxx",
  "user_id": "usr_xxx",
  "resource_id": "key_xxx"
}
```

Jangan memasukkan secret API key.

---

# 46. AUTHENTICATION AUDIT EVENTS

Untuk authentication, pertimbangkan audit event:

```text
user.registered
user.login
user.login_failed
password_reset.requested
password_reset.completed
refresh_token.rotated
session.revoked
```

Tidak semua event harus menjadi database audit record jika requirement belum membutuhkannya.

Bedakan kebutuhan:

```text
Application Log
```

vs

```text
Security Audit Event
```

---

# 47. RATE LIMIT LOGGING

PRD menggunakan Redis-backed rate limiting.

Jika rate limit terlampaui:

```text
HTTP 429
```

Log dapat menggunakan:

```text
WARN
```

dengan:

```text
operation
request_id
identifier_type
error_code=RATE_LIMIT_EXCEEDED
```

Jangan log raw sensitive identifier jika tidak diperlukan.

---

# 48. HEALTH CHECK LOGGING

Jangan menghasilkan INFO log untuk setiap:

```text
GET /health
GET /ready
```

jika endpoint dipanggil sangat sering oleh orchestrator/load balancer.

Health check normal sebaiknya tidak memenuhi log.

Log hanya jika:

```text
dependency degraded
readiness failed
unexpected error
```

atau gunakan dedicated metrics.

---

# 49. STARTUP LOGGING

Saat application startup:

INFO dapat mencatat:

```text
service
environment
version
build
server address
```

Jangan mencatat:

```text
database password
JWT secret
Redis password
API keys
provider credentials
```

Contoh:

```json
{
  "level": "INFO",
  "service": "chatsolv-api",
  "operation": "server.start",
  "version": "1.0.0",
  "environment": "production"
}
```

---

# 50. SHUTDOWN LOGGING

Graceful shutdown harus mencatat:

```text
server.shutdown.started
server.shutdown.completed
```

Jika shutdown disebabkan error:

sertakan error context.

---

# 51. ENVIRONMENT

Logging behavior dapat berbeda berdasarkan environment.

## Development

Boleh lebih verbose.

Contoh:

```text
DEBUG
INFO
WARN
ERROR
```

## Production

Prefer:

```text
INFO
WARN
ERROR
```

DEBUG harus disabled atau controlled.

Namun:

**Security rules berlaku di semua environment.**

Jangan log password di development dengan alasan:

> "Kan cuma local."

---

# 52. DEVELOPMENT LOGGING

Development logging boleh memberikan context tambahan, tetapi tetap tidak boleh mencetak:

```text
password
JWT
refresh token
reset token
API key
secret
private key
```

---

# 53. PRODUCTION LOGGING

Production log harus:

- structured
- machine-readable
- searchable
- safe
- concise
- correlated
- low overhead

Jangan mengaktifkan verbose debug logging production tanpa alasan yang jelas.

---

# 54. LOG RETENTION

Application log retention harus ditentukan oleh infrastructure/operations layer.

Jangan menyimpan log selamanya tanpa alasan.

Pertimbangkan:

```text
storage cost
privacy
security
compliance
debugging needs
```

Application code tidak boleh mengasumsikan log tersedia selamanya.

---

# 55. LOG ROTATION

Jika application menulis ke local file:

pastikan log rotation dan retention ditangani dengan benar.

Namun untuk containerized deployment, prefer:

```text
stdout/stderr
```

dan biarkan infrastructure logging system menangani:

```text
collection
rotation
retention
aggregation
```

Jangan membuat custom logging infrastructure yang kompleks tanpa kebutuhan.

---

# 56. LOGGING OUTPUT

Production deployment sebaiknya menggunakan:

```text
stdout
stderr
```

dengan structured JSON.

Contoh pipeline:

```text
Go Application
      ↓
stdout / stderr
      ↓
Container Runtime
      ↓
Log Collector
      ↓
Observability Platform
```

---

# 57. TIME FORMAT

Gunakan timestamp yang konsisten.

Prefer:

```text
RFC3339 / RFC3339Nano
```

dan gunakan UTC.

Contoh:

```text
2026-08-25T12:00:00Z
```

Jangan mencampurkan timezone tanpa field yang jelas.

---

# 58. DURATION

Gunakan:

```text
duration_ms
```

untuk request/operation duration yang perlu dipantau.

Contoh:

```json
{
  "operation": "auth.login",
  "duration_ms": 42
}
```

---

# 59. LOGGING AND PERFORMANCE

Logging harus mengikuti `PERFORMANCE_RULES.md`.

Jangan melakukan optimization seperti:

```text
disable all logs
```

hanya demi latency.

Sebaliknya:

```text
Useful logs
+
Structured logs
+
Low allocation
+
No unnecessary serialization
+
No network call per log
```

adalah target.

---

# 60. NO EXPENSIVE LOG ARGUMENTS UNNECESSARILY

Hindari membuat object besar hanya untuk logging.

BAD:

```go
logger.Debug(
    "request",
    "full_response",
    expensiveSerialize(response),
)
```

jika DEBUG bahkan tidak aktif.

Gunakan structured logging dengan cara yang tidak melakukan expensive work sebelum logger memutuskan level tersebut.

---

# 61. NO DUPLICATE ERROR LOGGING

Satu error tidak perlu dicatat berulang kali di setiap layer.

BAD:

```text
Repository:
ERROR database failed

Service:
ERROR repository failed

Handler:
ERROR service failed

Middleware:
ERROR request failed
```

Prefer:

```text
Repository
→ wrap error

Service
→ return error

Handler / boundary
→ log final error
```

atau gunakan strategy yang konsisten.

---

# 62. ERROR OWNERSHIP

Tentukan boundary yang bertanggung jawab mencatat error.

Secara umum:

```text
Low-level layer
→ create/wrap error

Application boundary
→ log unexpected error

HTTP layer
→ translate error to API response
```

Jangan setiap layer melakukan logging terhadap error yang sama.

---

# 63. EXPECTED ERRORS

Expected errors seperti:

```text
invalid credentials
validation failed
resource not found
rate limit exceeded
```

tidak selalu membutuhkan ERROR level.

Gunakan level yang sesuai.

Tujuan:

```text
ERROR = sesuatu yang perlu investigation
```

bukan:

```text
ERROR = semua request yang tidak 2xx
```

---

# 64. SECURITY EVENTS

Security-sensitive events harus dapat diidentifikasi.

Contoh:

```text
auth.login_failed
auth.password_reset.requested
auth.password_reset.failed
auth.refresh.failed
rate_limit.exceeded
invalid_service_credential
```

Tetap jangan log credential yang digunakan.

---

# 65. FAILED LOGIN

Jangan membocorkan apakah email/user tertentu ada.

BAD:

```text
user_exists=false
```

atau:

```text
email_not_registered
```

jika informasi tersebut dapat membantu user enumeration.

Gunakan generic event:

```text
auth.login_failed
```

dengan safe metadata.

---

# 66. PASSWORD RESET

Password reset merupakan security-sensitive flow.

Log event seperti:

```text
auth.password_reset.requested
auth.password_reset.completed
auth.password_reset.failed
```

Tetapi jangan log:

```text
reset token
reset URL
password
```

Jika reset request menggunakan email:

prefer user identifier yang sudah diketahui.

Jika belum ada user:

jangan mencatat sensitive lookup details secara berlebihan.

---

# 67. TOKEN ROTATION

PRD menggunakan refresh token rotation.

Log:

```text
auth.refresh.success
auth.refresh.failed
auth.refresh.rotated
```

Jangan log token lama atau token baru.

---

# 68. REQUEST CONTEXT

Request-scoped metadata harus tersedia melalui context.

Contoh:

```text
request_id
user_id
workspace_id
operation
```

Handler/service/repository tidak perlu membuat request ID baru.

Gunakan request context yang sama.

---

# 69. LOGGER DEPENDENCY FLOW

Logging dependency harus mengikuti architecture.

Contoh:

```text
HTTP
 ↓
Handler
 ↓
Service
 ↓
Repository
```

Logger dapat digunakan oleh layer yang membutuhkan logging.

Namun jangan membuat setiap layer mencetak log yang sama.

Gunakan dependency injection atau context-based logger secara konsisten.

---

# 70. LOGGING MUST NOT CHANGE BUSINESS BEHAVIOR

Logging failure tidak boleh membuat business operation gagal kecuali logging infrastructure memang merupakan explicit critical dependency, yang bukan default untuk application logging.

Contoh:

```text
Database berhasil menyimpan user
+
Logger gagal menulis
=
User tidak boleh dianggap gagal dibuat
```

Application log harus best-effort.

---

# 71. OBSERVABILITY CORRELATION

Logging harus siap dikembangkan ke:

```text
metrics
tracing
error tracking
```

Gunakan:

```text
request_id
operation
resource IDs
duration
```

sebagai correlation foundation.

Jangan membangun distributed tracing sendiri hanya untuk memenuhi logging requirement.

---

# 72. METRICS VS LOGS

Gunakan metrics untuk data yang sering dihitung.

Contoh:

```text
request count
request latency
error rate
rate limit count
database pool utilization
```

Gunakan logs untuk:

```text
individual event
unexpected error
debugging context
security event
```

Jangan menggunakan log sebagai pengganti metrics.

---

# 73. HIGH-VOLUME EVENTS

Event dengan volume sangat tinggi tidak boleh menghasilkan unlimited verbose logs.

Contoh:

```text
health check
public API requests
incoming messages
agent runtime events
```

Jika volume meningkat:

gunakan:

```text
sampling
aggregation
metrics
```

daripada mencetak semua detail.

---

# 74. LOG SAMPLING

Sampling hanya boleh digunakan untuk high-volume informational logs.

Jangan melakukan sampling terhadap critical security events atau critical errors tanpa explicit policy.

Contoh yang dapat di-sample:

```text
high-volume successful requests
```

Contoh yang tidak boleh sembarangan di-sample:

```text
authentication security failures
critical errors
audit events
```

---

# 75. RESOURCE IDs

Gunakan public resource IDs yang sesuai dengan PRD jika tersedia:

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

Gunakan resource ID daripada raw internal database information jika cukup untuk correlation.

---

# 76. TENANT CONTEXT

ChatSolv adalah multi-tenant.

Jika operation berhubungan dengan tenant:

gunakan:

```text
workspace_id
```

jika sudah diketahui.

Jangan log raw tenant path.

Jangan log:

```text
vault filesystem path
```

PRD secara eksplisit menjaga vault path tetap internal.

---

# 77. TENANT ISOLATION

Jangan membocorkan data tenant lain melalui logs.

Contoh:

Jika request dari:

```text
workspace A
```

gagal mencoba mengakses:

```text
workspace B
```

log harus tetap aman dan tidak mencetak sensitive data milik workspace B.

---

# 78. FILE PATH LOGGING

Jangan log raw filesystem path jika path dapat:

- membocorkan infrastructure
- membocorkan tenant location
- membocorkan secret
- membantu path traversal investigation secara berlebihan

Jika diperlukan:

gunakan logical resource ID.

Contoh:

```text
second_brain_id
knowledge_id
workspace_id
```

bukan absolute filesystem path.

---

# 79. EXTERNAL PROVIDER CREDENTIALS

Provider credentials seperti:

```text
WhatsApp credential
Hermes credential
OAuth credential
webhook secret
API provider token
```

tidak boleh muncul dalam logs.

Jika external request gagal:

log:

```text
provider
operation
status
duration_ms
error_code
request_id
```

bukan credential.

---

# 80. WEBHOOK LOGGING

Webhook delivery dapat menggunakan:

```text
event_id
webhook_id
request_id
attempt
status
duration_ms
```

PRD menyimpan webhook delivery metadata.

Jangan log:

```text
webhook secret
authorization header
full sensitive payload
```

---

# 81. BACKGROUND JOB LOGGING

Background jobs harus memiliki correlation ID/job ID.

Contoh:

```text
job_id
job_type
operation
workspace_id
duration_ms
attempt
status
error_code
```

Job examples:

```text
workspace.provision
agent.provision
knowledge.ingest
knowledge.update
knowledge.delete
webhook.deliver
```

---

# 82. JOB RETRY LOGGING

Untuk retry:

gunakan fields:

```text
job_id
operation
attempt
max_attempts
error_code
next_retry_at
```

Jangan mencetak seluruh job payload jika tidak diperlukan.

---

# 83. LOG FORMAT CONSISTENCY

Semua service ChatSolv harus menggunakan naming convention yang konsisten.

Contoh:

```text
request_id
workspace_id
user_id
agent_id
conversation_id
operation
duration_ms
error_code
```

Jangan membuat service A menggunakan:

```text
workspaceId
```

dan service B:

```text
workspace_id
```

---

# 84. EXAMPLE — SUCCESSFUL LOGIN

Safe example:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "INFO",
  "service": "chatsolv-api",
  "environment": "production",
  "request_id": "req_01KXXX",
  "operation": "auth.login",
  "user_id": "usr_01KXXX",
  "duration_ms": 43,
  "status": 200
}
```

Tidak ada:

```text
password
JWT
refresh_token
email
authorization header
```

---

# 85. EXAMPLE — FAILED LOGIN

Safe:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "WARN",
  "service": "chatsolv-api",
  "environment": "production",
  "request_id": "req_01KXXX",
  "operation": "auth.login",
  "error_code": "AUTHENTICATION_FAILED",
  "status": 401,
  "duration_ms": 41
}
```

Jangan:

```json
{
  "email": "user@example.com",
  "password": "secret123",
  "reason": "email exists but password is wrong"
}
```

---

# 86. EXAMPLE — INTERNAL ERROR

Safe:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "ERROR",
  "service": "chatsolv-api",
  "environment": "production",
  "request_id": "req_01KXXX",
  "operation": "auth.register",
  "error_code": "INTERNAL_ERROR",
  "duration_ms": 18
}
```

Internal developer log may include the wrapped root cause if safe.

Client tetap menerima generic error.

---

# 87. EXAMPLE — RATE LIMIT

Safe:

```json
{
  "timestamp": "2026-08-25T12:00:00Z",
  "level": "WARN",
  "service": "chatsolv-api",
  "request_id": "req_01KXXX",
  "operation": "auth.login",
  "error_code": "RATE_LIMIT_EXCEEDED",
  "status": 429
}
```

Jangan log raw authentication credential.

---

# 88. LOGGING CHECKLIST

Sebelum feature dianggap selesai:

### Structure

- [ ] JSON structured logging
- [ ] Consistent field names
- [ ] Timestamp UTC
- [ ] Level defined
- [ ] Service defined
- [ ] Operation defined

### Request

- [ ] X-Request-ID supported
- [ ] Request ID propagated
- [ ] Request duration available
- [ ] HTTP status available

### Security

- [ ] No password
- [ ] No password hash
- [ ] No JWT
- [ ] No refresh token
- [ ] No reset token
- [ ] No API key
- [ ] No provider secret
- [ ] No authorization header
- [ ] No cookie value
- [ ] No sensitive request body
- [ ] No unnecessary PII

### Performance

- [ ] No DB query for logging
- [ ] No Redis operation for logging
- [ ] No external HTTP request for logging
- [ ] No expensive serialization unnecessarily
- [ ] No excessive logs on hot path
- [ ] No duplicate error logs

### Observability

- [ ] Error code available
- [ ] Resource IDs available where relevant
- [ ] Workspace context available where relevant
- [ ] User context available where relevant
- [ ] Duration available
- [ ] Logs can be correlated using request_id

---

# 89. FINAL RULE

Logging harus menjawab:

```text
What happened?
When did it happen?
Where did it happen?
Which request caused it?
Which operation was running?
How long did it take?
Which resource was affected?
Why did it fail?
```

tanpa menjawabnya dengan:

```text
password
token
secret
private data
full conversation
```

---

# GOLDEN RULE

> **Log events, not secrets.**

> **Log metadata, not entire application state.**

> **Log enough to debug, but never enough to compromise security.**

> **Application logging must remain cheap enough for the hot path.**

> **Application logs are not audit logs and not a database.**

> **Every important request must be traceable through request_id.**
