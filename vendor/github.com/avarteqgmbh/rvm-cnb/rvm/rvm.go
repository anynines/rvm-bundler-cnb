package rvm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

// Env represents an RVM environment
type Env struct {
	BuildPackYML  BuildPackYML
	Context       packit.BuildContext
	Logger        LogEmitter
	Configuration Configuration
	Environment   EnvironmentConfiguration
}

// BuildRvm builds the RVM environment
func (r Env) BuildRvm() (packit.BuildResult, error) {
	r.Logger.Title("%s %s", r.Context.BuildpackInfo.Name, r.Context.BuildpackInfo.Version)

	r.Logger.Process("Using RVM URI: %s\n", r.Configuration.URI)
	r.Logger.Process("RVM version: %s\n", r.rvmVersion())
	r.Logger.Process("build plan Ruby version: %s\n", r.rubyVersion())

	buildResult, err := r.installRVM()
	if err != nil {
		return packit.BuildResult{}, err
	}
	return buildResult, nil
}

// RunBashCmd executes a command using BASH
func (r Env) RunBashCmd(command string, rvmLayer *packit.Layer) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(
		os.Environ(),
		DefaultVariables(rvmLayer)...,
	)

	r.Logger.Process("Executing: %s", strings.Join(cmd.Args, " "))

	stdoutPipe, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderrBuf)

	if err := cmd.Start(); err != nil {
		r.Logger.Process("Failed to start command: %s", cmd.String())
		r.Logger.Break()
		return err
	}

	stdoutReader := bufio.NewReader(stdoutPipe)
	stdoutLine, err := stdoutReader.ReadString('\n')
	for err == nil {
		r.Logger.Subprocess(stdoutLine)
		stdoutLine, err = stdoutReader.ReadString('\n')
	}
	err = cmd.Wait()

	if err != nil {
		r.Logger.Process("Command failed: %s", cmd.String())
		r.Logger.Process("Error status code: %s", err.Error())
		if len(stderrBuf.String()) > 0 {
			r.Logger.Process("Command output on stderr:")
			r.Logger.Subprocess(stderrBuf.String())
		}
		r.Logger.Break()
		return err
	}

	r.Logger.Break()

	return nil
}

// RunRvmCmd executes a command in an RVM environment
func (r Env) RunRvmCmd(command string, rvmLayer *packit.Layer) error {
	profileDScript := filepath.Join(rvmLayer.Path, "profile.d", "rvm")
	fullRvmCommand := strings.Join([]string{
		"source",
		profileDScript,
		"&&",
		command,
	}, " ")

	return r.RunBashCmd(fullRvmCommand, rvmLayer)
}

func (r Env) rubyVersion() string {
	rubyVersion := r.Configuration.DefaultRubyVersion
	for _, entry := range r.Context.Plan.Entries {
		if entry.Name == "rvm" && entry.Metadata["ruby_version"] != nil {
			rubyVersion = fmt.Sprintf("%v", entry.Metadata["ruby_version"])
		}
	}
	return rubyVersion
}

func (r Env) rvmVersion() string {
	rvmVersion := r.Configuration.DefaultRVMVersion
	for _, entry := range r.Context.Plan.Entries {
		if entry.Name == "rvm" && entry.Metadata["rvm_version"] != nil {
			rvmVersion = fmt.Sprintf("%v", entry.Metadata["rvm_version"])
		}
	}
	cmp := r.BuildPackYML != BuildPackYML{}
	if cmp && len(r.BuildPackYML.RvmVersion) > 0 {
		rvmVersion = r.BuildPackYML.RvmVersion
	}
	return rvmVersion
}

func (r Env) installRVM() (packit.BuildResult, error) {
	rvmLayer, err := r.Context.Layers.Get("rvm")
	if err != nil {
		return packit.BuildResult{}, err
	}

	if rvmLayer.Metadata["rvm_version"] != nil &&
		rvmLayer.Metadata["rvm_version"].(string) == r.rvmVersion() &&
		rvmLayer.Metadata["ruby_version"] != nil &&
		rvmLayer.Metadata["ruby_version"].(string) == r.rubyVersion() {
		r.Logger.Process("Reusing cached layer %s", rvmLayer.Path)
		return packit.BuildResult{
			Plan: r.Context.Plan,
			Layers: []packit.Layer{
				rvmLayer,
			},
		}, nil
	}

	r.Logger.Process("Installing RVM version '%s' from URI '%s'", r.rvmVersion(), r.Configuration.URI)

	if rvmLayer, err = rvmLayer.Reset(); err != nil {
		r.Logger.Process("Resetting RVM layer failed")
		return packit.BuildResult{}, err
	}

	rvmLayer.Metadata = map[string]interface{}{
		"rvm_version":  r.rvmVersion(),
		"ruby_version": r.rubyVersion(),
	}

	rvmLayer.Build = true
	rvmLayer.Cache = true
	rvmLayer.Launch = true

	err = r.Environment.Configure(rvmLayer.SharedEnv, rvmLayer.Path)
	if err != nil {
		return packit.BuildResult{}, err
	}

	// The following commands import GPP keys:
	// curl -sSL https://rvm.io/mpapis.asc | gpg --import -
	// curl -sSL https://rvm.io/pkuczynski.asc | gpg --import -
	// this is necessary because installing RVM on a Ubuntu 16.04 container which
	// does not have these keys imported yet fails otherwise
	// see: https://rvm.io/rvm/security
	// commented for now and kept for future usage
	gpgBinaryInstalledOutput, _ := exec.Command("which", "gpg").Output()

	if len(gpgBinaryInstalledOutput) > 0 {
		importGPGKey1Cmd := strings.Join([]string{
			"curl",
			"-sSL",
			"https://rvm.io/mpapis.asc",
			"|",
			"gpg",
			"--import",
			"-",
		}, " ")
		err = r.RunBashCmd(importGPGKey1Cmd, &rvmLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		importGPGKey2Cmd := strings.Join([]string{
			"curl",
			"-sSL",
			"https://rvm.io/pkuczynski.asc",
			"|",
			"gpg",
			"--import",
			"-",
		}, " ")
		err = r.RunBashCmd(importGPGKey2Cmd, &rvmLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}
	}

	shellCmd := strings.Join([]string{
		"curl",
		"-vsSL",
		r.Configuration.URI,
		"| bash -s -- --version",
		r.rvmVersion(),
	}, " ")
	err = r.RunBashCmd(shellCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	autolibsCmd := strings.Join([]string{
		filepath.Join(rvmLayer.Path, "bin", "rvm"),
		"autolibs",
		"0",
	}, " ")
	err = r.RunRvmCmd(autolibsCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rubyInstallCmd := strings.Join([]string{
		filepath.Join(rvmLayer.Path, "bin", "rvm"),
		"install",
		r.rubyVersion(),
	}, " ")
	err = r.RunRvmCmd(rubyInstallCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rvmCleanupCmd := strings.Join([]string{"rvm", "cleanup", "all"}, " ")
	err = r.RunRvmCmd(rvmCleanupCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rvmSetDefaultRubyCmd := strings.Join([]string{"rvm", "alias", "create", "default", r.rubyVersion()}, " ")
	err = r.RunRvmCmd(rvmSetDefaultRubyCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	return packit.BuildResult{
		Plan: r.Context.Plan,
		Layers: []packit.Layer{
			rvmLayer,
		},
	}, nil
}
