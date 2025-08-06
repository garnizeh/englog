package handlers_test

import "github.com/garnizeh/englog/internal/logging"

func Logger() *logging.Logger {
	logConfig := logging.Config{
		Level:  logging.DebugLevel,
		Format: "json",
	}

	return logging.NewLogger(logConfig)
}
