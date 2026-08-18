package logutil

import (
	// Built-in/core modules.
	"io"
	"log/slog"
	"os"
)

// Returns a new *slog.Logger that logs to stderr. It also writes to a separate
// error log if the log level is ERROR or higher. The log file is created on
// the first write. The log level for stderr is set to INFO by default, but it
// can be changed at runtime using the returned *slog.LevelVar. Call Close on
// the returned io.Closer when the file when the logger is no longer needed to
// ensure that the log file is properly closed.
func NewMultilogger() (*slog.Logger, *slog.LevelVar, io.Closer) {
	errorHandler, closer := NewErrorHandler(nil)

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	stdHandler := slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{Level: logLevel})
	return slog.New(slog.NewMultiHandler(errorHandler, stdHandler)),
		logLevel, closer
}
