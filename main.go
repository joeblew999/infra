package main

import (
	"github.com/joeblew999/infra/pkg/cmd"
	"github.com/joeblew999/infra/pkg/log"
)

func main() {
	// Initialize the global logger
	log.InitLogger("", "info", false) // No log file, info level, text format

	cmd.Execute()
}
