// Copyright (C) 2017 Jan Delgado

package main

import (
	"log/slog"
	"os"
)

func initLogging(verbose bool) *slog.Logger {
	opts := slog.HandlerOptions{}
	if verbose {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelWarn
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &opts))
}
