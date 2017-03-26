package tun

import (
	"io"
	"net"
)

const (
	// MinPacketSize limits the min size of a packet in ethernet payload.
	// Ignore 802.1Q tag.
	MinPacketSize = 46
	// MaxPacketSize limits the max size of a packet in ethernet payload.
	MaxPacketSize = 1500
)

// Device describes an tunnel device. Read/Write one ip packet at once.
// Must not read/write data in parallel.
// If the read buffer is smaller than received packet length, device will truncate
// the packet and copy to read buffer. Then drop the packet.
// If the write buffer is larger than MaxPacketSize or smaller than MinPacketSize, device will drop the packet and
// return an error.
type Device struct {
	io.ReadWriteCloser
	// Name is device name
	Name string
	// SrcIP is the local ip of device
	SrcIP net.IP
	// DestIP is the remote ip of device
	DestIP net.IP
	// Routes contains all routes via the device
	Routes []*net.IPNet
	// addRoute add a route to system route table
	addRoute func(r *net.IPNet) error
	// deleteRoute delete a route from system route table
	deleteRoute func(r *net.IPNet) error
}

// AddRoute adds route for device
func (d *Device) AddRoute(r *net.IPNet) error {
	r = &net.IPNet{
		IP:   r.IP.Mask(r.Mask),
		Mask: r.Mask,
	}
	err := d.addRoute(r)
	if err != nil {
		return err
	}
	d.Routes = append(d.Routes, r)
	return nil
}

// ClearRoutes clears all routes
func (d *Device) ClearRoutes() error {
	for i, r := range d.Routes {
		err := d.deleteRoute(r)
		if err != nil {
			d.Routes = d.Routes[i:]
			return err
		}
	}
	d.Routes = make([]*net.IPNet, 0)
	return nil
}

// Close closes the device
func (d *Device) Close() error {
	err := d.ClearRoutes()
	if err != nil {
		return err
	}
	return d.ReadWriteCloser.Close()
}
