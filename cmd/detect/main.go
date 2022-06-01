package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-cnb/rvm"

	"github.com/paketo-buildpacks/packit/v2"
)

func main() {
	logger := rvm.NewLogEmitter(os.Stdout)
	bundlerVersionParser := bundler.NewBundlerVersionParser()
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	packit.Detect(bundler.Detect(logger, bundlerVersionParser, buildpackYMLParser))
}
