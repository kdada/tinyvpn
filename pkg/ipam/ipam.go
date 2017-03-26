/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package ipam

import (
	"fmt"
	"net"
	"sync"
)

// IPAM provides a capacity to manage ip range.
type IPAM struct {
	sync.Mutex
	Range *net.IPNet
	Last  net.IP
	Pool  []net.IP
}

// NewIPAM creates an IPAM to manage a scope of IPv4.
// The scope should have more IPs than one. Otherwise It will
// throw an error.
func NewIPAM(scope *net.IPNet) (*IPAM, error) {
	ones, bits := scope.Mask.Size()
	if ones >= bits-1 {
		return nil, fmt.Errorf("there is no IP to manage")
	}
	return &IPAM{
		Range: scope,
		Last:  scope.IP.Mask(scope.Mask),
		Pool:  make([]net.IP, 0, 10),
	}, nil
}

// Assign assigns an unused IP or an error when there is no
// more available IP.
func (m *IPAM) Assign() (net.IP, error) {
	m.Lock()
	defer m.Unlock()
	plen := len(m.Pool)
	if plen > 0 {
		ip := m.Pool[plen-1]
		m.Pool = m.Pool[:plen-1]
		return ip, nil
	}
	value := ConvertIPToInt(m.Last.To4())
	value++
	ip := ConvertIntToIP(value)
	if !m.Range.Contains(ip) {
		return nil, fmt.Errorf("no rest ip")
	}
	m.Last = ip
	return ip, nil
}

// Retire recycles an IP
func (m *IPAM) Retire(ip net.IP) error {
	m.Lock()
	defer m.Unlock()
	m.Pool = append(m.Pool, ip)
	return nil
}

// ConvertIntToIP converts a value to IP
func ConvertIntToIP(v uint32) net.IP {
	ip := make(net.IP, 4)
	for i := 0; i < 4 && v > 0; i++ {
		ip[3-i] = byte(v)
		v >>= 8
	}
	return ip
}

// ConvertIPToInt converts ip to int value
func ConvertIPToInt(ip net.IP) uint32 {
	value := uint32(0)
	for i := 0; i < 4; i++ {
		value = value<<8 | uint32(ip[i])
	}
	return value
}
