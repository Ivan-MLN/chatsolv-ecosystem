package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidKey = errors.New("invalid API key")
	ErrForbidden  = errors.New("API key scope forbidden")
)

type Record struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Prefix      string     `json:"prefix"`
	Hash        string     `json:"-"`
	LastFour    string     `json:"last_four"`
	Name        string     `json:"name"`
	Scopes      []string   `json:"scopes"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}
type Created struct {
	Record Record `json:"record"`
	Secret string `json:"secret"`
}
type Repository interface {
	Create(context.Context, Record) error
	FindByPrefix(context.Context, string) (Record, error)
	AuthorizeWorkspace(context.Context, string, string) (string, error)
	List(context.Context, string, string) ([]Record, error)
	Revoke(context.Context, string, string) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID, workspaceID string) ([]Record, error) {
	role, err := s.repo.AuthorizeWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "admin" {
		return nil, ErrForbidden
	}
	return s.repo.List(ctx, workspaceID, userID)
}
func (s *Service) CreateForUser(ctx context.Context, userID, workspaceID, name string, scopes []string) (Created, error) {
	role, err := s.repo.AuthorizeWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return Created{}, err
	}
	if role != "owner" && role != "admin" {
		return Created{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 || !validScopes(scopes) {
		return Created{}, ErrInvalidKey
	}
	return s.Create(ctx, workspaceID, name, scopes)
}
func (s *Service) Revoke(ctx context.Context, userID, keyID string) error {
	if strings.TrimSpace(keyID) == "" {
		return ErrInvalidKey
	}
	return s.repo.Revoke(ctx, keyID, userID)
}
func (s *Service) Create(ctx context.Context, workspaceID, name string, scopes []string) (Created, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Created{}, err
	}
	secret := "cs_live_" + base64.RawURLEncoding.EncodeToString(raw)
	prefix := secret[:20]
	record := Record{ID: uuid.NewString(), WorkspaceID: workspaceID, Prefix: prefix, Hash: hash(secret), LastFour: secret[len(secret)-4:], Name: name, Scopes: append([]string(nil), scopes...), CreatedAt: time.Now().UTC()}
	if err := s.repo.Create(ctx, record); err != nil {
		return Created{}, err
	}
	return Created{Record: record, Secret: secret}, nil
}
func (s *Service) Authenticate(ctx context.Context, secret, scope string) (Record, error) {
	if !strings.HasPrefix(secret, "cs_live_") || len(secret) < 24 {
		return Record{}, ErrInvalidKey
	}
	record, err := s.repo.FindByPrefix(ctx, secret[:20])
	if err != nil || record.RevokedAt != nil {
		return Record{}, ErrInvalidKey
	}
	expected, _ := hex.DecodeString(record.Hash)
	actualSum := sha256.Sum256([]byte(secret))
	if len(expected) != len(actualSum) || subtle.ConstantTimeCompare(expected, actualSum[:]) != 1 {
		return Record{}, ErrInvalidKey
	}
	for _, allowed := range record.Scopes {
		if allowed == scope {
			return record, nil
		}
	}
	return Record{}, ErrForbidden
}
func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func validScopes(scopes []string) bool {
	if len(scopes) == 0 || len(scopes) > 6 {
		return false
	}
	allowed := map[string]bool{"agent:invoke": true, "conversation:read": true, "conversation:write": true, "knowledge:read": true, "knowledge:write": true, "webhook:manage": true}
	seen := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		if !allowed[scope] || seen[scope] {
			return false
		}
		seen[scope] = true
	}
	return true
}
