package fakes

import (
	"sync"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type PumaInstaller struct {
	CreatePumaProcessCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context       packit.BuildContext
			Configuration bundler.Configuration
			Logger        scribe.Logger
		}
		Returns struct {
			Process packit.Process
			Error   error
		}
		Stub func(packit.BuildContext, bundler.Configuration, scribe.Logger) (packit.Process, error)
	}
	InstallPumaCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Context       packit.BuildContext
			Configuration bundler.Configuration
			Logger        scribe.Logger
		}
		Returns struct {
			Error error
		}
		Stub func(packit.BuildContext, bundler.Configuration, scribe.Logger) error
	}
}

func (f *PumaInstaller) CreatePumaProcess(param1 packit.BuildContext, param2 bundler.Configuration, param3 scribe.Logger) (packit.Process, error) {
	f.CreatePumaProcessCall.mutex.Lock()
	defer f.CreatePumaProcessCall.mutex.Unlock()
	f.CreatePumaProcessCall.CallCount++
	f.CreatePumaProcessCall.Receives.Context = param1
	f.CreatePumaProcessCall.Receives.Configuration = param2
	f.CreatePumaProcessCall.Receives.Logger = param3
	if f.CreatePumaProcessCall.Stub != nil {
		return f.CreatePumaProcessCall.Stub(param1, param2, param3)
	}
	return f.CreatePumaProcessCall.Returns.Process, f.CreatePumaProcessCall.Returns.Error
}
func (f *PumaInstaller) InstallPuma(param1 packit.BuildContext, param2 bundler.Configuration, param3 scribe.Logger) error {
	f.InstallPumaCall.mutex.Lock()
	defer f.InstallPumaCall.mutex.Unlock()
	f.InstallPumaCall.CallCount++
	f.InstallPumaCall.Receives.Context = param1
	f.InstallPumaCall.Receives.Configuration = param2
	f.InstallPumaCall.Receives.Logger = param3
	if f.InstallPumaCall.Stub != nil {
		return f.InstallPumaCall.Stub(param1, param2, param3)
	}
	return f.InstallPumaCall.Returns.Error
}
