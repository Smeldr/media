package media

import (
	"context"

	"smeldr.dev/core"
)

// migrateLegacyTableNames renames the forge_media table to smeldr_media if it
// still exists. It is called from [CreateMediaTable] once at startup before the
// CREATE TABLE statement runs. It works on SQLite and Postgres through
// [smeldr.RenameLegacyTables]: a pair whose source is absent is skipped, and
// one whose source and destination both exist is skipped with a warning, so
// re-running on a migrated database is safe.
func migrateLegacyTableNames(ctx context.Context, db smeldr.DB) error {
	return smeldr.RenameLegacyTables(ctx, db, [][2]string{
		{"forge_media", "smeldr_media"},
	})
}
