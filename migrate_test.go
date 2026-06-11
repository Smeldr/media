package media

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"smeldr.dev/core"
)

func tableExists(t *testing.T, db smeldr.DB, name string) bool {
	t.Helper()
	var n int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&n)
	if err != nil {
		t.Fatalf("tableExists(%q): %v", name, err)
	}
	return n > 0
}

func TestMigrateLegacyMediaTable_freshDB(t *testing.T) {
	db := openTestDB(t)
	if err := CreateMediaTable(db); err != nil {
		t.Fatalf("CreateMediaTable: %v", err)
	}
	if !tableExists(t, db, "smeldr_media") {
		t.Error("expected smeldr_media to exist")
	}
	if tableExists(t, db, "forge_media") {
		t.Error("expected forge_media to not exist")
	}
}

func TestMigrateLegacyMediaTable_existingForge(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("sqlite unavailable: %v", err)
	}
	defer rawDB.Close()

	var db smeldr.DB = rawDB
	_, err = rawDB.Exec(`
		CREATE TABLE forge_media (
			id TEXT PRIMARY KEY, filename TEXT, original_filename TEXT,
			media_type TEXT, mime_type TEXT, description TEXT,
			size_bytes INTEGER, uploaded_at DATETIME
		)`)
	if err != nil {
		t.Fatalf("create legacy table: %v", err)
	}

	if err := CreateMediaTable(db); err != nil {
		t.Fatalf("CreateMediaTable: %v", err)
	}
	if !tableExists(t, db, "smeldr_media") {
		t.Error("expected smeldr_media to exist after migration")
	}
	if tableExists(t, db, "forge_media") {
		t.Error("expected forge_media to be gone after migration")
	}
}

func TestMigrateLegacyMediaTable_idempotent(t *testing.T) {
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("sqlite unavailable: %v", err)
	}
	defer rawDB.Close()

	var db smeldr.DB = rawDB
	_, err = rawDB.Exec(`
		CREATE TABLE forge_media (
			id TEXT PRIMARY KEY, filename TEXT, original_filename TEXT,
			media_type TEXT, mime_type TEXT, description TEXT,
			size_bytes INTEGER, uploaded_at DATETIME
		)`)
	if err != nil {
		t.Fatalf("create legacy table: %v", err)
	}

	ctx := context.Background()
	if err := migrateLegacyTableNames(ctx, db); err != nil {
		t.Fatalf("first call: %v", err)
	}
	// Second call: forge_media is gone; smeldr_media exists — must not error.
	if err := migrateLegacyTableNames(ctx, db); err != nil {
		t.Fatalf("second call: %v", err)
	}
}
