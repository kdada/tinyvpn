package stage

import (
	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

type ConfigHandler struct {
	*utils.ConnectionStateHandler
}

func NewConfigHandler(machine state.Machine, conn *proto.Connection) *ConfigHandler {
	handler := utils.NewConnectionStateHandler(utils.DownloadConfig, machine, conn)
	ah := &ConfigHandler{
		ConnectionStateHandler: handler,
	}
	ah.AfterResume = ah.afterResume
	ah.Register(types.TypeConfig, ah.config)
	return ah
}
func (h *ConfigHandler) afterResume() {
	h.Connection.Send(&types.ConfigRequest{})
}
func (h *ConfigHandler) config(obj interface{}) {
	data := obj.(*types.Config)
	AddRoutes(data.Routes)
	h.Machine.Trigger(utils.DownloadConfigSucessfully, data)
}
