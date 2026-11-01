package main

import (
	"log/slog"
	"os"

	"github.com/phanivarun05/kubeclusterflow/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}
