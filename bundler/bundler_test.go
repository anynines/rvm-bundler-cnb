package bundler_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	bundler "github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBundler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir      string
		layerPath       string
		versionResolver *fakes.VersionResolver
		calculator      *fakes.Calculator
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		layerPath, err = ioutil.TempDir("", "layer")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.RemoveAll(layerPath)).To(Succeed())

		versionResolver = &fakes.VersionResolver{}
		calculator = &fakes.Calculator{}
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(layerPath)).To(Succeed())
	})

	context("ShouldRun", func() {
		it.Before(func() {
			versionResolver.LookupCall.Returns.Version = "ruby-2.3.4"

			calculator.SumCall.Returns.String = "other-checksum"

			Expect(os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), nil, 0600)).To(Succeed())
		})

		it("indicates that the install process should run", func() {
			ok, checksum, rubyVersion, err := bundler.ShouldRun(map[string]interface{}{
				"cache_sha":    "some-checksum",
				"ruby_version": "ruby-1.2.3",
			}, workingDir,
				versionResolver,
				calculator)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(checksum).To(Equal("other-checksum"))
			Expect(rubyVersion).To(Equal("ruby-2.3.4"))

			Expect(versionResolver.LookupCall.CallCount).To(Equal(1))

			Expect(calculator.SumCall.Receives.Paths).To(Equal([]string{
				filepath.Join(workingDir, "Gemfile"),
				filepath.Join(workingDir, "Gemfile.lock"),
			}))
		})

		context("when the checksum matches, but the ruby version does not", func() {
			it.Before(func() {
				versionResolver.LookupCall.Returns.Version = "ruby-2.3.4"

				calculator.SumCall.Returns.String = "some-checksum"
			})

			it("indicates that the install process should run", func() {
				ok, checksum, rubyVersion, err := bundler.ShouldRun(map[string]interface{}{
					"cache_sha":    "some-checksum",
					"ruby_version": "ruby-1.2.3",
				}, workingDir,
					versionResolver,
					calculator)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())
				Expect(checksum).To(Equal("some-checksum"))
				Expect(rubyVersion).To(Equal("ruby-2.3.4"))
			})
		})

		context("when the checksum doesn't match, but the ruby version does", func() {
			it.Before(func() {
				versionResolver.LookupCall.Returns.Version = "jruby-1.2.3.4"

				calculator.SumCall.Returns.String = "other-checksum"
			})

			it("indicates that the install process should run", func() {
				ok, checksum, rubyVersion, err := bundler.ShouldRun(map[string]interface{}{
					"cache_sha":    "some-checksum",
					"ruby_version": "jruby-1.2.3.4",
				}, workingDir,
					versionResolver,
					calculator)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())
				Expect(checksum).To(Equal("other-checksum"))
				Expect(rubyVersion).To(Equal("jruby-1.2.3.4"))
			})
		})

		context("when the checksum and ruby version matches", func() {
			it.Before(func() {
				versionResolver.LookupCall.Returns.Version = "jruby-1.2.3"

				calculator.SumCall.Returns.String = "some-checksum"
			})

			it("indicates that the install process should not run", func() {
				ok, checksum, rubyVersion, err := bundler.ShouldRun(map[string]interface{}{
					"cache_sha":    "some-checksum",
					"ruby_version": "jruby-1.2.3",
				}, workingDir,
					versionResolver,
					calculator)
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
				Expect(checksum).To(Equal("some-checksum"))
				Expect(rubyVersion).To(Equal("jruby-1.2.3"))
			})
		})

		context("failure cases", func() {
			context("when the ruby version cannot be looked up", func() {
				it.Before(func() {
					versionResolver.LookupCall.Returns.Err = errors.New("failed to lookup ruby version")
				})

				it("returns an error", func() {
					_, _, _, err := bundler.ShouldRun(map[string]interface{}{
						"cache_sha":    "some-checksum",
						"ruby_version": "1.2.3",
					}, workingDir,
						versionResolver,
						calculator)
					Expect(err).To(MatchError("failed to lookup ruby version"))
				})
			})

			context("when the Gemfile.lock cannot be stat'd", func() {
				it.Before(func() {
					Expect(os.Chmod(workingDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, _, _, err := bundler.ShouldRun(map[string]interface{}{
						"cache_sha":    "some-checksum",
						"ruby_version": "ruby-1.2.3",
					}, workingDir,
						versionResolver,
						calculator)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when a checksum cannot be calculated", func() {
				it.Before(func() {
					calculator.SumCall.Returns.Error = errors.New("failed to calculate checksum")
				})

				it.After(func() {
					Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, _, _, err := bundler.ShouldRun(map[string]interface{}{
						"cache_sha":    "some-checksum",
						"ruby_version": "ruby-1.2.3",
					}, workingDir,
						versionResolver,
						calculator)
					Expect(err).To(MatchError("failed to calculate checksum"))
				})
			})
		})
	})
}
