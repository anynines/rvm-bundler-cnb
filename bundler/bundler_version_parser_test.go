package bundler_test

import (
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

			workingDir, err = os.MkdirTemp("", "workingDir")
			Expect(err).NotTo(HaveOccurred())

			gemfileLock, err := os.ReadFile("../test/fixtures/Gemfile.lock")
			Expect(err).NotTo(HaveOccurred())

			gemFileLockPath := filepath.Join(workingDir, "Gemfile.lock")
			err = os.WriteFile(gemFileLockPath, gemfileLock, 0644)
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
	context("when a Gemfile.lock is couldn't be opened", func() {
		it("returns the error on file Gemfile.lock open failed", func() {
			_, err := bundlerVersionParser.ParseVersion(filepath.Join(workingDir, "../test/fixtures/nonexistent-file.lock"))
			Expect(err).To(HaveOccurred())
		})
	})
	context("when a Gemfile.lock is present but empty", func() {
		it.Before(func() {
			var err error

			workingDir, err = os.MkdirTemp("", "workingDir")
			Expect(err).NotTo(HaveOccurred())

			gemFileLockPath := filepath.Join(workingDir, "Gemfile.lock")
			err = os.WriteFile(gemFileLockPath, []byte(""), 0644)
			Expect(err).NotTo(HaveOccurred())

			bundlerVersionParser = bundler.NewBundlerVersionParser()
		})

		it("returns the error after parsing Gemfile.lock", func() {
			bundlerVersion, err := bundlerVersionParser.ParseVersion(filepath.Join(workingDir, "Gemfile.lock"))
			Expect(err).NotTo(HaveOccurred())
			Expect(bundlerVersion).To(Equal(""))
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})
	})
}
