package auth

import (
	"context"
	"log/slog"
)

type DevelopmentEmailSender struct{ log *slog.Logger }

func NewDevelopmentEmailSender(l *slog.Logger) *DevelopmentEmailSender {
	return &DevelopmentEmailSender{l}
}
func (e *DevelopmentEmailSender) SendPasswordReset(ctx context.Context, email, token string) error {
	e.log.InfoContext(ctx, "development password reset delivery", "email", email, "reset_token", token)
	return nil
}
