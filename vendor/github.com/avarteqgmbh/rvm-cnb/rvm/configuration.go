package rvm

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Configuration represents this buildpack's configuration read from a table
// named "configuration"
type Configuration struct {
	URI                string `toml:"uri"`
	DefaultRVMVersion  string `toml:"default_rvm_version"`
	DefaultRubyVersion string `toml:"default_ruby_version"`
	DefaultNodeVersion string `toml:"default_node_version"`
	DefaultRequireNode bool   `toml:"default_require_node"`
}

// MetaData represents this buildpack's metadata
type MetaData struct {
	Metadata struct {
		Configuration Configuration `toml:"configuration"`
	} `toml:"metadata"`
}

// ReadConfiguration returns the configuration for this buildpack
func ReadConfiguration(cnbPath string) (Configuration, error) {
	file, err := os.Open(filepath.Join(cnbPath, "buildpack.toml"))
	if err != nil {
		return Configuration{}, err
	}

	var meta MetaData
	_, err = toml.DecodeReader(file, &meta)
	if err != nil {
		return Configuration{}, err
	}

	return meta.Metadata.Configuration, nil
}
