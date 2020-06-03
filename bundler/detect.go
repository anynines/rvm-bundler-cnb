package bundler

import (
	"os"
	"path/filepath"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/paketo-buildpacks/packit"
)

// VersionParser represents a parser for files like .ruby-version and Gemfiles
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

// BuildPlanMetadata represents this buildpack's metadata
type BuildPlanMetadata struct {
	BundlerVersion string `toml:"bundler_version"`
}

// Detect whether this buildpack should install RVM
func Detect(logger rvm.LogEmitter, bundlerVersionParser VersionParser, buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, "Gemfile"))
		if os.IsNotExist(err) {
			return packit.DetectResult{}, err
		}

		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		bundlerVersion := configuration.DefaultBundlerVersion

		// NOTE: the order of the parsers is important, the last one to return a
		// ruby version string "wins"
		versionEnvs := []rvm.VersionParserEnv{
			{
				Parser:  bundlerVersionParser,
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
			err = rvm.ParseVersion(env, &bundlerVersion)
			if err != nil {
				logger.Detail("Parsing '%s' failed", env.Path)
				return packit.DetectResult{}, err
			}
		}

		logger.Detail("Detected Bundler version: %s", bundlerVersion)
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm-bundler"},
				},
				Requires: []packit.BuildPlanRequirement{
					// TODO: programmatically add a dependency to the RVM CNB
					// {
					// 	Name:    "rvm",
					// 	Version: "1.29.10",
					// 	Metadata: rvm.BuildPlanMetadata{
					// 		RubyVersion: "2.5.3",
					// 	},
					// },
					{
						Name:    "rvm-bundler",
						Version: bundlerVersion,
					},
				},
			},
		}, nil
	}
}
