package media_test

import (
	"database/sql"

	"smeldr.dev/core"
	"smeldr.dev/media"
)

// ExampleRegister demonstrates the minimal media wiring pattern.
//
// Call Register to create the [Server] and mount all four HTTP routes
// (upload, serve, list, delete) on the application in a single step.
//
// To expose media via MCP, pass the returned *Server to
// mcp.WithModule when constructing the MCP server:
//
//	mcpSrv := mcp.New(app, mcp.WithModule(mediaSrv))
func ExampleRegister() {
	var db *sql.DB // initialise from your application database setup

	app := smeldr.New(smeldr.MustConfig(smeldr.Config{
		BaseURL: "https://example.com",
		Secret:  []byte("change-this-secret-in-production!"),
		DB:      db,
	}))

	store := media.NewLocalMediaStore(app)
	media.Register(app, store)
}
