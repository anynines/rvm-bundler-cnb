package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpacks struct {
		Bundler struct {
			Online  string
			Offline string
		}
		BuildPlan struct {
			Online string
		}
		RVM struct {
			Online  string
			Offline string
		}
	}

	Buildpack struct {
		ID   string
		Name string
	}

	Config struct {
		BuildPlan string `json:"build-plan"`
		RVM       string `json:"rvm"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.Bundler.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Bundler.Offline, err = buildpackStore.Get.
		WithVersion("1.2.3").
		WithOfflineDependencies().
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.RVM.Online, err = buildpackStore.Get.
		Execute(settings.Config.RVM)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.RVM.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(settings.Config.RVM)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("gemfile.lock", testGemfileLock)
	suite("Ruby2", testRuby2)
	suite("Default", testDefault)
	suite("Layer Reuse", testLayerReuse)
	suite.Run(t)
}
