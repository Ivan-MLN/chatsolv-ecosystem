package internalapi

import (
	"context"
	"errors"
	"strings"

	"authbackend/internal/conversation"
)

var ErrInvalidInput = errors.New("invalid internal API input")

type MessageContent struct {
	Text string `json:"text"`
}
type IncomingMessage struct {
	ChannelID         string         `json:"channel_id"`
	ExternalMessageID string         `json:"external_message_id"`
	ExternalUserID    string         `json:"external_user_id"`
	MessageType       string         `json:"message_type"`
	Content           MessageContent `json:"content"`
}
type ChannelStatusInput struct {
	ChannelID   string `json:"channel_id"`
	Status      string `json:"status"`
	PhoneNumber string `json:"phone_number"`
	SessionID   string `json:"session_id"`
}
type ChannelEventInput struct {
	ChannelID   string `json:"channel_id"`
	Event       string `json:"event"`
	PhoneNumber string `json:"phone_number"`
	SessionID   string `json:"session_id"`
}
type RespondInput struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}
type RuntimeContext struct{ ChannelID, ChannelType, ExternalUserID string }
type AgentHealth struct {
	AgentID     string `json:"agent_id"`
	Status      string `json:"status"`
	BrainStatus string `json:"brain_status"`
	Ready       bool   `json:"ready"`
}
type Repository interface {
	UpdateChannelStatus(context.Context, string, string, string, string) error
	ResolveRuntimeContext(context.Context, string, string) (RuntimeContext, error)
	GetAgentHealth(context.Context, string) (AgentHealth, error)
}
type ConversationService interface {
	Handle(context.Context, conversation.Incoming) (conversation.Result, error)
}

type HandoffCommandService interface {
	AcceptByCommand(ctx context.Context, channelID, senderPhone, shortCode string) (string, error)
	ResolveByCommand(ctx context.Context, channelID, senderPhone, shortCode string) (string, error)
}

type Service struct {
	repo          Repository
	conversations ConversationService
	handoffs      HandoffCommandService
}

func NewService(repo Repository, conversations ConversationService) *Service {
	return &Service{repo: repo, conversations: conversations}
}

func (s *Service) SetHandoffs(h HandoffCommandService) {
	s.handoffs = h
}

func (s *Service) Incoming(ctx context.Context, in IncomingMessage) (conversation.Result, error) {
	if in.ChannelID == "" || in.ExternalMessageID == "" || in.ExternalUserID == "" || in.MessageType != "text" || strings.TrimSpace(in.Content.Text) == "" {
		return conversation.Result{}, ErrInvalidInput
	}

	trimmedText := strings.TrimSpace(in.Content.Text)

	// Intercept Admin WhatsApp Commands: /acc and /done
	if strings.HasPrefix(trimmedText, "/acc") || strings.HasPrefix(trimmedText, "/ACC") {
		if s.handoffs != nil {
			parts := strings.Fields(trimmedText)
			code := ""
			if len(parts) > 1 {
				code = parts[1]
			}
			replyText, _ := s.handoffs.AcceptByCommand(ctx, in.ChannelID, in.ExternalUserID, code)
			return conversation.Result{
				MessageID:        in.ExternalMessageID,
				ConversationID:   "",
				Content:          replyText,
				HandoffRequested: false,
			}, nil
		}
	} else if strings.HasPrefix(trimmedText, "/done") || strings.HasPrefix(trimmedText, "/DONE") {
		if s.handoffs != nil {
			parts := strings.Fields(trimmedText)
			code := ""
			if len(parts) > 1 {
				code = parts[1]
			}
			replyText, _ := s.handoffs.ResolveByCommand(ctx, in.ChannelID, in.ExternalUserID, code)
			return conversation.Result{
				MessageID:        in.ExternalMessageID,
				ConversationID:   "",
				Content:          replyText,
				HandoffRequested: false,
			}, nil
		}
	}

	return s.conversations.Handle(ctx, conversation.Incoming{ChannelID: in.ChannelID, ChannelType: "whatsapp", ExternalMessageID: in.ExternalMessageID, ExternalUserID: in.ExternalUserID, Text: in.Content.Text, Provider: "whatsapp", Environment: "production"})
}

func (s *Service) ChannelStatus(ctx context.Context, in ChannelStatusInput) error {
	if in.ChannelID == "" || !validChannelStatus(in.Status) {
		return ErrInvalidInput
	}
	return s.repo.UpdateChannelStatus(ctx, in.ChannelID, in.Status, in.PhoneNumber, in.SessionID)
}

func (s *Service) ChannelEvent(ctx context.Context, in ChannelEventInput) error {
	status := in.Event
	switch in.Event {
	case "pair_success":
		status = "connected"
	case "logged_out":
		status = "disconnected"
	case "qr_refresh":
		status = "waiting_pairing"
	}
	if !validChannelStatus(status) {
		return nil
	}
	return s.ChannelStatus(ctx, ChannelStatusInput{ChannelID: in.ChannelID, Status: status, PhoneNumber: in.PhoneNumber, SessionID: in.SessionID})
}

func (s *Service) Respond(ctx context.Context, agentID string, in RespondInput) (conversation.Result, error) {
	if agentID == "" || in.ConversationID == "" || strings.TrimSpace(in.Message) == "" {
		return conversation.Result{}, ErrInvalidInput
	}
	resolved, err := s.repo.ResolveRuntimeContext(ctx, agentID, in.ConversationID)
	if err != nil {
		return conversation.Result{}, err
	}
	return s.conversations.Handle(ctx, conversation.Incoming{ChannelID: resolved.ChannelID, ChannelType: resolved.ChannelType, ExternalMessageID: in.ConversationID + ":" + agentID + ":" + in.Message, ExternalUserID: resolved.ExternalUserID, Text: in.Message, Provider: "internal", Environment: "production"})
}

func (s *Service) Health(ctx context.Context, agentID string) (AgentHealth, error) {
	if agentID == "" {
		return AgentHealth{}, ErrInvalidInput
	}
	return s.repo.GetAgentHealth(ctx, agentID)
}

func validChannelStatus(v string) bool {
	switch v {
	case "waiting_pairing", "connecting", "connected", "reconnecting", "disconnected", "error", "suspended":
		return true
	}
	return false
}
