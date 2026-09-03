package template

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestGetTemplatesFromDB(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres@127.0.0.1:5433/auth_db?sslmode=disable")
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresRepository(pool)
	templates, err := repo.List(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, templates)
	t.Logf("Found %d templates in DB", len(templates))
	for _, tmpl := range templates {
		t.Logf("- %s (%s): %s", tmpl.ID, tmpl.Industry, tmpl.Title)
	}
}
