package stage

import (
	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

type ConfigHandler struct {
	*utils.ConnectionStateHandler
	ip uint32
}

func NewConfigHandler(machine state.Machine, conn *proto.Connection) *ConfigHandler {
	handler := utils.NewConnectionStateHandler(utils.DownloadConfig, machine, conn)
	ah := &ConfigHandler{
		ConnectionStateHandler: handler,
	}
	ah.Register(types.TypeConfigRequest, ah.config)
	ah.ErrorHandler = ah.errorHandler
	return ah
}
func (h *ConfigHandler) EnterState(m state.Machine, event state.Event, data interface{}) error {
	h.ip = data.(uint32)
	return h.ConnectionStateHandler.EnterState(m, event, data)
}
func (h *ConfigHandler) config(obj interface{}) {
	h.Connection.Send(&types.Config{
		Routes: Routes,
	})
	h.Machine.Trigger(utils.DownloadConfigSucessfully, h.ip)
}

func (h *ConfigHandler) errorHandler(reason string) {
	RetireIPByConnection(h.Connection)
}
