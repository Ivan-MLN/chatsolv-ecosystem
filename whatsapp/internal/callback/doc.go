// Package callback sends HMAC-signed HTTP callbacks to the ChatSolv backend.
// It is the bridge between whatsmeow events and the backend's internal API.
//
// Responsibilities:
//   - types.go  — shared payload structs (IncomingMessage, StatusPayload, EventPayload)
//   - client.go — HMAC-signed HTTP POST transport
//   - event.go  — whatsmeow event → backend HTTP call routing
package callback
