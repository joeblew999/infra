package main

import (
	"fmt"
	"os"

	// https://github.com/F1bonacc1/process-compose/tree/main/src/cmd
	pccmd "github.com/f1bonacc1/process-compose/src/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "process-compose: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Delegate to upstream Process Compose CLI. Keeping this indirection makes
	// it easy to layer local setup/teardown hooks before or after the upstream
	// command runs.
	pccmd.Execute()
	return nil
}
