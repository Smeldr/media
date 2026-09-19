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
package media
