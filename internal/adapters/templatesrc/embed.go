package templatesrc

import "embed"

//go:embed all:cpp all:go all:rust
var FS embed.FS
