package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-cnb/rvm"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fs"
)

func main() {
	logEmitter := rvm.NewLogEmitter(os.Stdout)
	vr := bundler.NewRubyVersionResolver()
	calc := fs.NewChecksumCalculator()
	packit.Build(bundler.Build(logEmitter, vr, calc))
}
