package quotes

import (
	"bufio"
	"os"
	"strings"
)

// Quote represents a quote with its author.
type Quote struct {
	Author string
	Text   string
}

// LoadQuotes reads quotes from a file in the format: Author<TAB>Quote
func LoadQuotes(path string) ([]Quote, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var quotes []Quote
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue // skip malformed lines
		}
		quotes = append(quotes, Quote{
			Author: parts[0],
			Text:   parts[1],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return quotes, nil
}
