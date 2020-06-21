package bundler

import (
	"bytes"
	"fmt"
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

	logger.Process("Installing Bundler version '%s'", bundlerVersion(context, configuration))

	bundlerLayer.Metadata = map[string]interface{}{
		"version": bundlerVersion(context, configuration),
	}

	bundlerLayer.Build = true
	bundlerLayer.Cache = true
	bundlerLayer.Launch = true

	err = InstallPuma(context, configuration, logger)
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

	bundleCleanCmd := strings.Join([]string{
		"bundle",
		"clean",
	}, " ")
	err = RunBashCmd(bundleCleanCmd, context)
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

	var stdOutBytes bytes.Buffer
	cmd.Stdout = &stdOutBytes

	var stdErrBytes bytes.Buffer
	cmd.Stderr = &stdErrBytes

	err := cmd.Run()

	if err != nil {
		logger.Process("Command failed: %s", cmd.String())
		if len(stdErrBytes.String()) > 0 {
			logger.Process("Command stderr:")
			logger.Subprocess(stdErrBytes.String())
		}
		logger.Process("Error status code: %s", err.Error())
		logger.Break()
		return err
	}

	logger.Process("Command succeeded: %s", cmd.String())
	if len(stdOutBytes.String()) > 0 {
		logger.Process("Command output:")
		logger.Subprocess(stdOutBytes.String())
	}

	logger.Break()
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
