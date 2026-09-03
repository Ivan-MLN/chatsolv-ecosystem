package server

import (
	"chatsolv-whatsapp/internal/whatsapp"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// SessionManager is the interface Handler uses to control WhatsApp sessions.
// It is satisfied by *whatsapp.Manager.
type SessionManager interface {
	Connect(ctx context.Context, channelID string, phoneNumber string) (ConnectResult, error)
	Disconnect(channelID string) error
	IsConnected(channelID string) bool
	GetProfile(ctx context.Context, channelID string) (whatsapp.WhatsAppProfile, error)
	SendTextMessage(ctx context.Context, channelID string, recipient string, text string) error
}

// ConnectResult mirrors whatsapp.ConnectResult without creating an import cycle.
type ConnectResult struct {
	SessionID   string
	Status      string
	QR          string
	PairingCode string
}

// Handler exposes the internal WhatsApp control API.
type Handler struct {
	mgr SessionManager
	log *slog.Logger
}

// NewHandler creates a Handler.
func NewHandler(mgr SessionManager, log *slog.Logger) *Handler {
	return &Handler{mgr: mgr, log: log}
}

// Connect handles POST /internal/v1/channels/connect
func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID   string `json:"channel_id"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ChannelID) == "" {
		writeError(w, http.StatusBadRequest, "channel_id is required")
		return
	}

	result, err := h.mgr.Connect(r.Context(), req.ChannelID, strings.TrimSpace(req.PhoneNumber))
	if err != nil {
		h.log.Error("connect failed", "channel_id", req.ChannelID, "error", err)
		writeError(w, http.StatusInternalServerError, "connect failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"session_id":   result.SessionID,
			"status":       result.Status,
			"qr":           result.QR,
			"pairing_code": result.PairingCode,
		},
	})
}

// Disconnect handles POST /internal/v1/channels/disconnect
func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID string `json:"channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ChannelID) == "" {
		writeError(w, http.StatusBadRequest, "channel_id is required")
		return
	}

	if err := h.mgr.Disconnect(req.ChannelID); err != nil {
		h.log.Error("disconnect failed", "channel_id", req.ChannelID, "error", err)
		writeError(w, http.StatusInternalServerError, "disconnect failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]string{"channel_id": req.ChannelID, "status": "disconnected"},
	})
}

// Status handles GET /internal/v1/channels/status?channel_id=xxx
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "channel_id is required")
		return
	}

	status := "disconnected"
	if h.mgr.IsConnected(channelID) {
		status = "connected"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]string{"channel_id": channelID, "status": status},
	})
}

// Profile handles GET /internal/v1/channels/profile?channel_id=xxx
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "channel_id is required")
		return
	}

	profile, err := h.mgr.GetProfile(r.Context(), channelID)
	if err != nil {
		h.log.Warn("get profile failed", "channel_id", channelID, "error", err)
		writeError(w, http.StatusNotFound, "channel profile not available: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    profile,
	})
}

// SendMessage handles POST /internal/v1/messages/send
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID string `json:"channel_id"`
		Recipient string `json:"recipient"`
		Text      string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ChannelID) == "" || strings.TrimSpace(req.Recipient) == "" || strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "channel_id, recipient, and text are required")
		return
	}

	if err := h.mgr.SendTextMessage(r.Context(), req.ChannelID, strings.TrimSpace(req.Recipient), req.Text); err != nil {
		h.log.Error("send message failed", "channel_id", req.ChannelID, "recipient", req.Recipient, "error", err)
		writeError(w, http.StatusInternalServerError, "send message failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]string{
			"channel_id": req.ChannelID,
			"recipient":  req.Recipient,
			"status":     "sent",
		},
	})
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "status": "ok"})
}
