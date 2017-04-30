package tun

import (
	"fmt"
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
	_ = IPPacket(data)
}

func (md *MasqDevice) masqIn(data []byte) {
	_ = IPPacket(data)

}
