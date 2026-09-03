package publicapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"authbackend/internal/conversation"
	"authbackend/internal/developer/apikey"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid public API input")
	ErrInvalidToken = errors.New("invalid client token")
)

const sessionTTL = time.Hour

type CreateSessionInput struct {
	ExternalUserID string         `json:"external_user_id"`
	Metadata       map[string]any `json:"metadata"`
}
type Session struct {
	ID, WorkspaceID, AgentID, ChannelID, ExternalUserID, TokenHash string
	Metadata                                                       map[string]any
	ExpiresAt                                                      time.Time
}
type CreatedSession struct {
	SessionID   string `json:"session_id"`
	ClientToken string `json:"client_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
type MessageResult struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}
type KeyAuthenticator interface {
	Authenticate(context.Context, string, string) (apikey.Record, error)
}
type SessionRepository interface {
	Create(context.Context, Session) error
	Resolve(context.Context, string, string) (Session, error)
}
type ConversationService interface {
	Handle(context.Context, conversation.Incoming) (conversation.Result, error)
}
type Service struct {
	keys          KeyAuthenticator
	repo          SessionRepository
	conversations ConversationService
}

func NewService(keys KeyAuthenticator, repo SessionRepository, conversations ConversationService) *Service {
	return &Service{keys: keys, repo: repo, conversations: conversations}
}
func (s *Service) CreateSession(ctx context.Context, apiKey string, input CreateSessionInput) (CreatedSession, error) {
	input.ExternalUserID = strings.TrimSpace(input.ExternalUserID)
	if input.ExternalUserID == "" || len(input.ExternalUserID) > 255 {
		return CreatedSession{}, ErrInvalidInput
	}
	record, err := s.keys.Authenticate(ctx, apiKey, "agent:invoke")
	if err != nil {
		return CreatedSession{}, err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return CreatedSession{}, err
	}
	token := "csc_" + base64.RawURLEncoding.EncodeToString(raw)
	session := Session{ID: uuid.NewString(), WorkspaceID: record.WorkspaceID, ExternalUserID: input.ExternalUserID, TokenHash: tokenHash(token), Metadata: input.Metadata, ExpiresAt: time.Now().UTC().Add(sessionTTL)}
	if err = s.repo.Create(ctx, session); err != nil {
		return CreatedSession{}, err
	}
	return CreatedSession{SessionID: session.ID, ClientToken: token, ExpiresIn: int64(sessionTTL.Seconds())}, nil
}
func (s *Service) SendMessage(ctx context.Context, sessionID, clientToken, message string) (MessageResult, error) {
	message = strings.TrimSpace(message)
	if message == "" || len(message) > 8000 || sessionID == "" || clientToken == "" {
		return MessageResult{}, ErrInvalidInput
	}
	session, err := s.repo.Resolve(ctx, sessionID, tokenHash(clientToken))
	if err != nil {
		return MessageResult{}, ErrInvalidToken
	}
	result, err := s.conversations.Handle(ctx, conversation.Incoming{ChannelID: session.ChannelID, ChannelType: "web", ExternalMessageID: uuid.NewString(), ExternalUserID: session.ExternalUserID, Text: message, Provider: "public_api", Environment: "production"})
	if err != nil {
		return MessageResult{}, err
	}
	return MessageResult{MessageID: result.MessageID, ConversationID: result.ConversationID, Content: result.Content}, nil
}
func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
