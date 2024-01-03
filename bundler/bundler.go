package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Calculator --output fakes/calculator.go
//go:generate faux --interface VersionResolver --output fakes/version_resolver.go
//go:generate faux --interface BashCmd --output fakes/bash_cmd.go
//go:generate faux --interface PumaInstaller --output fakes/puma.go

// VersionResolver defines the interface for looking up and comparing the
// versions of Ruby installed in the environment.
type VersionResolver interface {
	Lookup(workingDir string, bashcmd BashCmd) (version string, err error)
}

// Calculator defines the interface for calculating a checksum of the given set
// of file paths.
type Calculator interface {
	Sum(paths ...string) (string, error)
}

// BashCmd defines the interface for running a bash command.
type BashCmd interface {
	RunBashCmd(command string, WorkingDir string) (string, error)
}

// PumaInst defines the interface for running a bash command.
type PumaInstaller interface {
	InstallPuma(context packit.BuildContext, configuration Configuration, logger scribe.Logger) error
	CreatePumaProcess(context packit.BuildContext, configuration Configuration, logger scribe.Logger) (packit.Process, error)
}

// InstallBundler install bundler in a given RVM environment
//
// To configure the Bundler environment, InstallBundler will copy the local
// Bundler configuration, if any, into the target layer path. The configuration
// file created in the layer will become the defacto configuration file by
// setting `BUNDLE_USER_CONFIG` in the local environment while executing the
// subsequent Bundle CLI commands. The configuration will then be modifed with
// any settings specific to the invocation of Execute.  These configurations
// will override any settings previously applied in the local Bundle
// configuration.
func InstallBundler(context packit.BuildContext, configuration Configuration, logger scribe.Logger, versionResolver VersionResolver, calculator Calculator, bashcmd BashCmd, pumainstaller PumaInstaller) (packit.BuildResult, error) {
	logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
	logger.Process("default Bundler version: %s\n", bundlerVersion(context, configuration))

	clock := chronos.DefaultClock

	var buildMetadata packit.BuildMetadata
	var launchMetadata packit.LaunchMetadata

	bundlerMajorVersion, err := strconv.Atoi(bundlerVersion(context, configuration)[:1])
	if err != nil {
		logger.Process("Failed to determine bundler major version")
		return packit.BuildResult{}, err
	}

	bundlerLayer, err := context.Layers.Get("rvm-bundler")
	if err != nil {
		return packit.BuildResult{}, err
	}

	should, checksum, rubyVersion, err := ShouldRun(bundlerLayer.Metadata, context.WorkingDir, versionResolver, calculator, bashcmd)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rubyMajorVersion, rubyMinorVersion, err := extractRubyVersion(rubyVersion)
	if err != nil {
		return packit.BuildResult{}, err
	}

	localConfigPath := filepath.Join(context.WorkingDir, ".bundle", "config")
	backupConfigPath := filepath.Join(context.WorkingDir, ".bundle", "config.bak")
	globalConfigPath := filepath.Join(bundlerLayer.Path, "config")

	err = os.RemoveAll(globalConfigPath)
	if err != nil {
		return packit.BuildResult{}, err
	}

	if _, err := os.Stat(localConfigPath); err == nil {
		err := os.MkdirAll(bundlerLayer.Path, os.ModePerm)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if _, err := os.Stat(backupConfigPath); err == nil {
			err = fs.Copy(backupConfigPath, localConfigPath)
			if err != nil {
				return packit.BuildResult{}, err
			}
		}

		err = fs.Copy(localConfigPath, globalConfigPath)
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = fs.Copy(localConfigPath, backupConfigPath)
		if err != nil {
			return packit.BuildResult{}, err
		}
	}

	os.Setenv("BUNDLE_USER_CONFIG", globalConfigPath)

	if should {
		timeStartInstall := clock.Now()
		logger.Process("Installing Bundler version '%s'", bundlerVersion(context, configuration))

		rubyGemsVersion := ""
		rubyMajorCheck := "2"
		rubyMinorCheck := []string{"6", "7"}
		if bundlerMajorVersion == 1 {
			rubyGemsVersion = "3.0.8"
		} else if rubyMajorVersion == rubyMajorCheck && slices.Contains(rubyMinorCheck, rubyMinorVersion) {
			// ruby gems 3.4.22 is the latest with Ruby 2.6 and 2.7 support
			rubyGemsVersion = "3.4.22"
		}

		logger.Process("rubygems-update version explicitly set to '%s'", rubyGemsVersion)

		installRubyGemsUpdateSystemCmd := strings.Join([]string{
			"gem",
			"install",
			"-N",
			"rubygems-update",
		}, " ")
		if len(rubyGemsVersion) > 0 {
			installRubyGemsUpdateSystemCmd = strings.Join([]string{
				installRubyGemsUpdateSystemCmd,
				"-v",
				rubyGemsVersion,
			}, " ")
		}
		_, err = bashcmd.RunBashCmd(installRubyGemsUpdateSystemCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		gemUpdateSystemCmd := strings.Join([]string{
			"gem",
			"update",
			"-N",
			"--system",
			rubyGemsVersion,
		}, " ")
		_, err = bashcmd.RunBashCmd(gemUpdateSystemCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		gemCleanupCmd := strings.Join([]string{"gem", "cleanup"}, " ")
		_, err = bashcmd.RunBashCmd(gemCleanupCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = pumainstaller.InstallPuma(context, configuration, logger)
		if err != nil {
			return packit.BuildResult{}, err
		}

		gemInstallBundlerCmd := strings.Join([]string{
			"gem",
			"install",
			"-N",
			"--default",
			"bundler",
		}, " ")
		if bundlerVersion(context, configuration) != "" {
			gemInstallBundlerCmd = strings.Join([]string{
				gemInstallBundlerCmd,
				"-v",
				bundlerVersion(context, configuration),
			}, " ")
		}
		_, err = bashcmd.RunBashCmd(gemInstallBundlerCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = configureBundlerPath(context, bundlerLayer, bundlerMajorVersion, bashcmd)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundleInstallCmd := strings.Join([]string{
			"bundle",
			"install",
		}, " ")
		_, err = bashcmd.RunBashCmd(bundleInstallCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundleCleanCmd := strings.Join([]string{
			"bundle",
			"clean",
		}, " ")
		_, err = bashcmd.RunBashCmd(bundleCleanCmd, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bundlerLayer.Metadata = map[string]interface{}{
			"version":      bundlerVersion(context, configuration),
			"built_at":     clock.Now().Format(time.RFC3339Nano),
			"cache_sha":    checksum,
			"ruby_version": rubyVersion,
		}

		timeDuration := clock.Now().Sub(timeStartInstall)
		logger.Action("RVM Bundler CNB completed in %s", timeDuration.Round(time.Millisecond))
		logger.Break()
	} else {
		logger.Process("Reusing cached layer %s", bundlerLayer.Path)
		logger.Break()

		err = configureBundlerPath(context, bundlerLayer, bundlerMajorVersion, bashcmd)
		if err != nil {
			return packit.BuildResult{}, err
		}
	}

	bundlerLayer.BuildEnv.Default("BUNDLE_USER_CONFIG", filepath.Join(bundlerLayer.Path, "config"))
	bundlerLayer.LaunchEnv.Default("BUNDLE_USER_CONFIG", filepath.Join(bundlerLayer.Path, "config"))

	bundlerLayer.Build, bundlerLayer.Cache, bundlerLayer.Launch = true, true, true

	buildResult := packit.BuildResult{
		Layers: []packit.Layer{bundlerLayer},
		Build:  buildMetadata,
		Launch: launchMetadata,
	}

	pumaProcess, err := pumainstaller.CreatePumaProcess(context, configuration, logger)
	if err == nil && pumaProcess.Type == "web" && pumaProcess.Command != "" {
		buildResult.Launch.Processes = append(buildResult.Launch.Processes, pumaProcess)
	}
	return buildResult, nil
}

func bundlerVersion(context packit.BuildContext, configuration Configuration) string {
	bundlerVersion := configuration.DefaultBundlerVersion
	for _, entry := range context.Plan.Entries {
		if entry.Name == "rvm-bundler" {
			if version, ok := entry.Metadata["rvm_bundler_version"].(string); ok {
				bundlerVersion = version
			}
		}
	}
	return bundlerVersion
}

func configureBundlerPath(context packit.BuildContext, bundlerLayer packit.Layer, bundlerMajorVersion int, bashcmd BashCmd) error {
	cmdComponents := []string{"bundle", "config"}
	if bundlerMajorVersion > 1 {
		cmdComponents = append(cmdComponents, "set")
	}
	cmdComponents = append(cmdComponents, "--local", "path", bundlerLayer.Path)
	_, err := bashcmd.RunBashCmd(strings.Join(cmdComponents, " "), context.WorkingDir)
	if err != nil {
		return err
	}
	return nil
}

// ShouldRun will return true if it is determined that the BundleInstallProcess
// be executed during the build phase.
//
// The criteria for determining that the install process should be executed is
// if the major or minor version of Ruby has changed, or if the contents of the
// Gemfile or Gemfile.lock have changed.
//
// In addition to reporting if the install process should execute, this method
// will return the current version of Ruby and the checksum of the Gemfile and
// Gemfile.lock contents.
func ShouldRun(metadata map[string]interface{}, workingDir string, versionResolver VersionResolver, calculator Calculator, bashcmd BashCmd) (bool, string, string, error) {

	rubyVersion, err := versionResolver.Lookup(workingDir, bashcmd)
	if err != nil {
		return false, "", "", err
	}

	cachedRubyVersion, ok := metadata["ruby_version"].(string)
	rubyVersionMatch := true

	if ok {
		if cachedRubyVersion != rubyVersion {
			rubyVersionMatch = false
		}
	}

	var sum string
	_, err = os.Stat(filepath.Join(workingDir, "Gemfile.lock"))
	if err != nil {
		if !os.IsNotExist(err) {
			return false, "", "", err
		}
	} else {
		sum, err = calculator.Sum(filepath.Join(workingDir, "Gemfile"), filepath.Join(workingDir, "Gemfile.lock"))
		if err != nil {
			return false, "", "", err
		}
	}

	cachedSHA, ok := metadata["cache_sha"].(string)
	cacheMatch := ok && cachedSHA == sum
	shouldRun := !cacheMatch || !rubyVersionMatch

	return shouldRun, sum, rubyVersion, nil
}

// Extracts major ruby version from a substring like "ruby-1.2.3" or "ruby-head" from "^ruby" and "^jruby" strings
// Returns major + minor versions OR "head" + "" and an error.
func extractRubyVersion(version string) (string, string, error) {
	re := regexp.MustCompile(`ruby-(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)

	if len(matches) < 2 {
		if strings.Contains(version, "ruby-head") {
			return "head", "", nil
		}
		return "", "", fmt.Errorf("unable to extract Ruby version from: %s", version)
	}

	return matches[1], matches[2], nil
}
