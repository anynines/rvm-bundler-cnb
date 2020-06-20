package bundler

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/paketo-buildpacks/packit"
)

// InstallBundler install bundler in a given RVM environment
func InstallBundler(context packit.BuildContext, configuration Configuration, logger rvm.LogEmitter) (packit.BuildResult, error) {
	logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
	logger.Process("default Bundler version: %s\n", bundlerVersion(context, configuration))

	bundlerLayer, err := context.Layers.Get("rvm-bundler", packit.LaunchLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	// Assumption: the contents of the Gemfile and the Gemfile.lock files should
	// be computed to the same SHA256 hashsum between two runs of this CNB, but
	// the Gemfile.lock may not yet exist. If the Gemfile.lock doesn't exist,
	// then we definitely cannot re-use this CNB's layer because "bundle install"
	// was not run yet
	lockCacheEqual := false
	if _, err = os.Stat(filepath.Join(context.WorkingDir, "Gemfile.lock")); err == nil &&
		bundlerLayer.Metadata["gemfile_lock_sha256"] != nil {
		gemfileLockContents, err := ioutil.ReadFile(filepath.Join(context.WorkingDir, "Gemfile.lock"))
		if err != nil {
			return packit.BuildResult{}, err
		}
		gemfileLockSHA256 := sha256.Sum256(gemfileLockContents)
		lockCacheEqual = bundlerLayer.Metadata["gemfile_lock_sha256"].(string) == fmt.Sprintf("%x", gemfileLockSHA256)
	}

	gemfileContents, err := ioutil.ReadFile(filepath.Join(context.WorkingDir, "Gemfile"))
	gemfileSHA256 := sha256.Sum256(gemfileContents)

	if lockCacheEqual && bundlerLayer.Metadata["gemfile_sha256"] != nil &&
		bundlerLayer.Metadata["gemfile_sha256"].(string) == fmt.Sprintf("%x", gemfileSHA256) {
		logger.Process("Reusing cached layer %s", bundlerLayer.Path)

		err = configureBundlerPath(context, bundlerLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = InstallPuma(context, configuration, logger)
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = createProfileFile(bundlerLayer, context, logger)
		if err != nil {
			return packit.BuildResult{}, err
		}

		buildResult := packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				bundlerLayer,
			},
		}

		pumaProcess, err := CreatePumaProcess(context, configuration, logger)
		if err == nil && pumaProcess.Type == "web" && pumaProcess.Command != "" {
			buildResult.Processes = append(buildResult.Processes, pumaProcess)
		}

		return buildResult, nil
	}

	logger.Process("Installing Bundler version '%s'", bundlerVersion(context, configuration))

	if err = bundlerLayer.Reset(); err != nil {
		logger.Process("Resetting Bundler layer failed")
		return packit.BuildResult{}, err
	}

	bundlerLayer.Metadata = map[string]interface{}{
		"version":        bundlerVersion(context, configuration),
		"gemfile_sha256": fmt.Sprintf("%x", gemfileSHA256),
	}

	bundlerLayer.Build = true
	bundlerLayer.Cache = true
	bundlerLayer.Launch = true

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
	err = RunBashCmd(gemInstallBundlerCmd, context)
	if err != nil {
		return packit.BuildResult{}, err
	}

	err = configureBundlerPath(context, bundlerLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	bundleInstallCmd := strings.Join([]string{
		"bundle",
		"install",
	}, " ")
	err = RunBashCmd(bundleInstallCmd, context)
	if err != nil {
		return packit.BuildResult{}, err
	}

	err = createProfileFile(bundlerLayer, context, logger)
	if err != nil {
		return packit.BuildResult{}, err
	}

	gemfileLockContentsInstalled, err := ioutil.ReadFile(filepath.Join(context.WorkingDir, "Gemfile.lock"))
	if err != nil {
		return packit.BuildResult{}, err
	}
	bundlerLayer.Metadata["gemfile_lock_sha256"] = fmt.Sprintf("%x", sha256.Sum256(gemfileLockContentsInstalled))

	err = InstallPuma(context, configuration, logger)
	if err != nil {
		return packit.BuildResult{}, err
	}

	buildResult := packit.BuildResult{
		Plan: context.Plan,
		Layers: []packit.Layer{
			bundlerLayer,
		},
	}

	pumaProcess, err := CreatePumaProcess(context, configuration, logger)
	if err == nil && pumaProcess.Type == "web" && pumaProcess.Command != "" {
		buildResult.Processes = append(buildResult.Processes, pumaProcess)
	}
	return buildResult, nil
}

// RunBashCmd executes a command in an interactive BASH shell
func RunBashCmd(command string, context packit.BuildContext) error {
	logger := rvm.NewLogEmitter(os.Stdout)

	cmd := exec.Command("bash")
	cmd.Dir = context.WorkingDir
	cmd.Args = append(
		cmd.Args,
		"--login",
		"-c",
		strings.Join(
			[]string{
				"source",
				filepath.Join(os.ExpandEnv("$rvm_path"), "profile.d", "rvm"),
				"&&",
				command,
			},
			" ",
		),
	)
	cmd.Env = os.Environ()

	logger.Process("Executing: %s", strings.Join(cmd.Args, " "))
	logger.Break()

	var stdOutBytes bytes.Buffer
	cmd.Stdout = &stdOutBytes

	var stdErrBytes bytes.Buffer
	cmd.Stderr = &stdErrBytes

	err := cmd.Run()

	if err != nil {
		logger.Subprocess("Command failed: %s", cmd.String())
		logger.Subprocess("Command stderr:\n\n%s", stdErrBytes.String())
		logger.Subprocess("Error status code: %s", err.Error())
		return err
	}

	logger.Subprocess("Command succeeded: %s", cmd.String())
	logger.Subprocess("Command output:\n\n%s", stdOutBytes.String())

	return nil
}

func bundlerVersion(context packit.BuildContext, configuration Configuration) string {
	bundlerVersion := configuration.DefaultBundlerVersion
	for _, entry := range context.Plan.Entries {
		if entry.Name == "rvm-bundler" {
			bundlerVersion = fmt.Sprintf("%v", entry.Version)
		}
	}
	return bundlerVersion
}

func configureBundlerPath(context packit.BuildContext, bundlerLayer packit.Layer) error {
	bundleConfigCmd := strings.Join([]string{
		"bundle",
		"config",
		"set",
		"--local",
		"path",
		bundlerLayer.Path,
	}, " ")
	err := RunBashCmd(bundleConfigCmd, context)
	if err != nil {
		return err
	}
	return nil
}

// createProfileFile creates a file called ".profile" in the applicaton
// directory in order to add Bundler's Gem path to the $GEM_PATH variable
// The .profile file is loaded during launch according to:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch
func createProfileFile(bundlerLayer packit.Layer, context packit.BuildContext, logger rvm.LogEmitter) error {
	// the following is a bit "hacky" but I have not yet found a better way to
	// determine the path where bundler stores its gems because that path name
	// depends on the Ruby platform. I assume the following will work because
	// only subdirectory exists in the "ruby" platform directory
	gemPath := filepath.Join(bundlerLayer.Path, "ruby")
	pathMatches, err := filepath.Glob(gemPath + string(os.PathSeparator) + "*")
	if err != nil {
		return err
	}
	if len(pathMatches) > 0 {
		profileFilePath := filepath.Join(context.WorkingDir, ".profile")
		logger.Process("Writing .profile to path: '%s'", profileFilePath)

		profileFile, err := os.OpenFile(profileFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer profileFile.Close()

		exportGemPathString := fmt.Sprintf("export GEM_PATH=\"$GEM_PATH:%s\"\n", pathMatches[0])
		_, err = profileFile.Write([]byte(exportGemPathString))
		if err != nil {
			return err
		}
	}
	return nil
}
