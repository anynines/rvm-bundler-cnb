package bundler

import (
	"fmt"
	"regexp"
	"strings"
)

// RubyVersionResolver identifies and compares versions of Ruby used in the
// build environment.
type RubyVersionResolver struct {
}

// NewRubyVersionResolver initializes an instance of RubyVersionResolver.
func NewRubyVersionResolver() RubyVersionResolver {
	return RubyVersionResolver{}
}

// Lookup returns the version of Ruby installed in the build environment.
func (r RubyVersionResolver) Lookup(workingDir string, bashcmd BashCmd) (string, error) {
	getRubyVersionCmd := strings.Join([]string{
		"rvm",
		"current",
	}, " ")
	cmdStdOut, err := bashcmd.RunBashCmd(getRubyVersionCmd, workingDir)
	if err != nil {
		return "", fmt.Errorf("failed to obtain ruby version: %w: %s", err, cmdStdOut)
	}

	var versions []string
	regexes := []string{
		`(jruby-\d+\.\d+\.\d+).\d+`,
		`(jruby-\d+\.\d+)\.\d+`,
		`(jruby-head)`,
		`(ruby-\d+\.\d+)\.\d+`,
		`(ruby-head)`,
	}
	for _, s := range regexes {
		versions = regexp.MustCompile(s).FindStringSubmatch(cmdStdOut)
		if versions != nil {
			return versions[1], nil
		}
	}
	return "", fmt.Errorf("no string with ruby version found")
}
