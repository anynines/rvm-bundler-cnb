package bundler_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/avarteqgmbh/rvm-cnb/rvm/fakes"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cnbDir     string
		workingDir string

		bundlerVersionParser *fakes.VersionParser
		buildpackYMLParser   *fakes.VersionParser
		detect               packit.DetectFunc
	)

	it.Before(func() {
		bundlerVersionParser = &fakes.VersionParser{}
		buildpackYMLParser = &fakes.VersionParser{}

		logEmitter := rvm.NewLogEmitter(os.Stdout)
		detect = bundler.Detect(logEmitter, bundlerVersionParser, buildpackYMLParser)
	})

	it("returns a plan that does not provide RVM bundler because no Gemfile was found", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).To(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{Provides: nil, Requires: nil, Or: nil}))
	})

	context("when the app presents a Gemfile", func() {
		it.Before(func() {
			var err error
			cnbDir, err = ioutil.TempDir("", "cnb")
			Expect(err).NotTo(HaveOccurred())

			someBuildPackTomlFile, err := ioutil.ReadFile("../test/fixtures/some_buildpack.toml")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			gemFile, err := ioutil.ReadFile("../test/fixtures/Gemfile")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, "Gemfile"), gemFile, 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a plan that provides RVM Bundler", func() {
			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm-bundler"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:    "rvm-bundler",
						Version: "2.1.4",
					},
				},
			}))
		})

		it("returns a plan that provides RVM, requires node and determines the ruby version by reading Gemfile.lock", func() {
			gemfileLock, err := ioutil.ReadFile("../test/fixtures/Gemfile.lock")
			Expect(err).NotTo(HaveOccurred())

			gemFileLockPath := filepath.Join(workingDir, "Gemfile.lock")
			err = ioutil.WriteFile(gemFileLockPath, gemfileLock, 0644)
			Expect(err).NotTo(HaveOccurred())

			bundlerVersionParser.ParseVersionCall.Receives.Path = gemFileLockPath
			bundlerVersionParser.ParseVersionCall.Returns.Version = "2.1.4"

			result, err := detect(packit.DetectContext{
				CNBPath:    cnbDir,
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "rvm-bundler"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:    "rvm-bundler",
						Version: "2.1.4",
					},
				},
			}))
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
		})
	})
}
