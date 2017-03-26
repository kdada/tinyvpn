package stage

import (
	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

// NewClientMachine creates a client state machine
func NewClientMachine(host, account, password string) (state.Machine, error) {
	m := utils.NewStateMachine()

	// connect server
	conn, err := proto.CreateConnection(host, types.DefaultConverter)
	if err != nil {
		return nil, err
	}

	// add state handlers
	m.AddStateHandler(NewAuthHandler(m, conn, account, password))
	m.AddStateHandler(NewConfigHandler(m, conn))
	m.AddStateHandler(NewCommunicationHandler(m, conn))

	return m, nil
}

// Start starts a client state machine. Check machine state and terminate when machine
// falled in state Fail.
func Start(host, account, password string) (state.Machine, error) {
	machine, err := NewClientMachine(host, account, password)
	if err != nil {
		return nil, err
	}
	return machine, machine.Start(utils.Authentication, nil)
}
