// +build linux

package tun

import (
	"fmt"
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
	err = startDevice(devName, srcIP)
	if err != nil {
		return nil, err
	}
	dev := &Device{
		ReadWriteCloser: ifce.ReadWriteCloser,
		Name:            devName,
		SrcIP:           srcIP,
		DestIP:          destIP,
		Routes:          make([]*net.IPNet, 0, 10),
		addRoute: func(r *net.IPNet) error {
			return addRoute(devName, r)
		},
		deleteRoute: deleteRoute,
	}
	return dev, nil
}

// startDevice start the device
func startDevice(devName string, srcIP net.IP) error {
	cmd := exec.Command("ip", "addr", "add", srcIP.String(), "dev", devName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("ip", "link", "set", devName, "up")
	return cmd.Run()
}

// addRoute adds route to specified device
func addRoute(devName string, r *net.IPNet) error {
	cmd := exec.Command("ip", "r", "add", r.String(), "dev", devName)
	return cmd.Run()
}

// deleteRoute deletes route from specified device
func deleteRoute(r *net.IPNet) error {
	cmd := exec.Command("ip", "r", "delete", r.String())
	return cmd.Run()
}

// ipReadWriteCloser handles ip packets
type ipReadWriteCloser struct {
	io.ReadWriteCloser
}

// Write writes an ip packet to device. It packs ip packet with an ethernet packet
func (rwc *ipReadWriteCloser) Write(p []byte) (n int, err error) {
	if len(p) < MinPacketSize || len(p) > MaxPacketSize {
		return 0, fmt.Errorf("bad packet length: %d", len(p))
	}
	return rwc.ReadWriteCloser.Write(p)
}
