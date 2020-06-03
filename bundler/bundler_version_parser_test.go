package bundler_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBundlerVersionParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir           string
		bundlerVersionParser bundler.BundlerVersionParser
	)

	context("when a Gemfile.lock is present", func() {
		it.Before(func() {
			var err error

			workingDir, err = ioutil.TempDir("", "workingDir")
			Expect(err).NotTo(HaveOccurred())

			gemfileLock, err := ioutil.ReadFile("../test/fixtures/Gemfile.lock")
			Expect(err).NotTo(HaveOccurred())

			gemFileLockPath := filepath.Join(workingDir, "Gemfile.lock")
			err = ioutil.WriteFile(gemFileLockPath, gemfileLock, 0644)
			Expect(err).NotTo(HaveOccurred())

			bundlerVersionParser = bundler.NewBundlerVersionParser()
		})

		it("returns the bundler version after parsing Gemfile.lock", func() {
			bundlerVersion, err := bundlerVersionParser.ParseVersion(filepath.Join(workingDir, "Gemfile.lock"))
			Expect(err).NotTo(HaveOccurred())
			Expect(bundlerVersion).To(Equal("2.1.4"))
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})
	})
}
