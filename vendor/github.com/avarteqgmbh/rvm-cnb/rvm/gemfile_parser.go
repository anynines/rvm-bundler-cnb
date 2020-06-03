package rvm

import (
	"bufio"
	"os"
	"regexp"
)

// RubyVersionRegEx is a regular expression used to fin the ruby version in a
// call to the "ruby" method, see: https://bundler.io/man/gemfile.5.html#RUBY
const RubyVersionRegEx = `^ruby ["']([[:alnum:]\.\-]+)["'].*$`

// GemfileParser represents a Gemfile parser
type GemfileParser struct{}

// NewGemfileParser creates a new Gemfile parser
func NewGemfileParser() GemfileParser {
	return GemfileParser{}
}

// ParseVersion looks for a Gemfile file in a given path and, if it
// exists, parses it to find a ruby version spec
func (r GemfileParser) ParseVersion(path string) (string, error) {
	gemfile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer gemfile.Close()

	scanner := bufio.NewScanner(gemfile)
	re, err := regexp.Compile(RubyVersionRegEx)
	if err == nil {
		for scanner.Scan() {
			rubySubSlices := re.FindSubmatch([]byte(scanner.Text()))
			if rubySubSlices != nil {
				return string(rubySubSlices[1]), nil
			}
		}
	}

	return "", nil
}
