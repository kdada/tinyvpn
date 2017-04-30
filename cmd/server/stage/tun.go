package stage

import (
	"log"
	"net"

	"github.com/kdada/tinyvpn/cmd/utils"
	"github.com/kdada/tinyvpn/pkg/ipam"
	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/tun"
)

var Device tun.Device
var dispatcher map[uint32]utils.Handler

func StartDevice(ip *net.IPNet) error {
	device, err := tun.CreateDevice(ip.IP, ip.IP)
	if err != nil {
		return err
	}
	err = device.AddRoute(ip)
	if err != nil {
		return err
	}
	dispatcher = make(map[uint32]utils.Handler)
	Device = device
	go read()
	return nil
}

func Register(ip uint32, handler utils.Handler) {
	dispatcher[ip] = handler
}

func Unregister(ip uint32) {
	_, ok := dispatcher[ip]
	if ok {
		delete(dispatcher, ip)
	}
}

func read() {
	for true {
		data := make([]byte, tun.MaxPacketSize)
		n, err := Device.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}
		pack := tun.IPPacket(data[:n])
		dest := pack.DestIP()
		dip := ipam.ConvertIPToInt(dest)
		h, ok := dispatcher[dip]
		if ok {
			h(types.Packet(data[:n]))
		}
	}
}

func Write(obj types.Packet) {
	Device.Write([]byte(obj))
}
