package quotes

import (
	"os"
	"testing"
)

func TestLoadQuotes(t *testing.T) {
	// Create a temporary file with test quotes
	tmpFile, err := os.CreateTemp("", "quotes_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Author1\tQuote one.\nAuthor2\tQuote two.\nMalformed line\nAuthor3\tQuote three."
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	quotes, err := LoadQuotes(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadQuotes failed: %v", err)
	}

	if len(quotes) != 3 {
		t.Errorf("Expected 3 quotes, got %d", len(quotes))
	}

	expected := []Quote{
		{"Author1", "Quote one."},
		{"Author2", "Quote two."},
		{"Author3", "Quote three."},
	}
	for i, q := range expected {
		if quotes[i] != q {
			t.Errorf("Quote %d mismatch: got %+v, want %+v", i, quotes[i], q)
		}
	}
}
