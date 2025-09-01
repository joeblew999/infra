package docs

import (
	"embed"
)

//go:embed *
//go:embed business/*
//go:embed technical/*
var EmbeddedFS embed.FS
