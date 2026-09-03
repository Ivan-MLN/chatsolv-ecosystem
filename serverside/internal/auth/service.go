package auth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Service struct {
	users                UserRepository
	tokens               TokenStore
	email                EmailSender
	hasher               *Argon2Hasher
	jwt                  *JWTManager
	refreshTTL, resetTTL time.Duration
}

func NewService(u UserRepository, t TokenStore, e EmailSender, h *Argon2Hasher, j *JWTManager, rt, pt time.Duration) *Service {
	return &Service{u, t, e, h, j, rt, pt}
}
func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func (s *Service) Register(ctx context.Context, in RegisterInput) (User, error) {
	hash, e := s.hasher.Hash(in.Password)
	if e != nil {
		return User{}, e
	}
	u := User{ID: uuid.NewString(), Name: strings.TrimSpace(in.Name), Email: normalizeEmail(in.Email), PasswordHash: hash}
	if e = s.users.Create(ctx, u); e != nil {
		return User{}, e
	}
	return u, nil
}
func (s *Service) Login(ctx context.Context, in LoginInput) (AuthTokens, error) {
	u, e := s.users.GetByEmail(ctx, normalizeEmail(in.Email))
	if errors.Is(e, ErrUserNotFound) {
		return AuthTokens{}, ErrInvalidCredentials
	}
	if e != nil {
		return AuthTokens{}, e
	}
	ok, e := s.hasher.Verify(in.Password, u.PasswordHash)
	if e != nil || !ok {
		return AuthTokens{}, ErrInvalidCredentials
	}
	access, exp, e := s.jwt.Generate(u.ID)
	if e != nil {
		return AuthTokens{}, e
	}
	refresh, e := randomToken()
	if e != nil {
		return AuthTokens{}, e
	}
	if e = s.tokens.SaveRefresh(ctx, u.ID, tokenHash(refresh), s.refreshTTL); e != nil {
		return AuthTokens{}, e
	}
	return AuthTokens{access, refresh, "Bearer", int64(time.Until(exp).Seconds())}, nil
}
func (s *Service) Refresh(ctx context.Context, refreshToken string) (AuthTokens, error) {
	userID, err := s.tokens.ConsumeRefresh(ctx, tokenHash(refreshToken))
	if err != nil {
		return AuthTokens{}, err
	}
	access, expiresAt, err := s.jwt.Generate(userID)
	if err != nil {
		return AuthTokens{}, err
	}
	rotated, err := randomToken()
	if err != nil {
		return AuthTokens{}, err
	}
	if err = s.tokens.SaveRefresh(ctx, userID, tokenHash(rotated), s.refreshTTL); err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: access, RefreshToken: rotated, TokenType: "Bearer", ExpiresIn: int64(time.Until(expiresAt).Seconds())}, nil
}
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	// Keep the externally observable response time above a common floor so an
	// unknown address does not return substantially faster than a known one.
	started := time.Now()
	defer func() {
		remaining := 100*time.Millisecond - time.Since(started)
		if remaining <= 0 {
			return
		}
		timer := time.NewTimer(remaining)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
		}
	}()
	u, e := s.users.GetByEmail(ctx, normalizeEmail(email))
	if errors.Is(e, ErrUserNotFound) {
		return nil
	}
	if e != nil {
		return e
	}
	token, e := randomToken()
	if e != nil {
		return e
	}
	if e = s.tokens.SaveReset(ctx, tokenHash(token), u.ID, s.resetTTL); e != nil {
		return e
	}
	return s.email.SendPasswordReset(ctx, u.Email, token)
}
func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	uid, e := s.tokens.ConsumeReset(ctx, tokenHash(token))
	if errors.Is(e, ErrInvalidResetToken) {
		return ErrInvalidResetToken
	}
	if e != nil {
		return e
	}
	hash, e := s.hasher.Hash(password)
	if e != nil {
		return e
	}
	if e = s.users.UpdatePassword(ctx, uid, hash); e != nil {
		return e
	}
	return s.tokens.RevokeUserSessions(ctx, uid)
}
