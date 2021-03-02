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
	if !configuration.InstallPuma {
		return nil
	}

	configPumaRbPath := filepath.Join(context.WorkingDir, "config", "puma.rb")
	_, err := os.Stat(configPumaRbPath)
	if os.IsNotExist(err) {
		logger.Process("Creating configuration file for Puma at: '%s'", configPumaRbPath)

		configPumaRb, err := os.OpenFile(configPumaRbPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer configPumaRb.Close()

		pumacfg := ""
		pumacfg += fmt.Sprintf("bind '%s'\n", configuration.Puma.Bind)
		pumacfg += fmt.Sprintf("workers %s\n", configuration.Puma.Workers)
		pumacfg += fmt.Sprintf("threads %s, %s\n", configuration.Puma.Threads, configuration.Puma.Threads)
		pumacfg += "log_requests true\n"
		if configuration.Puma.Preload {
			pumacfg += "preload_app!\n"
		}
		pumacfg += "activate_control_app 'unix:///tmp/pumactl.sock', { no_token: true }\n"
		_, err = configPumaRb.WriteString(pumacfg)
		if err != nil {
			return err
		}
	}

	logger.Process("Using config/puma.rb supplied by application")

	GemfileLock, err := os.Open(filepath.Join(context.WorkingDir, "Gemfile.lock"))
	if err != nil {
		return err
	}
	defer GemfileLock.Close()

	scanner := bufio.NewScanner(GemfileLock)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "puma (") {
			logger.Process("Puma is present in Gemfile.lock")
			logger.Break()
			return nil
		}
	}

	logger.Process("Adding Puma version: '%s' to Gemfile", configuration.Puma.Version)

	Gemfile, err := os.OpenFile(filepath.Join(context.WorkingDir, "Gemfile"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer Gemfile.Close()

	_, err = Gemfile.WriteString(fmt.Sprintf("\ngem \"puma\", \"%s\"\n", configuration.Puma.Version))
	if err != nil {
		return err
	}

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
		r, _ := regexp.Compile("^web:.*$")
		for scanner.Scan() {
			matched := r.MatchString(strings.TrimSpace(scanner.Text()))
			if err == nil && matched {
				installPumaCommand = false
				logger.Process("Do not return a process because a Procfile with process type 'web' already exists")
			}
		}
	}

	if installPumaCommand {
		logger.Process("Returning process type 'web' with command 'bundle exec puma'")

		return packit.Process{
			Type:    "web",
			Command: "bundle exec puma",
		}, nil
	}

	return packit.Process{}, nil
}
