package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-cnb/rvm"

	"github.com/paketo-buildpacks/packit"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	bundlerVersionParser := bundler.NewBundlerVersionParser()
	buildpackYMLParser := bundler.NewBuildpackYMLParser()
	packit.Detect(bundler.Detect(logEmitter, bundlerVersionParser, buildpackYMLParser))
}
