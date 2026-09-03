// Package whatsapp provides WhatsApp session management using the whatsmeow library.
// It is organized into three sub-concerns:
//   - store:   opening and upgrading per-channel SQLite stores
//   - session: the Session type and its lifecycle (connect, disconnect)
//   - manager: the Manager that owns all sessions and fans out events
package whatsapp
