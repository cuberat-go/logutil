package logutil

import (
	// Built-in/core modules.
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"testing"

	// Third-party modules.
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	topDir := t.TempDir()

	opts := &HandlerOptions{
		topDir: topDir,
	}
	t.Run("default path", func(st *testing.T) {
		runHandlerTest(st, opts)
	})

	tmpFile := t.TempDir() + "/test.log"
	opts = &HandlerOptions{
		File: tmpFile,
	}
	t.Run("custom path", func(st *testing.T) {
		runHandlerTest(st, opts)
	})
}

func runHandlerTest(t *testing.T, opts *HandlerOptions) {
	topDir := opts.topDir
	if opts.File != "" {
		topDir = path.Dir(opts.File)
	}

	handler, closer := NewErrorHandler(opts)
	defer closer.Close()
	logger := slog.New(handler)

	logger.Debug("This is a debug message", "key", "value")
	logger.Info("This is an info message", "key", "value")

	expected := []map[string]any{}
	logger2 := logger
	logger2.Error("This is a simple error message", "foo", "bar")
	expected = append(expected, map[string]any{
		"foo": "bar",
	})

	logger2 = logger.With("top_attr1", "value1", "top_attr2", 123)
	logger2.Error("This is a simple error message", "foo", "bar")
	expected = append(expected, map[string]any{
		"top_attr1": "value1",
		"top_attr2": float64(123),
		"foo":       "bar",
	})

	logger2 = logger.With("top_attr1", "value1", "top_attr2", 123).
		WithGroup("group1")
	logger2.Error("This is an error message", "foo", "bar")
	expected = append(expected, map[string]any{
		"top_attr1": "value1",
		"top_attr2": float64(123),
		"group1": map[string]any{
			"foo": "bar",
		},
	})

	logger2 = logger.With("top_attr1", "value1", "top_attr2", 123).
		WithGroup("group1").With("group1_attr1", "value1", "group1_attr2", 456)
	logger2.Error("This is an error message", "foo", "bar")
	expected = append(expected, map[string]any{
		"top_attr1": "value1",
		"top_attr2": float64(123),
		"group1": map[string]any{
			"group1_attr1": "value1",
			"group1_attr2": float64(456),
			"foo":          "bar",
		},
	})

	logger2 = logger.With("top_attr1", "value1", "top_attr2", 123).
		WithGroup("group1").With("group1_attr1", "value1", "group1_attr2", 456).
		WithGroup("group2").With("group2_attr1", "value1", "group2_attr2", 789)
	logger2.Error("This is an error message", "foo", "bar")
	expected = append(expected, map[string]any{
		"top_attr1": "value1",
		"top_attr2": float64(123),
		"group1": map[string]any{
			"group1_attr1": "value1",
			"group1_attr2": float64(456),
			"group2": map[string]any{
				"group2_attr1": "value1",
				"group2_attr2": float64(789),
				"foo":          "bar",
			},
		},
	})

	dirs := []string{}

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	}

	fsys := os.DirFS(topDir)
	err := fs.WalkDir(fsys, ".", walkFunc)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	if len(dirs) != 1 {
		t.Errorf("Expected 1 file, got %d", len(dirs))
		return
	}

	t.Logf("Log file found: %s", dirs[0])

	content, err := fs.ReadFile(fsys, dirs[0])
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	t.Logf("Log file content: %s", string(content))

	fh, err := fsys.Open(dirs[0])
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)
	i := 0
	for dec.More() {
		var record map[string]any
		err := dec.Decode(&record)
		if err != nil {
			t.Errorf("%s", err)
			return
		}

		if i >= len(expected) {
			t.Errorf("Unexpected log record: %v", record)
			return
		}

		if !assert.Equal(t, record["data"].(map[string]any), expected[i]) {
			t.Errorf("Log record %d does not match expected.\n"+
				"Got: %v\nExpected: %v",
				i, record["data"].(map[string]any), expected[i])
			return
		}

		i++
	}
}

func xTestStdHandler(t *testing.T) {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	logger := slog.New(slog.NewJSONHandler(
		os.Stderr, &slog.HandlerOptions{Level: logLevel}),
	)

	logger.Error("This is an error message",
		"msg1", map[string]any{"key1": "value1", "key2": 42},
		"msg2", "foo")
	logger.Error("This is another error message", "foo", "bar")

	attrLogger := logger.With("attr1", "value1", "attr2", 123)
	attrLogger.Error("This is an error message with attributes", "foo", "bar")

	groupLogger := logger.WithGroup("group1").WithGroup("group2")
	groupLogger.Error("This is an error message with groups", "b33f", "c0de")

	attrGroupLogger := attrLogger.WithGroup("group1").WithGroup("group2")
	attrGroupLogger.Error("This is an error message with attributes "+
		"and groups", "b33f", "c0de")

	groupLoggerWithAttr := groupLogger.With("gattr1", "value1", "gattr2", 123)
	groupLoggerWithAttr.Error("This is an error message with groups "+
		"and attributes", "b33f", "c0de")
	group2Logger := groupLoggerWithAttr.WithGroup("group3")
	group2Logger.Error("This is an error message with groups and "+
		"attributes", "b33f", "c0de")
}
