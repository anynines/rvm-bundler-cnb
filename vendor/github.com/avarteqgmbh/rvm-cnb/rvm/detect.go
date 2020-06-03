package rvm

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

// VersionParser represents a parser for files like .ruby-version and Gemfiles
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

// BuildPlanMetadata represents this buildpack's metadata
type BuildPlanMetadata struct {
	RubyVersion string `toml:"ruby_version"`
}

// NodebuildPlanMetadata represents the metadata for the node dependency
type NodebuildPlanMetadata struct {
	Build  bool `toml:"build"`
	Launch bool `toml:"launch"`
}

// VersionParserEnv represents an environment that contains everything that is
// needed to execute a particular ruby version parser
type VersionParserEnv struct {
	Parser  VersionParser
	Path    string
	Context packit.DetectContext
	Logger  LogEmitter
}

// ParseVersion is a generalized function that parses a particular ruby version
// source
func ParseVersion(env VersionParserEnv, version *string) error {
	fullPath := filepath.Join(env.Context.WorkingDir, env.Path)
	parseResultRubyVersion, err := env.Parser.ParseVersion(fullPath)
	if err == nil && parseResultRubyVersion != "" {
		*version = parseResultRubyVersion
		env.Logger.Detail("Found Ruby version in %s: %s", fullPath, *version)
		return nil
	}
	return err
}

// Detect whether this buildpack should install RVM
func Detect(logger LogEmitter, rubyVersionParser VersionParser, gemFileParser VersionParser, gemFileLockParser VersionParser, buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, "Gemfile"))
		if os.IsNotExist(err) {
			return packit.DetectResult{}, err
		}

		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		rubyVersion := configuration.DefaultRubyVersion

		// NOTE: the order of the parsers is important, the last one to return a
		// ruby version string "wins"
		versionEnvs := []VersionParserEnv{
			{
				Parser:  rubyVersionParser,
				Path:    ".ruby-version",
				Context: context,
				Logger:  logger,
			},
			{
				Parser:  gemFileParser,
				Path:    "Gemfile",
				Context: context,
				Logger:  logger,
			},
			{
				Parser:  gemFileLockParser,
				Path:    "Gemfile.lock",
				Context: context,
				Logger:  logger,
			},
			{
				Parser:  buildpackYMLParser,
				Path:    "buildpack.yml",
				Context: context,
				Logger:  logger,
			},
		}

		for _, env := range versionEnvs {
			err = ParseVersion(env, &rubyVersion)
			if err != nil && !os.IsNotExist(err) {
				logger.Detail("Parsing '%s' failed", env.Path)
				return packit.DetectResult{}, err
			}
		}
		logger.Detail("Detected Ruby version: %s", rubyVersion)

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "rvm",
				Metadata: BuildPlanMetadata{
					RubyVersion: rubyVersion,
				},
			},
		}

		buildPackYMLPath := filepath.Join(context.WorkingDir, "buildpack.yml")
		buildPackYML, err := BuildpackYMLParse(buildPackYMLPath)
		if err != nil {
			logger.Detail("Parsing '%s' failed", buildPackYMLPath)
			return packit.DetectResult{}, err
		}

		if buildPackYML.RequireNode {
			logger.Detail("The buildpack 'node' was requested as a requirement")
			nodeVersion := configuration.DefaultNodeVersion
			if buildPackYML.NodeVersion != "" {
				nodeVersion = buildPackYML.NodeVersion
			}
			logger.Detail("The buildpack requested node version: %s", nodeVersion)
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name:    "node",
				Version: nodeVersion,
				Metadata: NodebuildPlanMetadata{
					Build:  true,
					Launch: true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm"},
				},
				Requires: requirements,
			},
		}, nil
	}
}
