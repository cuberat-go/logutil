// This program creates a multi logger that writes log messages to both a log
// file in the ~/log/multilogger/ directory (when the log level is ERROR or
// higher) and to stderr (when the log level is INFO or higher). The error log
// file is created on the first write. The log file name is generated based on
// the program name, current timestamp, and process ID.
package main

import (
	// Built-in/core modules.
	"log/slog"
	"os"

	// Third-party modules.
	"github.com/cuberat-go/logutil"
)

func main() {
	errorHandler, closer := logutil.NewErrorHandler(nil)
	defer closer.Close()

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	stdHandler := slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{Level: logLevel})
	logger := slog.New(slog.NewMultiHandler(errorHandler, stdHandler))

	logger.Debug("This is a debug message", "key", "value")
	logger.Info("This is an info message", "key", "value")
	logger.Error("This is an example error message", "foo", "bar")

	logger2 := logger.With("top_attr1", "value1", "top_attr2", 123)
	logger2.Error("This is an example error message", "foo", "bar")

	logger3 := logger2.WithGroup("group1").With(
		"group1_attr1", "value1",
		"group1_attr2", 456,
	)
	logger3.Error("This is an example error message", "foo", "bar")
}
