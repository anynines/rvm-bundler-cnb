package fakes

import "sync"

type BashCmd struct {
	RunBashCmdCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Command    string
			WorkingDir string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *BashCmd) RunBashCmd(param1 string, param2 string) (string, error) {
	f.RunBashCmdCall.mutex.Lock()
	defer f.RunBashCmdCall.mutex.Unlock()
	f.RunBashCmdCall.CallCount++
	f.RunBashCmdCall.Receives.Command = param1
	f.RunBashCmdCall.Receives.WorkingDir = param2
	if f.RunBashCmdCall.Stub != nil {
		return f.RunBashCmdCall.Stub(param1, param2)
	}
	return f.RunBashCmdCall.Returns.String, f.RunBashCmdCall.Returns.Error
}
