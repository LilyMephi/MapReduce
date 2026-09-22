// Command logcounter is the naive, single-threaded baseline: it reads web
// server logs sequentially and counts requests per IP address in memory.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"logcounter/internal/counter"
	"logcounter/internal/parser"
	"logcounter/internal/reader"
)

// progressEvery controls how often per-line progress is logged in verbose mode.
const progressEvery = 100000

func main() {
	if err := run(); err != nil {
		log.Fatalf("logcounter: %v", err)
	}
}

func run() error {
	input := flag.String("input", "", "glob pattern or directory containing *.log files")
	out := flag.String("out", "counts.json", "path to the JSON output file")
	verbose := flag.Bool("verbose", false, "log per-line progress")
	flag.Parse()

	if *input == "" {
		return errors.New("missing required flag: -input")
	}

	paths, err := resolvePaths(*input)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no input files matched %q", *input)
	}

	c := counter.New()
	var lines int

	for i, path := range paths {
		handler := func(rec parser.Record) error {
			c.Add(rec.IP)
			lines++
			if *verbose && lines%progressEvery == 0 {
				log.Printf("progress: files=%d lines=%d unique_ips=%d", i+1, lines, c.Len())
			}
			return nil
		}
		if err := reader.ReadAll([]string{path}, handler); err != nil {
			return err
		}
		log.Printf("progress: files=%d lines=%d unique_ips=%d", i+1, lines, c.Len())
	}

	if err := writeJSON(*out, c.Snapshot()); err != nil {
		return err
	}

	log.Printf("done: files=%d lines=%d unique_ips=%d out=%s", len(paths), lines, c.Len(), *out)
	return nil
}

// resolvePaths expands a directory into its *.log files or treats the argument
// as a glob pattern.
func resolvePaths(input string) ([]string, error) {
	info, err := os.Stat(input)
	switch {
	case err == nil && info.IsDir():
		return filepath.Glob(filepath.Join(input, "*.log"))
	case err == nil:
		return []string{input}, nil
	case errors.Is(err, os.ErrNotExist):
		return filepath.Glob(input)
	default:
		return nil, fmt.Errorf("stat %q: %w", input, err)
	}
}

func writeJSON(path string, data map[string]int) error {
	blob, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal counts: %w", err)
	}
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
