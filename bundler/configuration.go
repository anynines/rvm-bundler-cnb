package bundler

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Puma represents the configuration structure for Puma
type Puma struct {
	Version string `toml:"version"`
	Bind    string `toml:"bind"`
	Workers string `toml:"workers"`
	Threads string `toml:"threads"`
	Preload bool   `toml:"preload"`
}

// Configuration represents this buildpack's configuration read from a table
// named "configuration"
type Configuration struct {
	DefaultBundlerVersion string `toml:"default_bundler_version"`
	InstallPuma           bool   `toml:"install_puma"`
	Puma                  Puma   `toml:"puma"`
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
