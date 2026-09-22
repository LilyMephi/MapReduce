// Package parser converts a single raw log line into a structured Record.
package parser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors returned by Parse. Callers may test them with errors.Is.
var (
	ErrEmptyLine         = errors.New("empty line")
	ErrInvalidFieldCount = errors.New("invalid field count")
	ErrInvalidStatus     = errors.New("invalid status code")
	ErrInvalidTimestamp  = errors.New("invalid timestamp")
	ErrMissingIP         = errors.New("missing ip address")
)

// Record is one parsed log entry.
type Record struct {
	IP        string
	Timestamp time.Time
	Status    int
	Path      string
}

// Parse turns a raw CSV line ("IP,timestamp,status_code,path") into a Record.
// Empty (or whitespace-only) lines yield ErrEmptyLine so callers can skip them.
func Parse(line string) (Record, error) {
	var rec Record

	if strings.TrimSpace(line) == "" {
		return rec, ErrEmptyLine
	}

	r := csv.NewReader(strings.NewReader(line))
	r.FieldsPerRecord = 4

	fields, err := r.Read()
	if err != nil {
		if errors.Is(err, csv.ErrFieldCount) {
			return rec, fmt.Errorf("parse %q: %w", line, ErrInvalidFieldCount)
		}
		return rec, fmt.Errorf("parse %q: %w", line, err)
	}

	ip := strings.TrimSpace(fields[0])
	if ip == "" {
		return rec, fmt.Errorf("parse %q: %w", line, ErrMissingIP)
	}

	status, err := strconv.Atoi(strings.TrimSpace(fields[2]))
	if err != nil {
		return rec, fmt.Errorf("parse status %q: %w", fields[2], ErrInvalidStatus)
	}

	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[1]))
	if err != nil {
		return rec, fmt.Errorf("parse timestamp %q: %w", fields[1], ErrInvalidTimestamp)
	}

	rec = Record{
		IP:        ip,
		Timestamp: ts,
		Status:    status,
		Path:      strings.TrimSpace(fields[3]),
	}
	return rec, nil
}
