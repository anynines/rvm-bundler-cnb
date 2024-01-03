package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
		format.MaxLength = 0
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.RVM.Online,
					settings.Buildpacks.Bundler.Online,
					settings.Buildpacks.BuildPlan.Online,
				)

			firstImage, _, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(3))

			Expect(firstImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

			firstContainer, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(3))

			Expect(secondImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  default Bundler version: 2.3.14",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf("  Reusing cached layer /layers/%s/rvm-bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))

			secondContainer, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			Expect(secondImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		context("when the Gemfile.lock changes", func() {
			it("rebuilds the layer", func() {
				var (
					err             error
					logs            fmt.Stringer
					firstImage      occam.Image
					secondImage     occam.Image
					firstContainer  occam.Container
					secondContainer occam.Container
				)

				source, err = occam.Source(filepath.Join("testdata", "default_app"))
				Expect(err).NotTo(HaveOccurred())

				build := pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.RVM.Online,
						settings.Buildpacks.Bundler.Online,
						settings.Buildpacks.BuildPlan.Online,
					)

				firstImage, _, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				imageIDs[firstImage.ID] = struct{}{}

				Expect(firstImage.Buildpacks).To(HaveLen(3))

				Expect(firstImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
				Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

				firstContainer, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(firstImage.ID)
				Expect(err).NotTo(HaveOccurred())

				containerIDs[firstContainer.ID] = struct{}{}

				Eventually(firstContainer).Should(BeAvailable())

				contents, err := os.ReadFile(filepath.Join(source, "Gemfile.lock"))
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(source, "Gemfile.lock"),
					[]byte(string(contents)+"\n"), 0600)
				Expect(err).NotTo(HaveOccurred())

				// Second pack build
				secondImage, logs, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				imageIDs[secondImage.ID] = struct{}{}

				Expect(secondImage.Buildpacks).To(HaveLen(3))

				Expect(secondImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
				Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

				Expect(logs.String()).NotTo(ContainSubstring(fmt.Sprintf("  Reusing cached layer /layers/%s/rvm-bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))))

				secondContainer, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(firstImage.ID)
				Expect(err).NotTo(HaveOccurred())

				containerIDs[secondContainer.ID] = struct{}{}

				Eventually(secondContainer).Should(BeAvailable())

				Expect(secondImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]))
			})
		})

		context("when the Gemfile changes", func() {
			it("rebuilds the layer", func() {
				var (
					err         error
					logs        fmt.Stringer
					firstImage  occam.Image
					secondImage occam.Image

					firstContainer  occam.Container
					secondContainer occam.Container
				)

				source, err = occam.Source(filepath.Join("testdata", "default_app"))
				Expect(err).NotTo(HaveOccurred())

				build := pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.RVM.Online,
						settings.Buildpacks.Bundler.Online,
						settings.Buildpacks.BuildPlan.Online,
					)

				firstImage, _, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				imageIDs[firstImage.ID] = struct{}{}

				Expect(firstImage.Buildpacks).To(HaveLen(3))

				Expect(firstImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
				Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

				firstContainer, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(firstImage.ID)
				Expect(err).NotTo(HaveOccurred())

				containerIDs[firstContainer.ID] = struct{}{}

				Eventually(firstContainer).Should(BeAvailable())

				contents, err := os.ReadFile(filepath.Join(source, "Gemfile"))
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(source, "Gemfile"),
					[]byte(strings.ReplaceAll(string(contents),
						`gem 'sinatra', '~>2.1.0'`,
						`gem 'sinatra', '~>2.0.8'`,
					)), 0600)
				Expect(err).NotTo(HaveOccurred())

				// Second pack build
				secondImage, logs, err = build.Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				imageIDs[secondImage.ID] = struct{}{}

				Expect(secondImage.Buildpacks).To(HaveLen(3))

				Expect(secondImage.Buildpacks[1].Key).To(Equal(settings.Buildpack.ID))
				Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("rvm-bundler"))

				Expect(logs.String()).To(MatchRegexp(`  default Bundler version: \d+\.\d+\.\d+`))
				Expect(logs.String()).To(MatchRegexp(`  Installing Bundler version \'\d+\.\d+\.\d+\'`))
				Expect(logs.String()).NotTo(ContainSubstring(fmt.Sprintf("  Reusing cached layer /layers/%s/rvm-bundler", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))))

				secondContainer, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(firstImage.ID)
				Expect(err).NotTo(HaveOccurred())

				containerIDs[secondContainer.ID] = struct{}{}

				Eventually(secondContainer).Should(BeAvailable())

				Expect(secondImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[1].Layers["rvm-bundler"].Metadata["built_at"]))
			})
		})
	})
}
