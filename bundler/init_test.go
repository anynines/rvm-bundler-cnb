package bundler_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBundler(t *testing.T) {
	suite := spec.New("bundler", spec.Report(report.Terminal{}))
	suite("Configuration", testConfiguration)
	suite("BundlerVersionParser", testBundlerVersionParser)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
