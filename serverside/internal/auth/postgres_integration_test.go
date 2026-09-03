package auth

import (
	"authbackend/generated/sqlc"
	"authbackend/internal/testutil"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgresRepositoryIntegration(t *testing.T) {
	p := testutil.Postgres(t)
	ctx := context.Background()
	r := NewPostgresRepository(sqlc.New(p))
	email := uuid.NewString() + "@example.com"
	u := User{ID: uuid.NewString(), Name: "Test", Email: email, PasswordHash: "hash"}
	require.NoError(t, r.Create(ctx, u))
	got, err := r.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)
	require.ErrorIs(t, r.Create(ctx, User{ID: uuid.NewString(), Name: "Other", Email: email, PasswordHash: "hash"}), ErrUserExists)
	require.NoError(t, r.UpdatePassword(ctx, u.ID, "newhash"))
	got, err = r.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, "newhash", got.PasswordHash)
}
