package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// openStore opens (or creates) the per-channel whatsmeow SQLite store.
// The store file lives at <dbRoot>/<channelID>.db.
// whatsmeow runs its own schema migrations via container.Upgrade.
func openStore(ctx context.Context, dbRoot, channelID string) (*sqlstore.Container, error) {
	path := filepath.Join(dbRoot, channelID+".db")

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create sessions dir: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// NewWithDB wraps the existing *sql.DB; dialect "sqlite3" is the whatsmeow
	// canonical name regardless of the actual driver in use.
	container := sqlstore.NewWithDB(db, "sqlite3", waLog.Noop)
	if err = container.Upgrade(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("upgrade store schema: %w", err)
	}

	return container, nil
}
