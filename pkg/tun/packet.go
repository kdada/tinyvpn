package tun

import (
	"net"
)

type IPPacket []byte

// HeaderLength returns the length of header in bytes
func (ip IPPacket) HeaderLength() int {
	return 4 * int(ip[0]&0x0f)
}

// Length returns total length of the packet
func (ip IPPacket) Length() int {
	return int(ip[2]) << 8 & int(ip[3])
}

// Payload returns packet body
func (ip IPPacket) Payload() []byte {
	return ip[ip.HeaderLength():ip.Length()]
}

// Protocol returns the payload protocol
func (ip IPPacket) Protocol() byte {
	return ip[9]
}

// SrcIP returns srouce ip
func (ip IPPacket) SrcIP() net.IP {
	return net.IP(ip[12:16])
}

// SetSrcIP sets the src ip
func (ip IPPacket) SetSrcIP(newIP net.IP) {
	copy(ip[12:16], newIP)
}

// DestIP returns destination ip
func (ip IPPacket) DestIP() net.IP {
	return net.IP(ip[16:20])
}

// SetDestIP sets the dest ip
func (ip IPPacket) SetDestIP(newIP net.IP) {
	copy(ip[16:20], newIP)
}

// Validate validates wether the packet is valid
func (ip IPPacket) Validate() bool {
	result := sum(ip[:ip.HeaderLength()])
	return result == 0xffff
}

// Resum recalculates the check sum of current packet
func (ip IPPacket) Resum() {
	ip[10] = 0
	ip[11] = 0
	result := sum(ip[:ip.HeaderLength()])
	ip[10] = byte(result >> 8)
	ip[11] = byte(result)
}

// sum calculates check sum of data
func sum(data []byte) uint16 {
	var result uint32 = 0
	for i := 0; i < len(data); i += 2 {
		result += uint32(data[i])<<8 + uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		result += uint32(data[len(data)-1])
	}
	for result&0xffff0000 != 0 {
		result = result>>16 + result&0x0000ffff
	}
	return ^uint16(result)
}
