package main

import (
	"embed"
)

//go:embed all:docs
var EmbeddedDocs embed.FS
