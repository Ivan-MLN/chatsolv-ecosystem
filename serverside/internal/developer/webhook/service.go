package webhook

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidInput        = errors.New("invalid webhook input")
	ErrForbidden           = errors.New("webhook forbidden")
	ErrEntitlementRequired = errors.New("webhook entitlement required")
)

type Endpoint struct {
	ID               string    `json:"id"`
	WorkspaceID      string    `json:"workspace_id"`
	URL              string    `json:"url"`
	Events           []string  `json:"events"`
	Status           string    `json:"status"`
	SecretCiphertext []byte    `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type CreateInput struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}
type UpdateInput struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Status string   `json:"status"`
}
type Created struct {
	Endpoint Endpoint `json:"endpoint"`
	Secret   string   `json:"secret"`
}
type Repository interface {
	Authorize(context.Context, string, string) (string, bool, error)
	Create(context.Context, Endpoint) error
	List(context.Context, string, string) ([]Endpoint, error)
	Update(context.Context, string, string, UpdateInput) (Endpoint, error)
	Delete(context.Context, string, string) error
}
type Service struct {
	repo Repository
	key  []byte
}

func NewService(r Repository, key []byte) *Service { return &Service{r, append([]byte(nil), key...)} }
func (s *Service) Create(ctx context.Context, userID, workspaceID string, in CreateInput) (Created, error) {
	role, enabled, e := s.repo.Authorize(ctx, userID, workspaceID)
	if e != nil {
		return Created{}, e
	}
	if role != "owner" && role != "admin" {
		return Created{}, ErrForbidden
	}
	if !enabled {
		return Created{}, ErrEntitlementRequired
	}
	if !valid(in.URL, in.Events) {
		return Created{}, ErrInvalidInput
	}
	raw := make([]byte, 32)
	if _, e = rand.Read(raw); e != nil {
		return Created{}, e
	}
	secret := "whsec_" + base64.RawURLEncoding.EncodeToString(raw)
	encrypted, e := encrypt(s.key, []byte(secret))
	if e != nil {
		return Created{}, e
	}
	now := time.Now().UTC()
	endpoint := Endpoint{ID: uuid.NewString(), WorkspaceID: workspaceID, URL: in.URL, Events: append([]string(nil), in.Events...), Status: "active", SecretCiphertext: encrypted, CreatedAt: now, UpdatedAt: now}
	if e = s.repo.Create(ctx, endpoint); e != nil {
		return Created{}, e
	}
	return Created{Endpoint: endpoint, Secret: secret}, nil
}
func (s *Service) List(ctx context.Context, userID, workspaceID string) ([]Endpoint, error) {
	role, _, e := s.repo.Authorize(ctx, userID, workspaceID)
	if e != nil {
		return nil, e
	}
	if role != "owner" && role != "admin" {
		return nil, ErrForbidden
	}
	return s.repo.List(ctx, workspaceID, userID)
}
func (s *Service) Update(ctx context.Context, userID, id string, in UpdateInput) (Endpoint, error) {
	if !valid(in.URL, in.Events) || (in.Status != "active" && in.Status != "disabled") {
		return Endpoint{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, id, userID, in)
}
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id, userID)
}
func valid(rawURL string, events []string) bool {
	u, e := url.ParseRequestURI(rawURL)
	if e != nil || u.Scheme != "https" || u.Host == "" || len(events) == 0 || len(events) > 6 {
		return false
	}
	allowed := map[string]bool{"conversation.created": true, "conversation.updated": true, "message.received": true, "message.created": true, "handoff.requested": true, "agent.error": true}
	seen := map[string]bool{}
	for _, v := range events {
		if !allowed[v] || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}
func encrypt(key, value []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	gcm, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, e = io.ReadFull(rand.Reader, nonce); e != nil {
		return nil, e
	}
	return gcm.Seal(nonce, nonce, value, nil), nil
}
