package tun

import (
	"fmt"
	"net"
)

type IPPacket []byte

// SrcIP returns srouce ip
func (ip IPPacket) SrcIP() net.IP {
	return net.IP(ip[12:16])
}

// DestIP returns destination ip
func (ip IPPacket) DestIP() net.IP {
	return net.IP(ip[16:20])
}

// Validate validates ip packet
func (ip IPPacket) Validate() error {
	// TODO(kdada): Validate header
	if len(ip) < 20 {
		return fmt.Errorf("ip packet at least has 20 bytes")
	}
	return nil
}
