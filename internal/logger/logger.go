package logger

import (
	"log/slog"
	"os"
)

// InitLogger must be called manually, allowing you to pass configurations
func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	customLogger := slog.New(jsonHandler)

	slog.SetDefault(customLogger)
}
