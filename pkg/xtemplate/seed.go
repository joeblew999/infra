package xtemplate

import "embed"

// assetsFS contains bundled template sets (local seeds + upstream sync).
//
//go:embed templates/seed/** templates/upstream/**
var assetsFS embed.FS
