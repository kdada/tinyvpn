package stage

import (
	"net"

	"github.com/kdada/tinyvpn/pkg/proto/types"
	"github.com/kdada/tinyvpn/pkg/tun"
)

var Device *tun.Device

func StartDevice(src, dest net.IP) error {
	device, err := tun.CreateDevice(src, dest)
	if err != nil {
		return err
	}
	Device = device
	return nil
}

func AddRoutes(routes [][5]byte) error {
	for _, route := range routes {
		r := &net.IPNet{
			IP:   route[:4:4],
			Mask: net.CIDRMask(int(route[4]), 32),
		}
		err := Device.AddRoute(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func Write(obj types.Packet) {
	Device.Write([]byte(obj))
}
