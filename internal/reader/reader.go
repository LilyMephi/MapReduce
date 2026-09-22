// Package reader performs sequential, single-threaded reading of log files.
package reader

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"logcounter/internal/parser"
)

// maxLineSize is the scanner buffer size (64 MB) for unusually long lines.
const maxLineSize = 1024 * 1024 * 64

// ReadAll reads the given files one after another (no concurrency), parses each
// line and invokes handler for every successfully parsed Record. Empty lines are
// skipped. The first non-nil error returned by handler or encountered while
// reading stops the traversal.
func ReadAll(paths []string, handler func(parser.Record) error) error {
	for _, path := range paths {
		if err := readFile(path, handler); err != nil {
			return err
		}
	}
	return nil
}

func ReadPart(paths []string, handler func(parser.Record) error) error {
	for _, path := range paths {
		if err := readFile(path, handler); err != nil {
			return err
		}
	}
	return nil
}

func readFile(path string, handler func(parser.Record) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, maxLineSize), maxLineSize)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		rec, err := parser.Parse(scanner.Text())
		if err != nil {
			if errors.Is(err, parser.ErrEmptyLine) {
				continue
			}
			return fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
		if err := handler(rec); err != nil {
			return fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %q: %w", path, err)
	}
	return nil
}
