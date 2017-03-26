package utils

import (
	"log"

	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/state"
)

type Handler func(obj interface{})
type ConnectionStateHandler struct {
	*state.BaseHandler
	Machine        state.Machine
	Connection     *proto.Connection
	Handlers       map[uint16]Handler
	DefaultHandler func(uint16, interface{})
	ErrorHandler   func(string)
	AfterResume    func()
}

func NewConnectionStateHandler(s state.State, machine state.Machine, conn *proto.Connection) *ConnectionStateHandler {
	return &ConnectionStateHandler{
		BaseHandler:    state.NewBaseHandler(s),
		Machine:        machine,
		Connection:     conn,
		Handlers:       make(map[uint16]Handler),
		DefaultHandler: func(typ uint16, data interface{}) { log.Println(typ, data) },
		ErrorHandler:   func(reason string) { log.Println(reason) },
	}
}

func (h *ConnectionStateHandler) EnterState(m state.Machine, event state.Event, data interface{}) error {
	log.Println("enter state", h.State())
	go h.do()
	return nil
}

func (h *ConnectionStateHandler) ExitState(m state.Machine, event state.Event, data interface{}) error {
	h.Connection.Suspend()
	log.Println("exit state", h.State())
	return nil
}
func (h *ConnectionStateHandler) do() {
	err := h.Connection.Resume(&proto.ConnectionHandler{
		Received: h.dispatch,
		Closed:   h.ErrorHandler,
	})
	if err != nil {
		log.Println(err)
		h.ErrorHandler(err.Error())
		return
	}
	if h.AfterResume != nil {
		h.AfterResume()
	}
}

func (h *ConnectionStateHandler) dispatch(typ uint16, obj interface{}) {
	handler, ok := h.Handlers[typ]
	if ok {
		handler(obj)
		return
	}
	h.DefaultHandler(typ, obj)
}

func (h *ConnectionStateHandler) Register(typ uint16, f func(obj interface{})) {
	h.Handlers[typ] = f
}
