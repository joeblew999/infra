package main

import (
	"os"

	infraLog "github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service/runtime"
)

func main() {
	infraLog.InitLogger("", "info", false)

	if err := runtime.Start(runtime.Options{Mode: "project-test"}); err != nil {
		infraLog.Error("runtime exited with error", "error", err)
		os.Exit(1)
	}
}
