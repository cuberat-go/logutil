package logutil

import (
	// Built-in/core modules.
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

// Closer is a structure that information used to close the log file writer. It
// ensures that only one goroutine can close writer (only one call to Close
// will actually call Close() on the underlying writer).
type Closer struct {
	writerCloser io.WriteCloser
	mutex        *sync.Mutex
}

// Closes the writer if it is open. Close is part of the io.Closer interface.
func (c *Closer) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.writerCloser != nil {
		err := c.writerCloser.Close()
		c.writerCloser = nil
		return err
	}
	return nil
}

// Structure for options to configure the Handler.
type HandlerOptions struct {
	// Top-level directory for log files if a file path is not specified. If
	// empty, defaults to the user's home directory.
	topDir string

	// Path to the log file. If empty, a default path is used.
	File string

	// Log level. Note this is a pointer to a slog.LevelVar, which allows
	// dynamic changes to the log level at runtime.
	LogLevel *slog.LevelVar
}

type groupInfo struct {
	Attrs []slog.Attr
	Name  string
}

// Structure for the log handler that implements slog.Handler interface.
type Handler struct {
	attrs     []slog.Attr
	groups    []string
	mutex     *sync.Mutex
	createdAt time.Time
	closer    *Closer
	options   *HandlerOptions
	logLevel  *slog.LevelVar
	groupData []*groupInfo
}

// Returns a new Handler with the given options and a Closer for closing the log
// file. The log file created on the first write. The log level is set to
// slog.LevelInfo if not overridden in opts.
func NewHandler(opts *HandlerOptions) (*Handler, io.Closer) {
	h := newHandler([]slog.Attr{}, []string{}, &sync.Mutex{}, opts)
	return h, h.closer
}

// Returns a new Handler with the given options and a Closer for closing the log
// file. The log file is created on the first write. The log level is set to
// slog.LevelError, regardless of the value in opts.
func NewErrorHandler(opts *HandlerOptions) (*Handler, io.Closer) {
	var newOpts *HandlerOptions
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelError)

	if opts == nil {
		newOpts = &HandlerOptions{LogLevel: logLevel}
	} else {
		newOpts = &HandlerOptions{}
		*newOpts = *opts
		newOpts.LogLevel = logLevel
	}
	h := newHandler([]slog.Attr{}, []string{}, &sync.Mutex{}, newOpts)
	return h, h.closer
}

// Returns a new Handler with the given attributes, groups, mutex, and options.
func newHandler(
	attrs []slog.Attr,
	groups []string,
	mutex *sync.Mutex,
	opts *HandlerOptions,
) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	logLevel := opts.LogLevel
	if logLevel == nil {
		logLevel = new(slog.LevelVar)
		logLevel.Set(slog.LevelInfo)
	}
	return &Handler{
		attrs:     attrs,
		groups:    groups,
		mutex:     mutex,
		createdAt: time.Now(),
		closer:    &Closer{mutex: &sync.Mutex{}},
		options:   opts,
		logLevel:  logLevel,
		groupData: []*groupInfo{{Name: "data"}},
	}
}

// Returns a new Handler that is a clone of the receiver. The clone has the same
// attributes, groups, mutex, options, and log level as the receiver. The clone
// also shares the same Closer containing the writer.
func (h *Handler) clone() *Handler {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	newAttrs := append([]slog.Attr{}, h.attrs...)
	newGroups := append([]string{}, h.groups...)
	newHandler := newHandler(newAttrs, newGroups, h.mutex, h.options)
	newHandler.closer = h.closer
	newHandler.logLevel = h.logLevel
	newHandler.groupData = append([]*groupInfo{}, h.groupData...)

	return newHandler
}

// Formats the log record as a JSON object and writes it to the log file. Handle
// is part of the slog.Handler interface
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	ts := r.Time.UTC().Format("2006-01-02T15:04:05.999999Z")
	src := map[string]any{}

	if r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		src["file"] = frame.File
		src["line"] = frame.Line
		src["func"] = path.Base(frame.Function)
	}

	recordAttrs := map[string]any{}

	addAttrs := func(attr slog.Attr) bool {
		recordAttrs[attr.Key] = attr.Value.Any()
		return true
	}

	r.Attrs(addAttrs)

	msg := map[string]any{
		"ts":  ts,
		"lvl": r.Level.String(),
		"src": src,
		"msg": r.Message,
	}

	outBytes := []byte{}
	outBytes = append(outBytes, '{')

	first := true
	for _, field := range []string{"ts", "lvl", "src", "msg"} {
		if !first {
			outBytes = append(outBytes, ',')
		}
		first = false
		fieldBytes, err := json.Marshal(field)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal field name: %v\n", err)
			return err
		}
		outBytes = append(outBytes, fieldBytes...)
		outBytes = append(outBytes, ':')
		valueBytes, err := json.Marshal(msg[field])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal field value: %v\n", err)
			return err
		}
		outBytes = append(outBytes, valueBytes...)
	}

	var groupData map[string]any
	lastGroupName := ""
	lastGroupData := map[string]any{}
	for i := len(h.groupData) - 1; i >= 0; i-- {
		thisGroup := h.groupData[i]

		groupData = map[string]any{}

		for _, attr := range thisGroup.Attrs {
			groupData[attr.Key] = attr.Value.Any()
		}

		if lastGroupName == "" {
			maps.Copy(groupData, recordAttrs)
		} else {
			groupData[lastGroupName] = lastGroupData
		}

		lastGroupData = groupData
		lastGroupName = thisGroup.Name
	}

	dataBytes, err := json.Marshal(groupData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log message: %v\n", err)
		return err
	}
	outBytes = append(outBytes, ',')
	outBytes = append(outBytes, []byte(`"data":`)...)
	outBytes = append(outBytes, dataBytes...)

	outBytes = append(outBytes, '}')

	outBytes = append(outBytes, '\n')

	h.mutex.Lock()
	defer h.mutex.Unlock()

	writer, err := h.getWriter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get writer: %v\n", err)
		return err
	}
	_, err = writer.Write(outBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write log message: %v\n", err)
		return err
	}

	return nil
}

// Reports whether the handler handles records at the given level. Enable is
// part of the slog.Handler interface
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.logLevel.Level()
}

// Returns a new Handler whose attributes consist of both the receiver's
// attributes and the arguments. WithAtrs is part of the slog.Handler
// interface
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := h.clone()
	// newHandler.attrs = append(newHandler.attrs, attrs...)
	groupData := newHandler.groupData[len(newHandler.groupData)-1]
	groupData.Attrs = append(groupData.Attrs, attrs...)
	return newHandler
}

// Returns a new Handler with the given group appended to the receiver's
// existing groups. WithGroup is part of the slog.Handler interface
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newHandler := h.clone()
	newHandler.groups = append(newHandler.groups, name)
	newHandler.groupData =
		append(newHandler.groupData, &groupInfo{Name: name})

	return newHandler
}

// Returns the writer for the log file, opening it if necessary.
func (h *Handler) getWriter() (io.Writer, error) {
	if h.closer.writerCloser == nil {
		writer, err := h.openFile()
		if err != nil {
			return nil, err
		}
		h.closer.writerCloser = writer
	}
	return h.closer.writerCloser, nil
}

// Generates a log file name based on the program name, current timestamp, and
// process ID. The format is: <program_name>.<timestamp>.<pid>.log
func (h *Handler) genFileName() string {
	prog := path.Base(os.Args[0])
	ts := time.Now().UTC().Format("2006-01-02T150405Z")
	fileName := fmt.Sprintf("%s.%s.%d.log", prog, ts, os.Getpid())
	return fileName
}

func (h *Handler) openFile() (io.WriteCloser, error) {
	if h.options.File == "" {
		return h.openFileDefault()
	}

	logDir := path.Dir(h.options.File)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	logFH, err := os.OpenFile(h.options.File,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return logFH, nil
}

func (h *Handler) openFileDefault() (io.WriteCloser, error) {
	if h.options != nil && h.options.topDir != "" {
		return h.openFileBase(h.options.topDir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	return h.openFileBase(home)
}

// Opens an opinionated log file path based on the given top-level directory.
// The log file is created in a subdirectory named "log" under the top-level
// directory, and the log file name is generated based on the program name,
// current timestamp, and process ID. If the log directory does not exist, it
// is created with appropriate permissions.
func (h *Handler) openFileBase(
	topDir string,
) (io.WriteCloser, error) {
	prog := path.Base(os.Args[0])
	logDir := path.Join(topDir, "log", prog)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	fileName := h.genFileName()
	logFilePath := path.Join(logDir, fileName)
	logFH, err := os.OpenFile(logFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return logFH, nil
}
