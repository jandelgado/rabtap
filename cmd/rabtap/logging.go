// Copyright (C) 2017 Jan Delgado

package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func initLogging(f *os.File, verbose bool, colored bool) *slog.Logger {
	opts := tint.Options{
		NoColor:    !colored,
		TimeFormat: time.TimeOnly,
	}
	if verbose {
		opts.Level = slog.LevelDebug
	} else {
		opts.Level = slog.LevelWarn
	}
	return slog.New(tint.NewHandler(os.Stderr, &opts))
}
