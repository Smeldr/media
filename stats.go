package media

import (
	"context"
	"fmt"
)

// StatsKey returns the key used in smeldr.SiteStats.External.
// Implements [smeldr.StatsExtProvider].
func (s *Server) StatsKey() string { return "media" }

// ProvideStats returns aggregate statistics for all media records:
// file_count (total files), total_bytes (sum of sizes), and by_type
// (MIME type → count). Implements [smeldr.StatsExtProvider].
func (s *Server) ProvideStats(ctx context.Context) (map[string]any, error) {
	var count, totalBytes int64
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(size_bytes), 0) FROM forge_media`)
	if err := row.Scan(&count, &totalBytes); err != nil {
		return nil, fmt.Errorf("media: stats: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT mime_type, COUNT(*) FROM forge_media GROUP BY mime_type`)
	if err != nil {
		return nil, fmt.Errorf("media: stats by type: %w", err)
	}
	defer rows.Close()
	byType := map[string]int64{}
	for rows.Next() {
		var mt string
		var n int64
		if err := rows.Scan(&mt, &n); err != nil {
			return nil, fmt.Errorf("media: stats by type scan: %w", err)
		}
		byType[mt] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("media: stats by type rows: %w", err)
	}

	return map[string]any{
		"file_count":  count,
		"total_bytes": totalBytes,
		"by_type":     byType,
	}, nil
}
