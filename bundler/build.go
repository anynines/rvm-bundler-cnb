package bundler

import (
	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fs"
)

// Build the RVM layer provided by this buildpack
func Build(logger rvm.LogEmitter, vr RubyVersionResolver, calc fs.ChecksumCalculator) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}
		return InstallBundler(context, configuration, logger, vr, calc)
	}
}
