package templates

import _ "embed"

// Page templates
//
//go:embed pages/index.html
var IndexHTML []byte

//go:embed pages/logs.html
var LogsHTML []byte

//go:embed pages/bento-playground.html
var BentoPlaygroundHTML []byte

//go:embed pages/404.html
var NotFoundHTML []byte
