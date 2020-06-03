package rvm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

// Env represents an RVM environment
type Env struct {
	Context       packit.BuildContext
	Logger        LogEmitter
	Configuration Configuration
	Environment   EnvironmentConfiguration
}

// BuildRvm builds the RVM environment
func (r Env) BuildRvm() (packit.BuildResult, error) {
	r.Logger.Title("%s %s", r.Context.BuildpackInfo.Name, r.Context.BuildpackInfo.Version)

	r.Logger.Process("Using RVM URI: %s\n", r.Configuration.URI)
	r.Logger.Process("default RVM version: %s\n", r.Configuration.DefaultRVMVersion)
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
	r.Logger.Subprocess("Environment variables:\n%s", strings.Join(cmd.Env, "\n"))
	r.Logger.Break()

	var stdOutBytes bytes.Buffer
	cmd.Stdout = &stdOutBytes

	var stdErrBytes bytes.Buffer
	cmd.Stderr = &stdErrBytes

	err := cmd.Run()

	if err != nil {
		r.Logger.Subprocess("Command failed: %s", cmd.String())
		r.Logger.Subprocess("Command stderr: %s", stdErrBytes.String())
		r.Logger.Subprocess("Error status code: %s", err.Error())
		return err
	}

	r.Logger.Subprocess("Command succeeded: %s", cmd.String())
	r.Logger.Subprocess("Command output: %s", stdOutBytes.String())

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
		if entry.Name == "rvm" {
			rubyVersion = fmt.Sprintf("%v", entry.Metadata["ruby_version"])
		}
	}
	return rubyVersion
}

func (r Env) installRVM() (packit.BuildResult, error) {
	rvmLayer, err := r.Context.Layers.Get("rvm", packit.LaunchLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	if rvmLayer.Metadata["rvm_version"] != nil &&
		rvmLayer.Metadata["rvm_version"].(string) == r.Configuration.DefaultRVMVersion {
		r.Logger.Process("Reusing cached layer %s", rvmLayer.Path)
		return packit.BuildResult{
			Plan: r.Context.Plan,
			Layers: []packit.Layer{
				rvmLayer,
			},
		}, nil
	}

	r.Logger.Process("Installing RVM version '%s' from URI '%s'", r.Configuration.DefaultRVMVersion, r.Configuration.URI)

	if err = rvmLayer.Reset(); err != nil {
		r.Logger.Process("Resetting RVM layer failed")
		return packit.BuildResult{}, err
	}

	rvmLayer.Metadata = map[string]interface{}{
		"rvm_version":  r.Configuration.DefaultRVMVersion,
		"ruby_version": r.rubyVersion(),
	}

	rvmLayer.Build = true
	rvmLayer.Cache = true
	rvmLayer.Launch = true

	err = r.Environment.Configure(rvmLayer.SharedEnv, rvmLayer.Path)
	if err != nil {
		return packit.BuildResult{}, err
	}

	shellCmd := strings.Join([]string{
		"curl",
		"-vsSL",
		r.Configuration.URI,
		"| bash -s -- --version",
		r.Configuration.DefaultRVMVersion,
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

	gemUpdateSystemCmd := strings.Join([]string{
		"gem",
		"update",
		"-N",
		"--system",
	}, " ")
	err = r.RunRvmCmd(gemUpdateSystemCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	gemCleanupCmd := strings.Join([]string{"gem", "cleanup"}, " ")
	err = r.RunRvmCmd(gemCleanupCmd, &rvmLayer)
	if err != nil {
		return packit.BuildResult{}, err
	}

	rvmCleanupCmd := strings.Join([]string{"rvm", "cleanup", "all"}, " ")
	err = r.RunRvmCmd(rvmCleanupCmd, &rvmLayer)
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
