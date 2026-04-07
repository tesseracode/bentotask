// Package store handles reading and writing BentoTask data to disk.
//
// Tasks are stored as Markdown files with YAML frontmatter (ADR-002).
// This package provides the I/O layer between the in-memory model.Task
// and the filesystem representation.
package store

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"

	"github.com/tesserabox/bentotask/internal/model"
)

// ParseFile reads a task from a Markdown file on disk.
// It parses the YAML frontmatter into a model.Task and captures
// the Markdown body (everything after the closing ---).
func ParseFile(path string) (*model.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return Parse(f)
}

// Parse reads a task from any io.Reader containing Markdown with
// YAML frontmatter. This is the core parsing function — ParseFile
// is a convenience wrapper.
func Parse(r io.Reader) (*model.Task, error) {
	var task model.Task

	// frontmatter.Parse reads the YAML between --- delimiters,
	// unmarshals it into task, and returns the remaining body bytes.
	body, err := frontmatter.Parse(r, &task)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	task.Body = strings.TrimSpace(string(body))

	return &task, nil
}

// Marshal converts a model.Task back into Markdown with YAML frontmatter.
// The output format matches what Parse expects:
//
//	---
//	id: 01JQX00001
//	title: My Task
//	...
//	---
//
//	Body content here.
func Marshal(task *model.Task) ([]byte, error) {
	var buf bytes.Buffer

	// Marshal the frontmatter (everything except Body, which has yaml:"-")
	yamlBytes, err := yaml.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n")

	if task.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(task.Body)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// WriteFile writes a task to disk as a Markdown file with YAML frontmatter.
// It uses atomic writes (write to temp file, then rename) to prevent
// corruption if the process is interrupted mid-write.
func WriteFile(path string, task *model.Task) error {
	data, err := Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Atomic write: write to temp file, then rename.
	// This prevents corruption if the process crashes mid-write.
	// os.Rename is atomic on POSIX systems (same filesystem).
	tmpPath := filepath.Join(dir, ".tmp-write-"+filepath.Base(path))

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure (best-effort)
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}

	return nil
}
