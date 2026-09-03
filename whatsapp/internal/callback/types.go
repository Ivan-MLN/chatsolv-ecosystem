package callback

import (
	"time"
)

// MessageContent holds the text body of an incoming message.
type MessageContent struct {
	Text string `json:"text"`
}

// IncomingMessage is the payload for POST /internal/v1/messages/incoming.
// It matches the ChatSolv backend's expected contract (PRD section 59).
type IncomingMessage struct {
	ChannelID         string         `json:"channel_id"`
	ExternalMessageID string         `json:"external_message_id"`
	ExternalUserID    string         `json:"external_user_id"`
	MessageType       string         `json:"message_type"`
	Content           MessageContent `json:"content"`
	Timestamp         time.Time      `json:"timestamp"`
}

// StatusPayload is the payload for POST /internal/v1/channels/status.
type StatusPayload struct {
	ChannelID   string `json:"channel_id"`
	Status      string `json:"status"`
	PhoneNumber string `json:"phone_number,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
}

// EventPayload is the payload for POST /internal/v1/channels/events.
type EventPayload struct {
	ChannelID   string `json:"channel_id"`
	Event       string `json:"event"`
	PhoneNumber string `json:"phone_number,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
}

// BackendResponse wraps the standard ChatSolv JSON response envelope.
type BackendResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// IncomingMessageResult is the data payload returned by /internal/v1/messages/incoming.
type IncomingMessageResult struct {
	MessageID        string  `json:"message_id"`
	ConversationID   string  `json:"conversation_id"`
	Content          string  `json:"content"`
	HandoffRequested bool    `json:"handoff_requested"`
	HandoffReason    *string `json:"handoff_reason"`
}
