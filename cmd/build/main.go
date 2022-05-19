package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fs"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logEmitter := scribe.NewLogger(os.Stdout)
	vr := bundler.NewRubyVersionResolver()
	calc := fs.NewChecksumCalculator()
	bc := bundler.NewRunBashCmd()
	pm := bundler.NewPumaInstaller()
	packit.Build(bundler.Build(logEmitter, vr, calc, bc, pm))
}
