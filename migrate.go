package media

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"smeldr.dev/core"
)

// migrateLegacyTableNames renames the forge_media table to smeldr_media if it
// still exists. It is called from [CreateMediaTable] once at startup before the
// CREATE TABLE statement runs.
//
// Only operates on SQLite databases (identified by sqlite_master). For other
// databases it returns nil immediately.
//
// Idempotency: if both forge_media and smeldr_media already exist the rename is
// skipped with a warning. Re-running on an already-migrated database is safe.
func migrateLegacyTableNames(ctx context.Context, db smeldr.DB) error {
	pairs := [][2]string{
		{"forge_media", "smeldr_media"},
	}

	var dummy int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master`).Scan(&dummy); err != nil {
		return nil // not SQLite — skip silently
	}

	var toRename [][2]string
	for _, pair := range pairs {
		var srcN int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, pair[0],
		).Scan(&srcN); err != nil || srcN == 0 {
			continue // source doesn't exist — nothing to rename
		}
		var dstN int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, pair[1],
		).Scan(&dstN); err == nil && dstN > 0 {
			slog.Warn("media: legacy table migration skipped — destination already exists",
				"src", pair[0], "dst", pair[1])
			continue
		}
		toRename = append(toRename, pair)
	}
	if len(toRename) == 0 {
		return nil
	}

	type transactor interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}

	execDB := db
	var commit func() error = func() error { return nil }
	var rollback func() error = func() error { return nil }

	if tr, ok := db.(transactor); ok {
		tx, err := tr.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("media: migrate legacy tables: begin: %w", err)
		}
		execDB = tx
		commit = tx.Commit
		rollback = tx.Rollback
	}

	for _, pair := range toRename {
		slog.Info("media: renaming legacy table", "from", pair[0], "to", pair[1])
		if _, err := execDB.ExecContext(ctx, `ALTER TABLE `+pair[0]+` RENAME TO `+pair[1]); err != nil {
			_ = rollback()
			return fmt.Errorf("media: migrate legacy tables: %s → %s: %w", pair[0], pair[1], err)
		}
	}
	return commit()
}
