package templates

import "embed"

// SeedFS exposes the starter templates used to bootstrap new projects.
//go:embed seed/* seed/**/*
var SeedFS embed.FS
