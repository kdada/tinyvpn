package stage

import (
	"net"
	"strings"

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

func AddRoutes(routes []byte) error {
	for i := 0; i < len(routes); i += 5 {
		route := routes[i : i+5]
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

func AddDefaultRoute(ip string) error {
	ip = strings.Split(ip, ":")[0]
	_, s, err := net.ParseCIDR(ip + "/32")
	if err != nil {
		return err
	}
	return tun.AddRoute(s)
}

func Write(obj types.Packet) {
	Device.Write([]byte(obj))
}
