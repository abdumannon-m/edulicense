package templates

import "embed"

// FS contains the server-rendered public and admin templates.
//
//go:embed *.html
var FS embed.FS
