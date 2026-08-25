package logutil

import (
	// Built-in/core modules.
	"bytes"
	"encoding/json"
	"fmt"
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

	opts = &HandlerOptions{
		LogDir: t.TempDir(),
	}
	t.Run("custom top dir", func(st *testing.T) {
		runHandlerTestCustomDir(st, opts)
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

	records, err := getJSONFileContents(topDir)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	if len(records) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(records))
		return
	}

	for i, record := range records {
		if !assert.Equal(t, record["data"].(map[string]any), expected[i]) {
			t.Errorf("Log record %d does not match expected.\n"+
				"Got: %v\nExpected: %v",
				i, record["data"].(map[string]any), expected[i])
			return
		}
	}
}

func runHandlerTestCustomDir(t *testing.T, opts *HandlerOptions) {
	logDir := opts.LogDir
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

	records, err := getJSONFileContents(logDir)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	if len(records) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(records))
		return
	}

	for i, record := range records {
		if !assert.Equal(t, record["data"].(map[string]any), expected[i]) {
			t.Errorf("Log record %d does not match expected.\n"+
				"Got: %v\nExpected: %v",
				i, record["data"].(map[string]any), expected[i])
			return
		}
	}
}

func TestMarshalFunc(t *testing.T) {
	marshalFunc := func(v any) ([]byte, error) {
		d := map[string]any{
			"custom": v,
		}
		return json.Marshal(d)
	}

	topDir := t.TempDir()

	opts := &HandlerOptions{
		topDir:      topDir,
		MarshalFunc: marshalFunc,
	}

	handler, closer := NewErrorHandler(opts)
	defer closer.Close()
	logger := slog.New(handler)

	expected := []map[string]any{
		{
			"custom": map[string]any{
				"key1": "value1",
				"key2": float64(42),
			},
		},
	}

	logger.Error("test custom marshal func", "key1", "value1", "key2", 42)

	records, err := getJSONFileContents(topDir)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	if len(records) != len(expected) {
		t.Errorf("Expected %d records, got %d", len(expected), len(records))
		return
	}

	for i, record := range records {
		if !assert.Equal(t, record["data"].(map[string]any), expected[i]) {
			t.Errorf("Log record %d does not match expected.\n"+
				"Got: %v\nExpected: %v",
				i, record["data"].(map[string]any), expected[i])
			return
		}
	}
}

func getJSONFileContents(topDir string) ([]map[string]any, error) {
	fsys := os.DirFS(topDir)
	var dirs []string

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

	err := fs.WalkDir(fsys, ".", walkFunc)
	if err != nil {
		return nil, err
	}

	if len(dirs) != 1 {
		return nil, fmt.Errorf("expected 1 file, got %d", len(dirs))
	}

	content, err := fs.ReadFile(fsys, dirs[0])
	if err != nil {
		return nil, err
	}

	var records []map[string]any
	dec := json.NewDecoder(bytes.NewReader(content))
	for dec.More() {
		var record map[string]any
		err := dec.Decode(&record)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
