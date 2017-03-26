package stage

import (
	"log"

	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

type CommunicationHandler struct {
	*utils.ConnectionStateHandler
}

func NewCommunicationHandler(machine state.Machine, conn *proto.Connection) *CommunicationHandler {
	handler := utils.NewConnectionStateHandler(utils.Communication, machine, conn)
	ah := &CommunicationHandler{
		ConnectionStateHandler: handler,
	}
	ah.Register(types.TypePacket, ah.packet)
	ah.ErrorHandler = ah.errorHandler
	return ah
}

func (h *CommunicationHandler) EnterState(m state.Machine, event state.Event, data interface{}) error {
	Register(data.(uint32), h.send)
	return h.ConnectionStateHandler.EnterState(m, event, data)
}
func (h *CommunicationHandler) packet(obj interface{}) {
	Write(obj.(types.Packet))
}

func (h *CommunicationHandler) send(obj interface{}) {
	err := h.Connection.Send(obj)
	if err != nil {
		log.Println(err)
	}
}
func (h *CommunicationHandler) errorHandler(reason string) {
	RetireIPByConnection(h.Connection)
}
