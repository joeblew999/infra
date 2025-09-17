package docs

import (
	"embed"
)

//go:embed * business technical examples
var EmbeddedFS embed.FS
