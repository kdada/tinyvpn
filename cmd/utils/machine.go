package utils

import "github.com/kdada/tinyvpn/pkg/state"

const (
	Authentication state.State = "Authentication"
	DownloadConfig state.State = "DownloadConfig"
	Communication  state.State = "Communication"
	Fail           state.State = "Fail"
)

const (
	AuthenticateSucessfully   state.Event = "AuthenticateSucessfully"
	DownloadConfigSucessfully state.Event = "DownloadConfigSucessfully"
	SomethingFail             state.Event = "SomethingFail"
)

// NewClientMachine creates a client state machine
func NewStateMachine() state.Machine {
	machine := state.NewBaseMachine()
	machine.AddTransition(AuthenticateSucessfully, Authentication, DownloadConfig)
	machine.AddTransition(DownloadConfigSucessfully, DownloadConfig, Communication)
	machine.AddTransition(SomethingFail, Authentication, Fail)
	machine.AddTransition(SomethingFail, DownloadConfig, Fail)
	machine.AddTransition(SomethingFail, Communication, Fail)
	return machine
}
