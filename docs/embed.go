package docs

import (
	"embed"
)

//go:embed *
//go:embed roadmap/*
//go:embed roadmap/ROADMAP.md
var EmbeddedFS embed.FS
