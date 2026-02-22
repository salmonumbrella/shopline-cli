package batch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	MaxInputSize = 10 * 1024 * 1024 // 10MB
	MaxItems     = 10000
)

// Result represents the result of a batch operation.
type Result struct {
	ID      string `json:"id,omitempty"`
	Index   int    `json:"index"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ReadItems reads items from a file or stdin.
func ReadItems(filename string) ([]json.RawMessage, error) {
	var r io.Reader
	if filename == "" || filename == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	return ReadItemsFromReader(r)
}

// ReadItemsFromReader reads items from a reader.
func ReadItemsFromReader(r io.Reader) ([]json.RawMessage, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxInputSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Try JSON array first
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err == nil {
		if len(items) > MaxItems {
			return nil, fmt.Errorf("too many items: %d (max %d)", len(items), MaxItems)
		}
		return items, nil
	}

	// Try NDJSON
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	items = nil
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		items = append(items, json.RawMessage(line))
		if len(items) > MaxItems {
			return nil, fmt.Errorf("too many items: max %d", MaxItems)
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found in input")
	}

	return items, nil
}

// WriteResults writes results to a writer.
func WriteResults(w io.Writer, results []Result) error {
	enc := json.NewEncoder(w)
	for _, r := range results {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
