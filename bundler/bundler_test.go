package bundler_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	bundler "github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBundler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir      string
		cnbDir          string
		layersDir       string
		buffer          *bytes.Buffer
		versionResolver *fakes.VersionResolver
		calculator      *fakes.Calculator
		bashCmd         *fakes.BashCmd
		pumainstaller   *fakes.PumaInstaller
		ctx             packit.BuildContext
		emptyBuffer     []byte
	)

	it.Before(func() {
		var err error

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		versionResolver = &fakes.VersionResolver{}
		calculator = &fakes.Calculator{}
		bashCmd = &fakes.BashCmd{}
		pumainstaller = &fakes.PumaInstaller{}
		emptyBuffer = []byte(``)

		someBuildPackTomlFile, err := ioutil.ReadFile("../buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
	})

	context("InstallBunler", func() {
		it.Before(func() {
			calculator.SumCall.Returns.String = "other-checksum"
			Expect(os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), nil, 0600)).To(Succeed())
		})

		it("returns a result", func() {
			ctx = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Layers:     packit.Layers{Path: layersDir},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "rvm-bundler",
							Metadata: map[string]interface{}{
								"rvm_bundler_version": "1.15.2",
							},
						},
					},
				},
			}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewLogger(buffer)
			configuration, _ := bundler.ReadConfiguration(ctx.CNBPath)

			// This line enables successfull exit from InstallPuma() (puma.go) call
			configuration.InstallPuma = false

			_, err := bundler.InstallBundler(ctx, configuration, logger, versionResolver, calculator, bashCmd, pumainstaller)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a result with creating `./bundle/config` file on the bundlerLayer", func() {

			err := os.MkdirAll(filepath.Join(workingDir, ".bundle"), 0700)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, ".bundle", "config"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Layers:     packit.Layers{Path: layersDir},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
			}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewLogger(buffer)
			configuration, _ := bundler.ReadConfiguration(ctx.CNBPath)

			// This line enables successfull exit from InstallPuma() (puma.go) call
			configuration.InstallPuma = false

			_, err = bundler.InstallBundler(ctx, configuration, logger, versionResolver, calculator, bashCmd, pumainstaller)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a result with creating `./bundle/config` file on the bundlerLayer with creating `./bundle/config.bak` backup", func() {

			err := os.MkdirAll(filepath.Join(workingDir, ".bundle"), 0700)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, ".bundle", "config"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(workingDir, ".bundle", "config.bak"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Layers:     packit.Layers{Path: layersDir},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
			}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewLogger(buffer)
			configuration, _ := bundler.ReadConfiguration(ctx.CNBPath)

			// This line enables successfull exit from InstallPuma() (puma.go) call
			configuration.InstallPuma = false

			_, err = bundler.InstallBundler(ctx, configuration, logger, versionResolver, calculator, bashCmd, pumainstaller)
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns an error if the version resolver fails to resolve the bundler version", func() {
			versionResolver.LookupCall.Returns.Err = errors.New("failed to obtain ruby version:")
			ctx = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Layers:     packit.Layers{Path: layersDir},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
			}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewLogger(buffer)
			configuration, _ := bundler.ReadConfiguration(ctx.CNBPath)

			_, err := bundler.InstallBundler(ctx, configuration, logger, versionResolver, calculator, bashCmd, pumainstaller)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(MatchError("failed to obtain ruby version:"))
		})

		it("returns an error on fail to resolve bundler's major version", func() {
			versionResolver.LookupCall.Returns.Err = errors.New("failed to obtain ruby version:")
			ctx = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Layers:     packit.Layers{Path: layersDir},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
			}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewLogger(buffer)
			configuration := bundler.Configuration{
				DefaultBundlerVersion: "#!@ invalid atoi() syntax",
			}

			_, err := bundler.InstallBundler(ctx, configuration, logger, versionResolver, calculator, bashCmd, pumainstaller)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(MatchError(MatchRegexp("invalid syntax")))
		})

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
				calculator,
				bashCmd)
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
					calculator,
					bashCmd)
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
					calculator,
					bashCmd)
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
					calculator,
					bashCmd)
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
						calculator,
						bashCmd)
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
						calculator,
						bashCmd)
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
						calculator,
						bashCmd)
					Expect(err).To(MatchError("failed to calculate checksum"))
				})
			})
		})
	})
}
