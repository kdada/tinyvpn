package stage

import (
	"log"

	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
	"github.com/kdada/tinyvpn/pkg/tun"
)

type CommunicationHandler struct {
	*utils.ConnectionStateHandler
}

func NewCommunicationHandler(machine state.Machine, conn *proto.Connection) *CommunicationHandler {
	handler := utils.NewConnectionStateHandler(utils.Communication, machine, conn)
	ah := &CommunicationHandler{
		ConnectionStateHandler: handler,
	}
	ah.AfterResume = ah.afterResume
	ah.Register(types.TypePacket, ah.communication)
	return ah
}

func (h *CommunicationHandler) afterResume() {
	go h.read()
}
func (h *CommunicationHandler) communication(obj interface{}) {
	Write(obj.(types.Packet))
}

func (h *CommunicationHandler) read() {
	for true {
		data := make([]byte, tun.MaxPacketSize)
		n, err := Device.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}
		err = h.Connection.Send(types.Packet(data[:n]))
		if err != nil {
			log.Println(err)
		}
	}
}
