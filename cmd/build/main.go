package main

import (
	"os"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	vr := bundler.NewRubyVersionResolver()
	calc := fs.NewChecksumCalculator()
	bc := bundler.NewRunBashCmd()
	pm := bundler.NewPumaInstaller()
	packit.Build(bundler.Build(logger, vr, calc, bc, pm))
}
