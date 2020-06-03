package bundler

import (
	"bufio"
	"os"
	"strings"
)

// BundlerVersionParser represents a Gemfile.lock parser
type BundlerVersionParser struct{}

// NewBundlerVersionParser creates a new Gemfile.lock parser
func NewBundlerVersionParser() BundlerVersionParser {
	return BundlerVersionParser{}
}

// ParseVersion looks for a Gemfile.lock file in a given path and, if it
// exists, parses it to find a string "BUNDLED WITH" and returns the string in
// the next line minus the whitespace
func (r BundlerVersionParser) ParseVersion(path string) (string, error) {
	Bundler, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer Bundler.Close()

	scanner := bufio.NewScanner(Bundler)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "BUNDLED WITH" {
			if scanner.Scan() {
				return strings.TrimSpace(scanner.Text()), nil
			}
		}
	}

	return "", nil
}
