// Package web embeds static files for the Gophermart admin panel.
package web

import "embed"

// FS contains the embedded admin HTML file.
//
//go:embed admin.html
var FS embed.FS
