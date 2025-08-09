package docs

import (
	"embed"
)

//go:embed *
//go:embed roadmap/*
var EmbeddedFS embed.FS
