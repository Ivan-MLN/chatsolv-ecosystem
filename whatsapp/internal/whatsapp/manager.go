package whatsapp

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// EventHandler is called for every relevant whatsmeow event.
// channelID identifies which session produced the event.
type EventHandler func(channelID string, evt interface{})

// Manager owns all active WhatsApp sessions.
// It is safe for concurrent use.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	dbRoot   string
	log      *slog.Logger
	handler  EventHandler
}

// NewManager creates a Manager.
// dbRoot is the directory where per-channel SQLite databases will be stored.
func NewManager(dbRoot string, log *slog.Logger, handler EventHandler) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		dbRoot:   dbRoot,
		log:      log,
		handler:  handler,
	}
}

// Connect starts (or restarts) a WhatsApp session for channelID.
// It returns the initial connection result synchronously; further events
// (status changes, incoming messages, QR refreshes) are delivered via
// the registered EventHandler.
func (m *Manager) Connect(ctx context.Context, channelID string, phoneNumber string) (ConnectResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Tear down any existing session first.
	if existing, ok := m.sessions[channelID]; ok {
		existing.Disconnect()
		delete(m.sessions, channelID)
	}

	sess, result, err := newSession(ctx, channelID, phoneNumber, m.dbRoot, func(evt interface{}) {
		m.onEvent(channelID, evt)
	})
	if err != nil {
		return ConnectResult{}, err
	}

	m.sessions[channelID] = sess
	return result, nil
}

// Disconnect tears down the session for channelID.
// Returns nil if the session does not exist (idempotent).
func (m *Manager) Disconnect(channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[channelID]
	if !ok {
		return nil
	}
	sess.Disconnect()
	delete(m.sessions, channelID)
	return nil
}

// IsConnected reports whether the session for channelID is currently connected.
func (m *Manager) IsConnected(channelID string) bool {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	return ok && sess.IsConnected()
}

// GetProfile returns the profile info for a channel session.
func (m *Manager) GetProfile(ctx context.Context, channelID string) (WhatsAppProfile, error) {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		// Try auto-resuming session from disk if paired
		m.mu.Lock()
		if existing, ok2 := m.sessions[channelID]; ok2 {
			sess = existing
		} else {
			var err error
			sess, _, err = newSession(ctx, channelID, "", m.dbRoot, func(evt interface{}) {
				m.onEvent(channelID, evt)
			})
			if err != nil {
				m.mu.Unlock()
				return WhatsAppProfile{}, &ErrSessionNotFound{ChannelID: channelID}
			}
			m.sessions[channelID] = sess
		}
		m.mu.Unlock()
	}
	return sess.GetProfile(ctx)
}
func (m *Manager) ResolvePhoneNumber(ctx context.Context, channelID string, jid types.JID) types.JID {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		return jid
	}
	return sess.ResolvePhoneNumber(ctx, jid)
}
func (m *Manager) DownloadAttachment(ctx context.Context, channelID string, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		return nil, &ErrSessionNotFound{ChannelID: channelID}
	}
	return sess.DownloadAttachment(ctx, msg)
}

// SendPresence sends typing (composing) or paused status to a WhatsApp chat.
func (m *Manager) SendPresence(ctx context.Context, channelID string, jid types.JID, state types.ChatPresence) error {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		return &ErrSessionNotFound{ChannelID: channelID}
	}
	return sess.SendPresence(ctx, jid, state)
}

// If quoted parameters are provided, it sends the message as a reply/quote.
func (m *Manager) SendText(ctx context.Context, channelID string, jid types.JID, text string, quotedMsgID string, quotedParticipant string, quotedMsg *waE2E.Message) error {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		return &ErrSessionNotFound{ChannelID: channelID}
	}
	return sess.SendText(ctx, jid, text, quotedMsgID, quotedParticipant, quotedMsg)
}

// SendTextMessage sends a plain-text message to a phone number or JID string.
func (m *Manager) SendTextMessage(ctx context.Context, channelID string, recipient string, text string) error {
	m.mu.RLock()
	sess, ok := m.sessions[channelID]
	m.mu.RUnlock()
	if !ok {
		return &ErrSessionNotFound{ChannelID: channelID}
	}
	jid := types.NewJID(strings.TrimSuffix(recipient, "@s.whatsapp.net"), types.DefaultUserServer)
	return sess.SendText(ctx, jid, text, "", "", nil)
}

// DisconnectAll tears down every active session. Used during graceful shutdown.
func (m *Manager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, sess := range m.sessions {
		m.log.Info("disconnecting session on shutdown", "channel_id", id)
		sess.Disconnect()
		delete(m.sessions, id)
	}
}

// onEvent is called by each session's whatsmeow event handler.
// It runs in whatsmeow's internal goroutine — never hold m.mu here.
func (m *Manager) onEvent(channelID string, evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		m.log.Debug("incoming message", "channel_id", channelID, "from", v.Info.Sender)
	case *events.Connected:
		m.log.Info("session connected", "channel_id", channelID)
	case *events.Disconnected:
		m.log.Info("session disconnected", "channel_id", channelID)
	case *events.LoggedOut:
		m.log.Warn("session logged out", "channel_id", channelID)
		// Remove the session without holding m.mu (avoid deadlock with Connect).
		// Use a goroutine so the event handler returns promptly.
		go func() {
			m.mu.Lock()
			sess := m.sessions[channelID]
			delete(m.sessions, channelID)
			m.mu.Unlock()
			if sess != nil {
				sess.Disconnect()
			}
		}()
	case *events.PairSuccess:
		m.log.Info("pairing successful", "channel_id", channelID, "jid", v.ID)
	}

	// Forward every event to the registered handler (e.g. callback client).
	m.handler(channelID, evt)
}

// ErrSessionNotFound is returned when an operation is requested for a
// channel that has no active session.
type ErrSessionNotFound struct {
	ChannelID string
}

func (e *ErrSessionNotFound) Error() string {
	return "no active session for channel: " + e.ChannelID
}
