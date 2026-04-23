package logger

import (
    "os"
    "log/slog"
)

func Init() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    slog.SetDefault(logger)
}