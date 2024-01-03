package bundler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	bundler "github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPuma(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir    string
		cnbDir        string
		layersDir     string
		buffer        *bytes.Buffer
		emptyBuffer   []byte
		ctx           packit.BuildContext
		logger        scribe.Logger
		puma          bundler.PumaGemInstaller
		configuration bundler.Configuration
	)

	it.Before(func() {
		var err error

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		emptyBuffer = []byte(``)
	})

	it.After(func() {
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
	})

	context("InstallPuma", func() {
		it.Before(func() {
			buffer = bytes.NewBuffer(nil)
			logger = scribe.NewLogger(buffer)
			configuration = bundler.Configuration{
				DefaultBundlerVersion: "2.1.4",
				InstallPuma:           true,
				Puma: bundler.Puma{
					Version: "2.0.0",
					Bind:    "tcp://0.0.0.0:8080",
					Workers: "2",
					Threads: "2",
					Preload: true,
				},
			}
		})

		it("creates config/puma.rb when it doesn't exist and Gemfile.lock doesn't contain puma version", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(workingDir, "config"), 0700)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile"), emptyBuffer, 0644)
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

			puma = bundler.NewPumaInstaller()

			err = puma.InstallPuma(ctx, configuration, logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("creates config/puma.rb when it doesn't exist and Gemfile.lock contains puma version", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(workingDir, "config"), 0700)
			Expect(err).NotTo(HaveOccurred())

			filledBuffer := []byte("puma (2.0.0)")
			err = os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), filledBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile"), emptyBuffer, 0644)
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

			puma = bundler.NewPumaInstaller()

			err = puma.InstallPuma(ctx, configuration, logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

	})

	context("PumaDisabled", func() {
		it.Before(func() {

			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(workingDir, "config"), 0700)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile"), emptyBuffer, 0644)
			Expect(err).NotTo(HaveOccurred())

			buffer = bytes.NewBuffer(nil)
			logger = scribe.NewLogger(buffer)

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
			configuration = bundler.Configuration{
				DefaultBundlerVersion: "2.1.4",
				InstallPuma:           false,
				Puma: bundler.Puma{
					Version: "2.0.0",
					Bind:    "tcp://0.0.0.0:8080",
					Workers: "2",
					Threads: "2",
					Preload: true,
				},
			}
		})

		it("exits when Puma is disabled", func() {

			puma = bundler.NewPumaInstaller()

			err := puma.InstallPuma(ctx, configuration, logger)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})
	})

	context("Fails on Puma installation", func() {
		it.Before(func() {

			buffer = bytes.NewBuffer(nil)
			logger = scribe.NewLogger(buffer)

			configuration = bundler.Configuration{
				DefaultBundlerVersion: "2.1.4",
				InstallPuma:           true,
				Puma: bundler.Puma{
					Version: "2.0.0",
					Bind:    "tcp://0.0.0.0:8080",
					Workers: "2",
					Threads: "2",
					Preload: true,
				},
			}
		})

		it("fails when config/Gemfile.lock doesn't exist", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(workingDir, "config"), 0700)
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

			puma = bundler.NewPumaInstaller()

			err = puma.InstallPuma(ctx, configuration, logger)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(MatchError(MatchRegexp("no such file or directory")))
		})

		it("fails when config/Gemfile can't be opened", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(workingDir, "config"), 0700)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Gemfile.lock"), emptyBuffer, 0644)
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

			puma = bundler.NewPumaInstaller()

			err = puma.InstallPuma(ctx, configuration, logger)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(MatchError(MatchRegexp("no such file or directory")))
		})
	})

	context("CreatePumaProcess", func() {
		it.Before(func() {

			buffer = bytes.NewBuffer(nil)
			logger = scribe.NewLogger(buffer)

			configuration = bundler.Configuration{
				DefaultBundlerVersion: "2.1.4",
				InstallPuma:           false,
				Puma: bundler.Puma{
					Version: "2.0.0",
					Bind:    "tcp://0.0.0.0:8080",
					Workers: "2",
					Threads: "2",
					Preload: true,
				},
			}
		})

		it("procfile exists but empty", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(workingDir, "Procfile"), emptyBuffer, 0644)
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

			puma = bundler.NewPumaInstaller()

			_, err = puma.CreatePumaProcess(ctx, configuration, logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("procfile exists and contains `web` process", func() {
			workingDir, err := os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			filledBuffer := []byte("web: bundle exec puma -C config/puma.rb\n")
			err = os.WriteFile(filepath.Join(workingDir, "Procfile"), filledBuffer, 0644)
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

			puma = bundler.NewPumaInstaller()

			_, err = puma.CreatePumaProcess(ctx, configuration, logger)
			Expect(err).NotTo(HaveOccurred())

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})
	})

}
