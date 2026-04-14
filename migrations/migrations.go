// Package migrations embeds SQL migration files for use with goose.
package migrations

import "embed"

// FS contains all goose SQL migration files.
//
//go:embed *.sql
var FS embed.FS
