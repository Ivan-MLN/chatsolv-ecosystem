package auth

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(context.Context, User) error
	GetByEmail(context.Context, string) (User, error)
	UpdatePassword(context.Context, string, string) error
}
type TokenStore interface {
	SaveRefresh(context.Context, string, string, time.Duration) error
	ConsumeRefresh(context.Context, string) (string, error)
	SaveReset(context.Context, string, string, time.Duration) error
	ConsumeReset(context.Context, string) (string, error)
	RevokeUserSessions(context.Context, string) error
}
type EmailSender interface {
	SendPasswordReset(context.Context, string, string) error
}
