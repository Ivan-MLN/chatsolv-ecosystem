# AGENTS.md

## Project

Production-ready authentication backend v1 built with Go, GoFiber, PostgreSQL, Redis, pgx/v5, pgxpool, and sqlc.

This file is the implementation guide for coding agents working in this repository. Keep changes simple, secure, idiomatic, and inside the v1 scope.

## Scope v1

Implement and maintain only these routes:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `GET /health`
- `GET /ready`

Do not add profile management, user CRUD, roles, permissions, admin APIs, OAuth/social login, 2FA, email verification, payments, uploads, search, or other product features unless the user explicitly changes the project scope.

There is intentionally no refresh-token HTTP endpoint in v1. Login still creates refresh-token session state in Redis.

## Architecture

Use simplified clean architecture:

```text
HTTP / GoFiber
    -> Handler
    -> Auth Service
    -> Repository or Store Interface
    -> PostgreSQL / Redis
```

Rules:

- Fiber belongs only in the HTTP delivery layer.
- Services must not accept or depend on `*fiber.Ctx`.
- Handlers parse requests, validate boundary input, call services, map errors, and write responses.
- Business rules belong in the auth service.
- SQL belongs only in `db/queries`; do not write SQL in handlers or services.
- PostgreSQL is the source of truth for users.
- Redis is only for TTL-bound temporary/distributed state.
- Prefer explicit dependencies and small functions.
- Do not introduce abstractions or packages without a concrete need.
- Do not use an ORM.
- Do not introduce global mutable state.

## Repository Layout

```text
cmd/server/                 application bootstrap and graceful shutdown
db/migrations/              PostgreSQL migrations
db/queries/                 manually authored sqlc queries
generated/sqlc/             generated database code
internal/auth/              auth models, contracts, service, handlers, adapters
internal/config/            environment configuration and validation
internal/database/          PostgreSQL and Redis client construction
internal/middleware/        request ID, logging, security, rate limiting
pkg/response/               consistent HTTP response envelopes
```

Avoid adding folders merely to imitate theoretical Clean Architecture.

## Data Model

The persistent schema contains only the `users` table:

- `id`: UUID primary key
- `name`: required, maximum 100 characters
- `email`: normalized lowercase address with a database unique constraint
- `password_hash`: Argon2id encoded hash
- `created_at`: timezone-aware timestamp
- `updated_at`: timezone-aware timestamp

Always rely on the database unique constraint to prevent duplicate-email races. Application checks alone are insufficient.

## Authentication and Security Model

### Passwords

- Hash passwords with Argon2id.
- Current parameters are 64 MiB memory, 3 iterations, parallelism 2.
- Generate a fresh cryptographically secure salt for every password.
- Never store or log plaintext passwords or password hashes.
- Never use MD5, SHA-1, or a direct SHA-256 password digest.

### Access tokens

- Use `github.com/golang-jwt/jwt/v5` with HS256.
- JWT secret comes from `JWT_SECRET` and must be at least 32 bytes.
- Required claims: `sub`, `iat`, `exp`, and `jti`.
- Keep access-token lifetime short and configurable.
- Never put passwords, secrets, or other sensitive data in JWT claims.

### Refresh tokens

- Generate opaque refresh tokens with `crypto/rand`.
- Return plaintext only to the login client.
- Store only the token hash in Redis.
- Every refresh-session key must have a configured TTL.
- Password reset must revoke the user's existing refresh sessions.

### Password reset

- Normalize email before lookup.
- Forgot-password responses must be identical for registered and unknown addresses.
- Generate reset tokens with `crypto/rand` and store only their hashes.
- Reset keys must have a TTL.
- Token consumption must be atomic and single-use.
- Invalid, expired, and reused tokens must produce the same sanitized client error.

### Development email sender

The development sender logs reset tokens so local developers can complete the flow without an email provider. It must never be enabled accidentally in production.

Configuration currently fails closed when `APP_ENV` is not `development` or `test`. Before enabling production, wire a real `EmailSender` and retain the rule that reset tokens must not appear in production logs.

### General security

- Never hardcode production secrets.
- Never log passwords, password hashes, access tokens, refresh tokens, JWT secrets, or database credentials.
- Do not return raw database, Redis, stack, or internal errors to clients.
- Keep CORS origins environment-configurable; do not default production to `*`.
- Keep request body limits and dependency timeouts bounded.
- All four auth endpoints must remain protected by Redis-backed distributed rate limiting.
- SQL must be parameterized.

## HTTP Contracts

Success envelope:

```json
{
  "success": true,
  "message": "Request successful",
  "data": {}
}
```

Error envelope:

```json
{
  "success": false,
  "message": "Something went wrong",
  "error": {
    "code": "INTERNAL_ERROR"
  }
}
```

Expected statuses:

- Register success: `201`
- Login, forgot password, reset password: `200`
- Invalid input or reset token: `400`
- Invalid credentials: `401`
- Duplicate email: `409`
- Rate limited: `429`
- Dependency readiness failure: `503`
- Unexpected internal failure: `500`

All auth request bodies must be JSON and validated at the handler boundary with `validator/v10`. Business validation still belongs in the service, and integrity constraints still belong in PostgreSQL.

## Context and Reliability

- Propagate `context.Context` from Fiber through service, PostgreSQL, Redis, and email operations.
- Use `pgxpool`; never create one PostgreSQL connection per request.
- All temporary Redis keys must expire.
- Avoid unbounded goroutines and unmanaged background work.
- Keep transactions short; do not perform external network operations inside a database transaction.
- Preserve graceful shutdown for SIGINT and SIGTERM.
- `/health` must remain lightweight.
- `/ready` may perform bounded dependency pings but no heavy operations.

## Logging

Use structured `log/slog` logging. Request logs should include relevant fields such as:

- `request_id`
- `method`
- `path`
- `status`
- `latency`
- sanitized error context where appropriate

Never add sensitive request bodies or token values to general request logs.

## SQL and Migrations

- Write explicit column lists; do not use `SELECT *`.
- Keep SQL in `db/queries` and generate Go code with sqlc.
- Add reversible up/down migrations for schema changes.
- Regenerate `generated/sqlc` after changing migrations or queries.
- Do not hand-edit generated sqlc files.

Generate code with:

```bash
make sqlc
```

## Development Commands

```bash
make run
make build
make test
make lint
make fmt
make sqlc
make migrate-up
make migrate-down
make docker-up
make docker-down
```

Environment setup:

```bash
cp .env.example .env
```

Replace `JWT_SECRET` with at least 32 random bytes before starting the server.

## Testing Rules

Use test-driven development for behavior changes:

1. Add a focused failing test.
2. Run it and verify it fails for the intended reason.
3. Add the smallest correct implementation.
4. Run the focused test.
5. Run the full suite.
6. Refactor only while tests remain green.

Required coverage areas:

- Register: success, normalization, duplicate email, secure hashing
- Login: success, unknown user, incorrect password, token/session creation
- Forgot password: known and unknown email, generic response behavior
- Reset password: valid, invalid, expired, reused token, password update, session revocation
- Redis: TTL behavior, atomic reset consumption, scoped session revocation
- PostgreSQL: migration/query correctness and duplicate constraint behavior
- Handler boundary: content type, malformed JSON, request validation, status/error mapping

PostgreSQL integration tests use `TEST_DATABASE_URL` when available and may skip cleanly when no test database or container runtime is available. Do not claim those tests ran if they were skipped.

## Required Verification

Before declaring work complete, run and fix failures from:

```bash
gofmt -w cmd internal generated pkg
go test ./... -count=1
go test -race ./internal/auth ./internal/config -count=1
go build ./...
go vet ./...
git diff --check
```

When available, also run:

```bash
golangci-lint run
govulncheck ./...
```

Do not report success based only on code inspection. Report the commands actually executed and any skipped checks or unavailable infrastructure honestly.

## Docker and Operations

- Keep the Docker image multi-stage and run the final process as a non-root user.
- Docker Compose must provide backend, PostgreSQL, and Redis services.
- Keep secrets out of Dockerfiles and committed environment files.
- Preserve configurable pool sizes, TTLs, CORS origins, rate limits, body limits, and shutdown timeouts.

## Change Policy

- Specification and security take priority over convenience.
- Correctness takes priority over performance.
- Measure before optimizing.
- Avoid N+1 queries and unnecessary database round trips, but do not add speculative caching or concurrency.
- Do not add third-party dependencies when the standard library or an existing dependency is sufficient.
- Do not leave TODOs or pseudocode for required behavior.
- If a requested feature is outside v1, explain that it is outside scope and obtain explicit confirmation before implementing it.
- If changing this architecture, first document the concrete technical reason and the simpler alternatives considered.

## Definition of Done

A change is complete only when:

- It stays within the approved v1 scope.
- Handler, service, and persistence responsibilities remain separated.
- Security-sensitive data is neither persisted nor logged improperly.
- Required Redis state has a TTL and distributed behavior remains correct.
- SQL and migrations are valid and generated code is current.
- Relevant tests exist and pass.
- Build, test, race, vet, formatting, and diff checks pass.
- Documentation and `.env.example` reflect configuration changes.
- Any unavailable Docker/PostgreSQL/lint checks are disclosed rather than fabricated.
