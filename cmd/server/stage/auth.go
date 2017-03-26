package stage

import (
	"fmt"

	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

type AuthHandler struct {
	*utils.ConnectionStateHandler
}

func NewAuthHandler(machine state.Machine, conn *proto.Connection) *AuthHandler {
	handler := utils.NewConnectionStateHandler(utils.Authentication, machine, conn)
	ah := &AuthHandler{
		ConnectionStateHandler: handler,
	}
	ah.Register(types.TypeAuthentication, ah.auth)
	ah.ErrorHandler = ah.errorHandler
	return ah
}

func (h *AuthHandler) auth(obj interface{}) {
	data := obj.(*types.Authentication)
	if data.Account != "admin" || data.Password != "123456" {
		h.Connection.Close(fmt.Sprintln("invalid authentication", obj))
	}
	ip, err := AssignIPForConnection(h.Connection)
	if err != nil {
		h.Connection.Close(err.Error())
		return
	}
	h.Connection.Send(&types.Authorization{
		Version:  utils.ServerVersion,
		ServerIP: ServerIP,
		ClientIP: ip,
	})
	h.Machine.Trigger(utils.AuthenticateSucessfully, ip)
}

func (h *AuthHandler) errorHandler(reason string) {
	RetireIPByConnection(h.Connection)
}
