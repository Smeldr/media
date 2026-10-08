// Package media provides media upload, storage, and serving for Smeldr
// applications. It is an optional submodule with zero additional dependencies.
//
// # Quick start
//
//	import "smeldr.dev/media"
//
//	store := media.NewLocalMediaStore(app)
//	srv   := media.New(app, store)
//	srv.Register(app)
//
// Uploaded files are validated by magic-byte MIME detection, stored in the
// directory configured by [smeldr.Config.MediaPath] (default ./media), and
// served at GET /media/{filename}. All write operations require at least the
// Author role.
//
// # Database
//
// Media records live in the application's own database ([smeldr.Config.DB]),
// in the smeldr_media table that [CreateMediaTable] creates. It works on SQLite
// and on Postgres (smeldr.dev/core/pgx): every query uses numbered
// placeholders, and a legacy forge_media table is renamed on either through
// [smeldr.RenameLegacyTables]. The integration tests (build tag integration,
// DATABASE_URL) run the store against Postgres; they are the module's only use
// of a Postgres driver.
package media
