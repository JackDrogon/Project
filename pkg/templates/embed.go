package templates

import "embed"

//go:embed all:cpp all:go
var FS embed.FS
