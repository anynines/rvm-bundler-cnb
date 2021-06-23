package fakes

import "sync"

type VersionResolver struct {
	LookupCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			Version string
			Err     error
		}
		Stub func(string) (string, error)
	}
}

func (f *VersionResolver) Lookup(param1 string) (string, error) {
	f.LookupCall.Lock()
	defer f.LookupCall.Unlock()
	f.LookupCall.CallCount++
	f.LookupCall.Receives.WorkingDir = param1
	if f.LookupCall.Stub != nil {
		return f.LookupCall.Stub(param1)
	}
	return f.LookupCall.Returns.Version, f.LookupCall.Returns.Err
}
