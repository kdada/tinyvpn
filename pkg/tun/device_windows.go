// +build windows

package tun

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

const (
	// SizeEthernetHeader is used to extend packet buffer to store ethernet header.
	SizeEthernetHeader = 22
)

var regIf = regexp.MustCompile(`1 *(\d+.\d+.\d+.\d+)`)

// AddRoute adds route to default device.
func AddRoute(ip *net.IPNet) error {
	ip.IP = ip.IP.Mask(ip.Mask)
	cmd := exec.Command("pathping", "-n", "-w", "1", "-h", "1", "-q", "1", ip.IP.String())
	err := cmd.Run()
	if err != nil {
		return err
	}
	data, err := cmd.Output()
	if err != nil {
		return err
	}
	result := regIf.FindSubmatch(data)
	if len(result) != 2 {
		return fmt.Errorf("can't find interface name by net: %s", ip.String())
	}
	sIf := string(result[1])
	cmd := exec.Command("route", "add", ip.String(), sIf)
	return cmd.Run()
}

// CreateDevice create a device via ip. Must install tap driver before creating device on windows.
func CreateDevice(srcIP net.IP, destIP net.IP) (*Device, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID: "tap0901",
			Network:     srcIP.String() + "/32",
		},
	})
	if err != nil {
		return nil, err
	}
	devName := ifce.Name()
	srcMac, index, err := startDevice(srcIP, destIP)
	if err != nil {
		return nil, err
	}
	destMac := make([]byte, 6)
	copy(destMac, srcMac)
	destMac[5]++
	dev := &Device{
		ReadWriteCloser: newIPReadWriteCloser(srcMac, destMac, srcIP, destIP, ifce.ReadWriteCloser),
		Name:            devName,
		SrcIP:           srcIP,
		DestIP:          destIP,
		Routes:          make([]*net.IPNet, 0, 10),
		addRoute: func(r *net.IPNet) error {
			return addRoute(index, destIP, r)
		},
		deleteRoute: deleteRoute,
	}
	return dev, nil
}

// tapRegexp find the index of tunnel interface
var tapRegexp = regexp.MustCompile(`(\d+).*?TAP-Windows Adapter V9`)

// startDevice start the device
func startDevice(srcIP net.IP, destIP net.IP) (net.HardwareAddr, string, error) {
	// set device
	cmd := exec.Command("route", "print", "tap")
	output, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}
	// find index
	strs := tapRegexp.FindStringSubmatch(string(output))
	if len(strs) != 2 {
		return nil, "", fmt.Errorf("can't find adapter")
	}
	index, err := strconv.Atoi(strs[1])
	if err != nil {
		return nil, "", err
	}

	// find interface
	ifces, err := net.Interfaces()
	if err != nil {
		return nil, "", err
	}
	var tunIf *net.Interface
	for _, ifce := range ifces {
		if ifce.Index == index {
			tunIf = &ifce
			break
		}
	}
	if tunIf == nil {
		return nil, "", fmt.Errorf("no adapter with index %d", index)
	}

	// set ip to interface
	cmd = exec.Command("netsh", "interface", "ip", "set", "address", "name=", tunIf.Name, "source=",
		"static", "addr=", srcIP.String(), "mask=", "255.255.255.255", "gateway=", destIP.String())
	return tunIf.HardwareAddr, strconv.Itoa(tunIf.Index), cmd.Run()
}

// addRoute adds route to specified device
func addRoute(index string, ip net.IP, r *net.IPNet) error {
	cmd := exec.Command("route", "add", r.String(), ip.String(), "IF", index)
	return cmd.Run()
}

// deleteRoute deletes route from specified device
func deleteRoute(r *net.IPNet) error {
	cmd := exec.Command("route", "delete", r.String())
	return cmd.Run()
}

// ipReadWriteCloser handles ethernet packets and filter ip packets.
type ipReadWriteCloser struct {
	srcMac  net.HardwareAddr
	destMac net.HardwareAddr
	srcIP   net.IP
	destIP  net.IP
	io.ReadWriteCloser
	buffer []byte
}

// newIPReadWriteCloser create ipReadWriteCloser
func newIPReadWriteCloser(srcMac, destMac net.HardwareAddr, srcIP, destIP net.IP, rwc io.ReadWriteCloser) *ipReadWriteCloser {
	return &ipReadWriteCloser{
		srcMac,
		destMac,
		srcIP,
		destIP,
		rwc,
		make([]byte, MaxPacketSize+SizeEthernetHeader),
	}
}

// Read reads an ip packet from device.
func (rwc *ipReadWriteCloser) Read(p []byte) (n int, err error) {
	for true {
		n, err = rwc.ReadWriteCloser.Read(rwc.buffer)
		if err != nil {
			return n, err
		}
		frame := ethernet.Frame(rwc.buffer[:n])
		typ := frame.Ethertype()
		if compareBytes(ethernet.IPv4[:], typ[:]) {
			pkg := frame.Payload()
			n = copy(p, pkg)
			return
		}
		if compareBytes(ethernet.ARP[:], typ[:]) {
			// handle arp requests
			// windows will send arp request before normal ip packet
			err = rwc.ARPReply(frame)
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}

// arp response fixed bytes
var arp = []byte{0x08, 0x06, 0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x02}

// ARPReply replies arp request
func (rwc *ipReadWriteCloser) ARPReply(req []byte) error {
	if compareBytes(req[38:42], rwc.srcIP.To4()) {
		// windows will send a arp request for its static ip and check if anyone has holded the ip.
		// we need ignore the arp request.
		return nil
	}
	// generate arp response
	resp := make([]byte, 42)
	copy(resp[:6], req[6:12])
	copy(resp[6:12], rwc.destMac)
	copy(resp[12:22], arp)
	copy(resp[22:28], rwc.destMac)
	copy(resp[28:32], req[38:42])
	copy(resp[32:38], req[22:28])
	copy(resp[38:42], req[28:32])
	n, err := rwc.ReadWriteCloser.Write(resp)
	if err != nil {
		return err
	}
	if n != len(resp) {
		return fmt.Errorf("can't send arp response")
	}
	return nil
}

// Write writes an ip packet to device. It packs ip packet with an ethernet packet
func (rwc *ipReadWriteCloser) Write(p []byte) (n int, err error) {
	if len(p) < MinPacketSize || len(p) > MaxPacketSize {
		return 0, fmt.Errorf("bad packet length: %d", len(p))
	}
	frame := ethernet.Frame([]byte{})
	frame.Prepare(rwc.destMac, rwc.srcMac, ethernet.NotTagged, ethernet.IPv4, len(p))
	copy(frame[len(frame)-len(p):], p)
	return rwc.ReadWriteCloser.Write(frame)
}

// Close closes ReadWriteCloser
func (rwc *ipReadWriteCloser) Close() error {
	return rwc.ReadWriteCloser.Close()
}

// compareBytes compares whether the two slices equal.
func compareBytes(typ1, typ2 []byte) bool {
	if len(typ1) != len(typ2) {
		return false
	}
	for i := 0; i < len(typ1); i++ {
		if typ1[i] != typ2[i] {
			return false
		}
	}
	return true
}
