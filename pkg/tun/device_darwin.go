// +build darwin

package tun

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"

	"github.com/songgao/water"
)

const (
	// SizePI is the size of PI header. Darwin tun device will add pi header in front of a packet.
	// So the max size of buffer to store a packet is MaxPacketSize + SizePI
	SizePI = 4
)

var regIf = regexp.MustCompile(`gateway: *(\d+\.\d+\.\d+\.\d+)`)

// AddRoute adds route to default device.
func AddRoute(ip *net.IPNet) error {
	ip.IP = ip.IP.Mask(ip.Mask)
	cmd := exec.Command("route", "-n", "get", ip.IP.String())
	data, err := cmd.Output()
	if err != nil {
		return err
	}
	result := regIf.FindSubmatch(data)
	if len(result) != 2 {
		return fmt.Errorf("can't find interface name by net: %s", ip.String())
	}
	gw := string(result[1])
	cmd = exec.Command("route", "-n", "add", ip.String(), gw)
	return cmd.Run()
}

// DeleteRoute deletes a route
func DeleteRoute(ip *net.IPNet) error {
	ip.IP = ip.IP.Mask(ip.Mask)
	cmd := exec.Command("route", "-n", "delete", ip.String())
	return cmd.Run()
}

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
		make([]byte, MaxPacketSize+SizePI),
		make([]byte, MaxPacketSize+SizePI),
	}
	// add pi header to wBuffer
	copy(p.wBuffer, []byte{0, 0, 0, 2})
	return p
}

// Read reads a packet from original ReadWriteCloser
func (rwc *noPIReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = rwc.ReadWriteCloser.Read(rwc.rBuffer)
	if err != nil {
		return 0, err
	}
	if n <= SizePI {
		return 0, fmt.Errorf("bad packet with length: %d", n)
	}
	n = copy(p, rwc.rBuffer[SizePI:n])
	return
}

// Write writes a packet to original ReadWriteCloser
func (rwc *noPIReadWriteCloser) Write(p []byte) (n int, err error) {
	if len(p) < MinPacketSize || len(p) > MaxPacketSize {
		return 0, fmt.Errorf("bad packet length: %d", len(p))
	}
	copy(rwc.wBuffer[4:], p)
	n, err = rwc.ReadWriteCloser.Write(rwc.wBuffer[:len(p)+4])
	return n - 4, err
}

// Close closes ReadWriteCloser
func (rwc *noPIReadWriteCloser) Close() error {
	return rwc.ReadWriteCloser.Close()
}
