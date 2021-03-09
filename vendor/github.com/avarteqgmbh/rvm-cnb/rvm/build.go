package rvm

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

// EnvironmentConfiguration represents an environment and a path to the RVM
// layer
type EnvironmentConfiguration interface {
	Configure(env packit.Environment, path string) error
}

// Build the RVM layer provided by this buildpack
func Build(environment EnvironmentConfiguration, logger LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}

		buildPackYMLPath := filepath.Join(context.WorkingDir, "buildpack.yml")
		buildPackYML, err := BuildpackYMLParse(buildPackYMLPath)
		if err != nil {
			logger.Detail("Parsing '%s' failed", buildPackYMLPath)
			return packit.BuildResult{}, err
		}

		rvmEnv := Env{
			BuildPackYML:  buildPackYML,
			Configuration: configuration,
			Context:       context,
			Environment:   environment,
			Logger:        logger,
		}

		return rvmEnv.BuildRvm()
	}
}
