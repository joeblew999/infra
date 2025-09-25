package main

import (
	"fmt"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
)

func main() {
	cfg := runtimecfg.Load()
	fmt.Printf("core runtime placeholder\n")
	fmt.Printf("environment: %s\n", cfg.Environment)
	fmt.Printf("app root: %s\n", cfg.Paths.AppRoot)
}
