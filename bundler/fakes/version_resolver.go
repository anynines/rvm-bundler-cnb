package fakes

import (
	"sync"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
)

type VersionResolver struct {
	LookupCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
			Bashcmd    bundler.BashCmd
		}
		Returns struct {
			Version string
			Err     error
		}
		Stub func(string, bundler.BashCmd) (string, error)
	}
}

func (f *VersionResolver) Lookup(param1 string, param2 bundler.BashCmd) (string, error) {
	f.LookupCall.mutex.Lock()
	defer f.LookupCall.mutex.Unlock()
	f.LookupCall.CallCount++
	f.LookupCall.Receives.WorkingDir = param1
	f.LookupCall.Receives.Bashcmd = param2
	if f.LookupCall.Stub != nil {
		return f.LookupCall.Stub(param1, param2)
	}
	return f.LookupCall.Returns.Version, f.LookupCall.Returns.Err
}
