package bundler

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fs"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Build the RVM layer provided by this buildpack
func Build(
	logger scribe.Logger,
	vr RubyVersionResolver,
	calc fs.ChecksumCalculator,
	bc RunBashCmd,
	pm PumaInstaller) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		configuration, err := ReadConfiguration(context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}
		return InstallBundler(context, configuration, logger, vr, calc, bc, pm)
	}
}
