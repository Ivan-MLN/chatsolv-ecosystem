# ChatSolv WhatsApp Service

Standalone Go service managing WhatsApp Web socket connections using whatsmeow. Provides an HMAC-authenticated HTTP control API for the ChatSolv backend and dispatches incoming messages/events back via HMAC-signed HTTP webhooks.

## Development & Environment

- Go version: 1.27.0
- Config: Copy `.env.example` to `.env`
- Required variables: `INTERNAL_SERVICE_SECRET` (min 32 bytes, shared with backend), `BACKEND_URL` (default `http://localhost:3000`), `PORT` (default `4010`), `DB_ROOT` (default `./data/sessions`)

## Commands

- Run: `make run` (or `go run ./cmd/server`)
- Build: `make build` (outputs to `bin/whatsapp`)
- Test: `make test` (runs `go test ./... -count=1`)
- Format: `make fmt` (runs `gofmt -w cmd internal`)
- Vet: `make vet` (runs `go vet ./...`)
- Tidy dependencies: `make tidy` (runs `go mod tidy`)

## Architecture & Code Conventions

- Directory layout:
  - `cmd/server/main.go`: Entrypoint, dynamic WhatsApp Web version sync, signal handling, server lifecycle.
  - `internal/config/`: Env loading and validation (`config.go`).
  - `internal/server/`: HTTP router (`net/http`), HMAC middleware, status recording, and handler endpoints.
  - `internal/whatsapp/`: whatsmeow manager, session lifecycle (`Session`), profile fetching, and modernc SQLite store initialization.
  - `internal/callback/`: Webhook client sending HMAC-signed events to backend, incoming message parsing, attachment downloading (`/tmp/chatsolv-media`).
- Interface decoupling: `server.SessionManager` and `callback.MessageSender` are satisfied by `whatsapp.Manager` via adapters in `cmd/server/main.go` to prevent package cycle imports.
- Session Context: Use `detachedSessionContext()` when spawning WhatsApp socket connections from HTTP requests so background sockets survive request completion.
- Logging: Structured JSON logs using stdlib `log/slog`.
- Filtering: Incoming messages strictly process 1-on-1 private chats (DMs); group chats (`@g.us`), broadcasts, and newsletters are discarded.

## Pitfalls & Security

- HMAC Authentication: All `/internal/v1/*` routes require `X-ChatSolv-Timestamp` (RFC3339) and `X-ChatSolv-Signature` (`hex(HMAC-SHA256(secret, timestamp + "." + body))`) with a 5-minute replay window tolerance. Only `/health` is unauthenticated.
- Secret Length: `INTERNAL_SERVICE_SECRET` shorter than 32 bytes causes immediate fatal error on startup.
- Multi-session Lock: Each channel has its own SQLite file in `DB_ROOT/<channel_id>.db`. Re-connecting an existing active channel replaces and disconnects the previous socket.
