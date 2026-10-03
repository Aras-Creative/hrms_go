// Package migrations embeds the SQL migration files so tooling can apply them
// without depending on the golang-migrate CLI or on the migrations/ directory
// being present next to the built binary.
//
// The files must live in this directory: //go:embed cannot walk upwards, so a
// pattern like ../../migrations is rejected by the compiler.
package migrations

import "embed"

// FS holds every .up.sql and .down.sql in this directory.
//
//go:embed *.up.sql *.down.sql
var FS embed.FS
