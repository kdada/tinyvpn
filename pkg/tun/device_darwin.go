// +build darwin

package tun

import (
	"io"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

// CreateDevice create a device via ip.
func CreateDevice(srcIP net.IP, destIP net.IP) (*Device, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}
	devName := ifce.Name()
	err = startDevice(devName, srcIP, destIP)
	if err != nil {
		return nil, err
	}
	dev := &Device{
		ReadWriteCloser: newNoPIReadWriteCloser(ifce.ReadWriteCloser),
		Name:            devName,
		SrcIP:           srcIP,
		DestIP:          destIP,
		Routes:          make([]*net.IPNet, 0, 10),
		addRoute: func(ip *net.IPNet) error {
			return addRoute(devName, ip)
		},
		deleteRoute: deleteRoute,
	}
	return dev, nil
}

// startDevice start the device
func startDevice(devName string, srcIP net.IP, destIP net.IP) error {
	cmd := exec.Command("ifconfig", devName, "inet", srcIP.String(), destIP.String(), "up")
	return cmd.Run()
}

// addRoute adds route to specified device
func addRoute(devName string, ip *net.IPNet) error {
	cmd := exec.Command("route", "add", ip.String(), "-interface", devName)
	return cmd.Run()
}

// deleteRoute deletes route from specified device
func deleteRoute(ip *net.IPNet) error {
	cmd := exec.Command("route", "delete", ip.String())
	return cmd.Run()
}

// noPIReadWriteCloser wraps a ReadWriteCloser and shields the PI flag.
// PI flag: 0x00 0x00 0x00 0x02
type noPIReadWriteCloser struct {
	io.ReadWriteCloser
	// rBuffer is read buffer
	rBuffer []byte
	// wBuffer is write buffer
	wBuffer []byte
}

// newNoPIReadWriteCloser create noPIReadWriteCloser
func newNoPIReadWriteCloser(rwc io.ReadWriteCloser) *noPIReadWriteCloser {
	p := &noPIReadWriteCloser{
		rwc,
		make([]byte, 4096),
		make([]byte, 4096),
	}
	// add pi to wBuffer
	copy(p.wBuffer, []byte{0, 0, 0, 2})
	return p
}

// Read reads a packet from original ReadWriteCloser
func (rwc *noPIReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = rwc.ReadWriteCloser.Read(rwc.rBuffer)
	if err == nil && n >= 4 {
		copy(p, rwc.rBuffer[4:n])
		n -= 4
	}
	return
}

// Write writes a packet to original ReadWriteCloser
func (rwc *noPIReadWriteCloser) Write(p []byte) (n int, err error) {
	copy(rwc.wBuffer[4:], p)
	return rwc.ReadWriteCloser.Write(rwc.wBuffer[:len(p)+4])
}

// Close closes ReadWriteCloser
func (rwc *noPIReadWriteCloser) Close() error {
	return rwc.ReadWriteCloser.Close()
}
