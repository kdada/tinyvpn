package stage

import (
	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/state"
)

// NewServerMachine creates a server state machine
func NewServerMachine(conn *proto.Connection) (state.Machine, error) {
	m := utils.NewStateMachine()

	// add state handlers
	m.AddStateHandler(NewAuthHandler(m, conn))
	m.AddStateHandler(NewConfigHandler(m, conn))
	m.AddStateHandler(NewCommunicationHandler(m, conn))

	return m, nil
}

func Start(conn *proto.Connection) (state.Machine, error) {
	machine, err := NewServerMachine(conn)
	if err != nil {
		return nil, err
	}
	return machine, machine.Start(utils.Authentication, nil)
}
