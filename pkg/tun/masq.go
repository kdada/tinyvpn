package tun

import (
	"fmt"
	"log"
	"math"
	"net"
)

type Mapper struct {
	Out map[uint64]uint16
	In  map[uint16]uint64
	ID  uint16
}

func (m *Mapper) MapOut(ip net.IP, id uint16) (uint16, error) {
	if m.ID == math.MaxUint16 {
		return 0, fmt.Errorf("id overflow")
	}
	key := uint64(convertIPToInt(ip))<<16 + uint64(id)
	masqID, ok := m.Out[key]
	if !ok {
		m.ID++
		masqID = m.ID
		m.Out[key] = masqID
		m.In[masqID] = key
	}
	return masqID, nil
}

func (m *Mapper) MapIn(id uint16) (net.IP, uint16, error) {
	key, ok := m.In[id]
	if !ok {
		return nil, 0, fmt.Errorf("invalid id: %d", id)
	}
	return convertIntToIP(uint32(key >> 16)), uint16(key & 0xffff), nil
}

func NewMapper(start uint16) *Mapper {
	return &Mapper{
		Out: make(map[uint64]uint16),
		In:  make(map[uint16]uint64),
		ID:  start,
	}
}

// MasqDevice masquerade ip with identitier
type MasqDevice struct {
	Device
	TCPMapper  *Mapper
	UDPMapper  *Mapper
	ICMPMapper *Mapper
}

// NewMasqDevice creates a MasqDevice
func NewMasqDevice(dev Device) Device {
	return &MasqDevice{
		Device:     dev,
		TCPMapper:  NewMapper(10000),
		UDPMapper:  NewMapper(10000),
		ICMPMapper: NewMapper(0),
	}
}

// Read a packet
func (md *MasqDevice) Read(p []byte) (n int, err error) {
	n, err = md.Device.Read(p)
	if err != nil {
		return
	}
	md.masqOut(p[:n])
	return
}

// Write a packet
func (md *MasqDevice) Write(p []byte) (n int, err error) {
	md.masqIn(p)
	return md.Device.Write(p)
}

func (md *MasqDevice) masqOut(data []byte) {
	packet := IPPacket(data)
	switch packet.Protocol() {
	case ProtocolTCP:
		payload := TCPPacket(packet.Payload())
		port, err := md.TCPMapper.MapOut(packet.SrcIP(), payload.SrcPort())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetSrcIP(md.DeviceIP())
		payload.SetSrcPort(port)
		packet.Resum()
		payload.Resum(md.DeviceIP(), packet.DestIP())
	case ProtocolUDP:
		payload := UDPPacket(packet.Payload())
		port, err := md.UDPMapper.MapOut(packet.SrcIP(), payload.SrcPort())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetSrcIP(md.DeviceIP())
		payload.SetSrcPort(port)
		packet.Resum()
		payload.Resum(md.DeviceIP(), packet.DestIP())
	case ProtocolICMP:
		payload := ICMPPacket(packet.Payload())
		id, err := md.ICMPMapper.MapOut(packet.SrcIP(), payload.ID())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetSrcIP(md.DeviceIP())
		payload.SetID(id)
		packet.Resum()
		payload.Resum()
	}
}

func (md *MasqDevice) masqIn(data []byte) {
	packet := IPPacket(data)
	switch packet.Protocol() {
	case ProtocolTCP:
		payload := TCPPacket(packet.Payload())
		ip, port, err := md.TCPMapper.MapIn(payload.DestPort())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetDestIP(ip)
		payload.SetDestPort(port)
		packet.Resum()
		payload.Resum(packet.SrcIP(), ip)
	case ProtocolUDP:
		payload := UDPPacket(packet.Payload())
		ip, port, err := md.UDPMapper.MapIn(payload.DestPort())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetDestIP(ip)
		payload.SetDestPort(port)
		packet.Resum()
		payload.Resum(packet.SrcIP(), ip)
	case ProtocolICMP:
		payload := ICMPPacket(packet.Payload())
		ip, id, err := md.ICMPMapper.MapIn(payload.ID())
		if err != nil {
			log.Println(err)
			return
		}
		packet.SetDestIP(ip)
		payload.SetID(id)
		packet.Resum()
		payload.Resum()
	}
}
