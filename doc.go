// logutil is an opinionated logging utility package that provides a structured
// logging handler for Go's slog package. It allows for logging to a file with
// support for log levels, attributes, and groups. The package is designed to
// be used in applications that require structured logging with the ability to
// dynamically change log levels at runtime.
//
// The handler writes log entries to a log file in JSON format. The log file is
// created on the first write. By default, the log level is set to
// slog.LevelInfo, but it can be overridden in the HandlerOptions. The package
// also provides a Closer to manage the lifecycle of the log file writer,
// ensuring that it is closed properly when no longer needed.
//
// By default, the log file is created under ~/log/ with a name matching the
// pattern <program_name>.<timestamp>.<pid>.log. The log file path can be
// customized by setting the File field in HandlerOptions.
//
// The impetus for writing this module was to make it easy to create a
// multilogger that writes to stderr, while also writing to a log file when
// the log level is ERROR or higher (but only creates the log file if there is
// at least one error message). See the NewMultilogger() function for an
// example of how to create a multilogger. See the examples directory for more
// examples of how to use this package.
package logutil
