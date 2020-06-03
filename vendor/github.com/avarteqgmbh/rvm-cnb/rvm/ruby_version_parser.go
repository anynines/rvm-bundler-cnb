package rvm

import (
	"io/ioutil"
	"os"
	"strings"
)

// RubyVersionParser represents a ruby version parser
type RubyVersionParser struct{}

// NewRubyVersionParser creates a new ruby version parser
func NewRubyVersionParser() RubyVersionParser {
	return RubyVersionParser{}
}

// ParseVersion looks for a .ruby-version file in a given path and, if it
// exists, parses it and removes trailing whitespace
func (r RubyVersionParser) ParseVersion(path string) (string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", err
	}

	rvFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer rvFile.Close()

	bytes, err := ioutil.ReadAll(rvFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}
