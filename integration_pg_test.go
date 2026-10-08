//go:build integration

package media

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smeldr.dev/core"
)

// pgSchemaDB gives the test a Postgres schema of its own (search_path),
// dropped on cleanup. Needs DATABASE_URL and the integration build tag.
func pgSchemaDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	ctx := context.Background()
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	schema := "t_" + strings.ReplaceAll(strings.ToLower(smeldr.NewID()), "-", "")
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatalf("create schema: %v", err)
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	db, err := sql.Open("pgx", dsn+sep+"search_path="+schema)
	if err != nil {
		t.Fatalf("open in schema: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	})
	return db
}

// The media table and every query on Postgres: create (twice), insert,
// list all and by type, get, stats with a size over 2 GiB, delete.
func TestPG_MediaStore(t *testing.T) {
	db := pgSchemaDB(t)
	for i := 0; i < 2; i++ {
		if err := CreateMediaTable(db); err != nil {
			t.Fatalf("CreateMediaTable call %d: %v", i+1, err)
		}
	}
	big := int64(3) << 30 // 3 GiB: would overflow a 32-bit INTEGER column
	recs := []MediaRecord{
		{ID: "m1", Filename: "a.jpg", OriginalFilename: "a.jpg", MediaType: "image", MIMEType: "image/jpeg", Description: "a", SizeBytes: 10, UploadedAt: time.Now().Add(-time.Minute)},
		{ID: "m2", Filename: "b.mp4", OriginalFilename: "b.mp4", MediaType: "video", MIMEType: "video/mp4", SizeBytes: big, UploadedAt: time.Now()},
	}
	for _, r := range recs {
		if err := insertMedia(db, r); err != nil {
			t.Fatalf("insertMedia %s: %v", r.ID, err)
		}
	}
	all, err := listMedia(db, "")
	if err != nil || len(all) != 2 || all[0].ID != "m2" {
		t.Fatalf("listMedia = %+v, %v; want m2 first", all, err)
	}
	images, err := listMedia(db, "image")
	if err != nil || len(images) != 1 || images[0].ID != "m1" {
		t.Fatalf("listMedia(image) = %+v, %v", images, err)
	}
	got, err := getMediaByID(db, "m2")
	if err != nil || got.SizeBytes != big {
		t.Fatalf("getMediaByID = %+v, %v; want the 3 GiB size back", got, err)
	}
	stats, err := (&Server{db: db}).ProvideStats(context.Background())
	if err != nil || stats["file_count"] != int64(2) || stats["total_bytes"] != big+10 {
		t.Fatalf("stats = %v, %v", stats, err)
	}
	if err := deleteMediaRecord(db, "m1"); err != nil {
		t.Fatal(err)
	}
	if _, err := getMediaByID(db, "m1"); err != smeldr.ErrNotFound {
		t.Errorf("after delete err = %v; want ErrNotFound", err)
	}
}

// A legacy forge_media table is renamed on Postgres, rows and all.
func TestPG_MediaLegacyRename(t *testing.T) {
	db := pgSchemaDB(t)
	if _, err := db.Exec(`CREATE TABLE forge_media (
		id TEXT PRIMARY KEY, filename TEXT NOT NULL UNIQUE, original_filename TEXT NOT NULL,
		media_type TEXT NOT NULL, mime_type TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
		size_bytes BIGINT NOT NULL DEFAULT 0, uploaded_at TIMESTAMP NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO forge_media VALUES ('old', 'o.png', 'o.png', 'image', 'image/png', 'o', 1, now())`); err != nil {
		t.Fatal(err)
	}
	if err := CreateMediaTable(db); err != nil {
		t.Fatalf("CreateMediaTable: %v", err)
	}
	if _, err := getMediaByID(db, "old"); err != nil {
		t.Errorf("renamed row not found: %v", err)
	}
}
