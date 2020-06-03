package rvm

import (
	"bufio"
	"os"
	"strings"
)

// GemfileLockParser represents a Gemfile.lock parser
type GemfileLockParser struct{}

// NewGemfileLockParser creates a new Gemfile.lock parser
func NewGemfileLockParser() GemfileLockParser {
	return GemfileLockParser{}
}

// ParseVersion looks for a Gemfile.lock file in a given path and, if it
// exists, parses it to find a string "RUBY VERSION" and returns the string
// in the next line minus the whitespace and the prefix string "ruby "
func (r GemfileLockParser) ParseVersion(path string) (string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", err
	}

	GemfileLock, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer GemfileLock.Close()

	scanner := bufio.NewScanner(GemfileLock)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "RUBY VERSION" {
			if scanner.Scan() {
				return strings.TrimSpace(strings.Trim(scanner.Text(), "ruby ")), nil
			}
		}
	}

	return "", nil
}
