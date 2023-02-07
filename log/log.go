package log

import (
	"golang.org/x/exp/slog"
	"os"
)

func init() {
	options := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	slog.SetDefault(slog.New(options.NewTextHandler(os.Stderr)))
}
