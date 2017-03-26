package ipam

import (
	"net"
	"reflect"
	"testing"
)

func TestIPAM(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.0.0.1/24")
	am, err := NewIPAM(ipnet)
	if err != nil {
		t.Fatal(err)
	}
	xip, err := am.Assign()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(xip)
	if !reflect.DeepEqual(xip.To4(), net.IP{10, 0, 0, 1}) {
		t.Fatal("error ip", xip)
	}
	t.Log(am)
	if !reflect.DeepEqual(am.Last.To4(), net.IP{10, 0, 0, 1}) {
		t.Fatal("error last ip", xip)
	}

	xip2, err := am.Assign()
	if err != nil {
		t.Fatal(err)
	}

	err = am.Retire(xip)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(am)
	if !reflect.DeepEqual(am.Last.To4(), net.IP{10, 0, 0, 2}) {
		t.Fatal("error last ip", xip)
	}
	if len(am.Pool) != 1 {
		t.Fatal("retire failed")
	}
	if !reflect.DeepEqual(am.Pool[0].To4(), net.IP{10, 0, 0, 1}) {
		t.Fatal("error pool ip", xip)
	}

	err = am.Retire(xip2)
	if err != nil {
		t.Fatal(err)
	}

	xip3, err := am.Assign()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(xip2, xip3) {
		t.Fatal("error assigned pool ip", xip)
	}
}
