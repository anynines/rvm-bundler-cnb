package bundler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	bundlerMajorVersion, err := strconv.Atoi(bundlerVersion(context, configuration)[:1])
	if err != nil {
		logger.Process("Failed to determine bundler major version")
		return packit.BuildResult{}, err
	}

	rubyGemsVersion := ""
	if bundlerMajorVersion == 1 {
		rubyGemsVersion = "3.0.8"
	}

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
	err = RunBashCmd(installRubyGemsUpdateSystemCmd, context)
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
	err = RunBashCmd(gemUpdateSystemCmd, context)
	if err != nil {
		return packit.BuildResult{}, err
	}

	gemCleanupCmd := strings.Join([]string{"gem", "cleanup"}, " ")
	err = RunBashCmd(gemCleanupCmd, context)
	if err != nil {
		return packit.BuildResult{}, err
	}

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

	err = configureBundlerPath(context, bundlerLayer, bundlerMajorVersion)
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

	stdoutPipe, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderrBuf)

	if err := cmd.Start(); err != nil {
		logger.Process("Failed to start command: %s", cmd.String())
		logger.Break()
		return err
	}

	stdoutReader := bufio.NewReader(stdoutPipe)
	stdoutLine, err := stdoutReader.ReadString('\n')
	for err == nil {
		logger.Subprocess(stdoutLine)
		stdoutLine, err = stdoutReader.ReadString('\n')
	}
	err = cmd.Wait()

	if err != nil {
		logger.Process("Command failed: %s", cmd.String())
		logger.Process("Error status code: %s", err.Error())
		if len(stderrBuf.String()) > 0 {
			logger.Process("Command output on stderr:")
			logger.Subprocess(stderrBuf.String())
		}
		return err
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

func configureBundlerPath(context packit.BuildContext, bundlerLayer packit.Layer, bundlerMajorVersion int) error {
	cmdComponents := []string{"bundle", "config"}
	if bundlerMajorVersion > 1 {
		cmdComponents = append(cmdComponents, "set")
	}
	cmdComponents = append(cmdComponents, "--local", "path", bundlerLayer.Path)
	err := RunBashCmd(strings.Join(cmdComponents, " "), context)
	if err != nil {
		return err
	}
	return nil
}
