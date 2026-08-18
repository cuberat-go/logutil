package main

import "github.com/cuberat-go/logutil"

func main() {
	logger, logLevel, closer := logutil.NewMultilogger()
	defer closer.Close()

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

	logger3.Error("With log level", "logLevel", logLevel.Level())
}
