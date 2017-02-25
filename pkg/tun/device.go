package tun

import (
	"io"
	"net"
)

// Device describes an tunnel device. Read/Write one ip packet at once.
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
