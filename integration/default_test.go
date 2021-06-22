package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
			format.MaxLength = 0
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("installs with the defaults", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.RVM.Online,
					settings.Buildpacks.Bundler.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())

			// Expected the most recent version of `bundler`, because `gems` are updated to the most recent version
			Eventually(container).Should(Serve(ContainSubstring("/layers/com.anynines.buildpacks.rvm/rvm/gems/")).OnPort(8080))
			Eventually(container).Should(Serve(MatchRegexp(`Bundler version 2\.\d+\.\d+`)).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`  default Bundler version: 2\.1\.4`),
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Installing RubyGems 3\.\d+\.\d+`),
				MatchRegexp(`      Successfully built RubyGem`),
				MatchRegexp(`      Name: bundler`),
				MatchRegexp(`      Version: 2\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Using config/puma.rb supplied by application",
				MatchRegexp(`  Adding Puma version: '\d+\.\d+\.\d+' to Gemfile`),
			))
			Expect(logs).To(ContainLines(
				"  Returning process type 'web' with command 'bundle exec puma'",
			))
		})
	})
}
