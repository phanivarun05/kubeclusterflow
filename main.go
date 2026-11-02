package main

import (
	"log/slog"
	"os"

	"github.com/phanivarun05/kubeclusterflow/cmd"
	"github.com/phanivarun05/kubeclusterflow/internal/logger"
)

func main() {
	logger.InitLogger(slog.LevelError)
	if err := cmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}
