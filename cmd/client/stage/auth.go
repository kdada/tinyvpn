package stage

import (
	"log"

	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/ipam"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/state"
)

type AuthHandler struct {
	*utils.ConnectionStateHandler
	Account  string
	Password string
}

func NewAuthHandler(machine state.Machine, conn *proto.Connection, account string, password string) *AuthHandler {
	handler := utils.NewConnectionStateHandler(utils.Authentication, machine, conn)
	ah := &AuthHandler{
		ConnectionStateHandler: handler,
		Account:                account,
		Password:               password,
	}
	ah.AfterResume = ah.afterResume
	ah.Register(types.TypeAuthorization, ah.auth)
	return ah
}

func (h *AuthHandler) afterResume() {
	h.Connection.Send(&types.Authentication{
		Version:  utils.ClientVersion,
		Account:  h.Account,
		Password: h.Password,
	})
}

func (h *AuthHandler) auth(obj interface{}) {
	data := obj.(*types.Authorization)
	err := StartDevice(ipam.ConvertIntToIP(data.ClientIP), ipam.ConvertIntToIP(data.ServerIP))
	if err != nil {
		log.Println(err)
		return
	}
	h.Machine.Trigger(utils.AuthenticateSucessfully, nil)
}
