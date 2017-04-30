package tun

import (
	"log"
	"net"
	"sync"
)

// Handler handles a packet from device
type Handler func(packet IPPacket)

// DeviceManager manages a tun device and send/recv packet to/from device
type DeviceManager struct {
	lock         sync.RWMutex
	Device       Device
	Default      Handler
	DestHandlers map[uint32]Handler
	SendBuffer   chan IPPacket
}

// NewDeviceManager creates DeviceManager
func NewDeviceManager(device Device, sendBufferSize int) *DeviceManager {
	return &DeviceManager{
		Device:     device,
		SendBuffer: make(chan IPPacket, sendBufferSize),
	}
}

// Run starts to read/write packet from/to device
func (dm *DeviceManager) Run(stopCh chan bool) {
	go dm.read(stopCh)
	go dm.write(stopCh)
}

// read reads packets from device
func (dm *DeviceManager) read(stopCh chan bool) {
	for true {
		data := make([]byte, MaxPacketSize)
		n, err := dm.Device.Read(data)
		if err != nil {
			log.Println(err)
		} else if dm.DestHandlers != nil {
			dm.lock.RLock()
			packet := IPPacket(data[:n])
			dip := convertIPToInt(packet.DestIP())
			if h, ok := dm.DestHandlers[dip]; ok {
				h(packet)
			} else if dm.Default != nil {
				dm.Default(packet)
			}
			dm.lock.RUnlock()
		} else if dm.Default != nil {
			dm.Default(IPPacket(data[:n]))
		}
		select {
		case <-stopCh:
			close(dm.SendBuffer)
			return
		default:
		}
	}
}

// write writes packets to device
func (dm *DeviceManager) write(stopCh chan bool) {
	for true {
		packet := <-dm.SendBuffer
		if packet != nil {
			dm.Device.Write(packet)
		}
		select {
		case <-stopCh:
			return
		default:
		}
	}
}

// AddDestHandler adds a destination handler for handling packets
func (dm *DeviceManager) AddDestHandler(ip net.IP, handler Handler) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	if dm.DestHandlers == nil {
		dm.DestHandlers = make(map[uint32]Handler)
	}
	value := convertIPToInt(ip)
	dm.DestHandlers[value] = handler
}

// DeleteHandler deletes a handler
func (dm *DeviceManager) DeleteHandler(ip net.IP) {
	if dm.DestHandlers != nil {
		dm.lock.Lock()
		defer dm.lock.Unlock()
		value := convertIPToInt(ip)
		delete(dm.DestHandlers, value)
	}
}

// Send sends Packet to device
func (dm *DeviceManager) Send(packet IPPacket) {
	dm.SendBuffer <- packet
}

// convertIPToInt converts ip to int value
func convertIPToInt(ip net.IP) uint32 {
	value := uint32(0)
	for i := 0; i < 4; i++ {
		value = value<<8 | uint32(ip[i])
	}
	return value
}
