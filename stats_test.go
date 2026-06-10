package media

import (
	"context"
	"testing"
	"time"
)

func TestProvideStats(t *testing.T) {
	srv, _ := newTestServer(t)

	seed := []MediaRecord{
		{ID: "1", Filename: "a.jpg", OriginalFilename: "a.jpg", MediaType: MediaTypeImage, MIMEType: "image/jpeg", SizeBytes: 100, UploadedAt: time.Now()},
		{ID: "2", Filename: "b.jpg", OriginalFilename: "b.jpg", MediaType: MediaTypeImage, MIMEType: "image/jpeg", SizeBytes: 200, UploadedAt: time.Now()},
		{ID: "3", Filename: "c.png", OriginalFilename: "c.png", MediaType: MediaTypeImage, MIMEType: "image/png", SizeBytes: 300, UploadedAt: time.Now()},
	}
	for _, r := range seed {
		if err := insertMedia(srv.db, r); err != nil {
			t.Fatalf("insertMedia: %v", err)
		}
	}

	if key := srv.StatsKey(); key != "media" {
		t.Errorf("StatsKey() = %q, want %q", key, "media")
	}

	stats, err := srv.ProvideStats(context.Background())
	if err != nil {
		t.Fatalf("ProvideStats: %v", err)
	}

	if got := stats["file_count"].(int64); got != 3 {
		t.Errorf("file_count = %d, want 3", got)
	}
	if got := stats["total_bytes"].(int64); got != 600 {
		t.Errorf("total_bytes = %d, want 600", got)
	}

	byType, ok := stats["by_type"].(map[string]int64)
	if !ok {
		t.Fatalf("by_type has unexpected type %T", stats["by_type"])
	}
	if got := byType["image/jpeg"]; got != 2 {
		t.Errorf("by_type[image/jpeg] = %d, want 2", got)
	}
	if got := byType["image/png"]; got != 1 {
		t.Errorf("by_type[image/png] = %d, want 1", got)
	}
}

func TestProvideStats_empty(t *testing.T) {
	srv, _ := newTestServer(t)

	stats, err := srv.ProvideStats(context.Background())
	if err != nil {
		t.Fatalf("ProvideStats on empty table: %v", err)
	}
	if got := stats["file_count"].(int64); got != 0 {
		t.Errorf("file_count = %d, want 0", got)
	}
	if got := stats["total_bytes"].(int64); got != 0 {
		t.Errorf("total_bytes = %d, want 0", got)
	}
}
