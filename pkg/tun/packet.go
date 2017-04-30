package tun

import (
	"net"
)

const (
	ProtocolICMP byte = 0x1
	ProtocolTCP  byte = 0x6
	ProtocolUDP  byte = 0x11
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
		result += uint32(data[len(data)-1]) << 8
	}
	for result&0xffff0000 != 0 {
		result = result>>16 + result&0x0000ffff
	}
	return ^uint16(result)
}

func ipsum(srcIP, destIP net.IP, protocol byte, data []byte) uint16 {
	external := make([]byte, 12)
	copy(external, srcIP)
	copy(external[4:], destIP)
	external[9] = protocol
	length := len(data)
	external[10] = byte(length >> 8)
	external[11] = byte(length)
	er := sum(external)
	r := sum(data)
	result := uint32(^er) + uint32(^r)
	for result&0xffff0000 != 0 {
		result = result>>16 + result&0x0000ffff
	}
	return uint16(result)
}

type TCPPacket []byte

func (tcp TCPPacket) Port() uint16 {
	return uint16(tcp[0])<<8 + uint16(tcp[1])
}

func (tcp TCPPacket) SetPort(port uint16) {
	tcp[0] = byte(port >> 8)
	tcp[1] = byte(port)
}

// Resum recalculates the check sum of current packet
func (tcp TCPPacket) Resum(srcIP, destIP net.IP) {
	tcp[16] = 0
	tcp[17] = 0
	result := ipsum(srcIP, destIP, ProtocolTCP, tcp)
	tcp[16] = byte(result >> 8)
	tcp[17] = byte(result)
}

type UDPPacket []byte

func (udp UDPPacket) Port() uint16 {
	return uint16(udp[0])<<8 + uint16(udp[1])
}

func (udp UDPPacket) SetPort(port uint16) {
	udp[0] = byte(port >> 8)
	udp[1] = byte(port)
}

// Resum recalculates the check sum of current packet
func (udp UDPPacket) Resum(srcIP, destIP net.IP) {
	udp[6] = 0
	udp[7] = 0
	result := ipsum(srcIP, destIP, ProtocolTCP, udp)
	udp[6] = byte(result >> 8)
	udp[7] = byte(result)
}

type ICMPPacket []byte

func (icmp ICMPPacket) ID() uint16 {
	return uint16(icmp[4])<<8 + uint16(icmp[5])
}

func (icmp ICMPPacket) SetID(id uint16) {
	icmp[4] = byte(id >> 8)
	icmp[5] = byte(id)
}

// Resum recalculates the check sum of current packet
func (icmp ICMPPacket) Resum() {
	icmp[2] = 0
	icmp[3] = 0
	result := sum(icmp)
	icmp[2] = byte(result >> 8)
	icmp[3] = byte(result)
}
