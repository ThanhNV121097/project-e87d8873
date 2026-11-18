// Package migrations embeds the schema's SQL files, so the one server binary
// carries them. The directive sits beside the files it embeds: a pattern is
// resolved from the file that declares it.
package migrations

import "embed"

// FS holds every *.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
