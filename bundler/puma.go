package bundler

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
	"github.com/paketo-buildpacks/packit"
)

// InstallPuma install the Puma gem and creates a workingDir/config/puma.rb if
// the file doesn't exist already
func InstallPuma(context packit.BuildContext, configuration Configuration, logger rvm.LogEmitter) error {
	if configuration.InstallPuma == false {
		return nil
	}

	logger.Process("Installing Puma version: '%s'", configuration.Puma.Version)

	gemInstallPumaCmd := strings.Join([]string{
		"gem",
		"install",
		"-N",
		"puma",
		"-v",
		configuration.Puma.Version,
	}, " ")
	err := RunBashCmd(gemInstallPumaCmd, context)
	if err != nil {
		return err
	}

	configPumaRbPath := filepath.Join(context.WorkingDir, "config", "puma.rb")
	_, err = os.Stat(configPumaRbPath)
	if os.IsNotExist(err) {
		logger.Process("Creating configuration file for Puma at: '%s'", configPumaRbPath)

		configPumaRb, err := os.OpenFile(configPumaRbPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer configPumaRb.Close()

		configPumaRb.Write([]byte(fmt.Sprintf("bind '%s'\n", configuration.Puma.Bind)))
		configPumaRb.Write([]byte(fmt.Sprintf("workers %d\n", configuration.Puma.Workers)))
		configPumaRb.Write([]byte(fmt.Sprintf("threads %d, %d\n", configuration.Puma.Threads, configuration.Puma.Threads)))
		configPumaRb.Write([]byte(fmt.Sprintf("log_requests true\n")))
		if configuration.Puma.Preload {
			configPumaRb.Write([]byte(fmt.Sprintf("preload_app!\n")))
		}

		return nil
	}

	logger.Process("Using config/puma.rb supplied by application")
	return nil
}

// CreatePumaProcess creates a packit.Process if this buildpack is configured
// to do so. If there is a Procfile in the application's directory and it
// it contains a process of type "web:", then no packit.Process will be returned
func CreatePumaProcess(context packit.BuildContext, configuration Configuration, logger rvm.LogEmitter) (packit.Process, error) {
	installPumaCommand := true
	procfile, err := os.Open(filepath.Join(context.WorkingDir, "Procfile"))
	if err == nil {
		defer procfile.Close()

		scanner := bufio.NewScanner(procfile)
		for scanner.Scan() {
			matched, err := regexp.MatchString(`^web:.*$`, strings.TrimSpace(scanner.Text()))
			if err == nil && matched {
				installPumaCommand = false
				logger.Process("Do not return a process because a Procfile with process type 'web' already exists")
			}
		}
	}

	if installPumaCommand {
		logger.Process("Returning process type 'web' with command 'puma'")

		return packit.Process{
			Type:    "web",
			Command: "puma",
		}, nil
	}

	return packit.Process{}, nil
}
