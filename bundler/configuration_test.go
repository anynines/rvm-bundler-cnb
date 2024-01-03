package bundler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testConfiguration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cnbDir            string
		buildPackTomlPath string = "../test/fixtures/some_buildpack.toml"
	)

	context("when the buildpack configuration is requested", func() {
		it.Before(func() {
			var err error
			cnbDir, err = os.MkdirTemp("", "cnb")
			Expect(err).NotTo(HaveOccurred())
		})

		it("cannot find the buildpack.toml file and returns an error", func() {
			_, err := bundler.ReadConfiguration(cnbDir)
			Expect(err).To(HaveOccurred())
		})

		it("reads an invalid buildpack.toml file and returns an error", func() {
			err := os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte("[[buildpack]"), 0644)
			Expect(err).NotTo(HaveOccurred())

			_, err = bundler.ReadConfiguration(cnbDir)
			Expect(err).To(HaveOccurred())
		})

		it("reads a valid buildpack.toml file without the [metadata.configuration] table and returns an empty configuration", func() {
			err := os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte("[buildpack]"), 0644)
			Expect(err).NotTo(HaveOccurred())

			configuration, _ := bundler.ReadConfiguration(cnbDir)
			Expect(configuration).To(Equal(bundler.Configuration{}))
		})

		it("reads the buildpack.toml file and returns the configuration", func() {
			someBuildPackTomlFile, err := os.ReadFile(buildPackTomlPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			configuration, err := bundler.ReadConfiguration(cnbDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(configuration).To(Equal(bundler.Configuration{
				DefaultBundlerVersion: "2.1.4",
				InstallPuma:           true,
				Puma: bundler.Puma{
					Version: "4.3.5",
					Bind:    "tcp://0.0.0.0:8080",
					Workers: "5",
					Threads: "5",
					Preload: true,
				},
			}))
		})

		it.After(func() {
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
		})
	})
}
