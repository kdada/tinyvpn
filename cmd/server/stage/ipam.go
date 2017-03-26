package stage

import (
	"log"
	"net"
	"os"

	"github.com/kdada/tinyvpn/pkg/ipam"
	"github.com/kdada/tinyvpn/pkg/proto"
)

var IPAddressManager *ipam.IPAM
var ServerIP uint32
var Routes [][5]byte
var connectionIPs map[*proto.Connection]uint32

func init() {
	connectionIPs = make(map[*proto.Connection]uint32)
	_, ips, err := net.ParseCIDR("10.0.0.1/24")
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	IPAddressManager, err = ipam.NewIPAM(ips)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	ip, err := IPAddressManager.Assign()
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	// server ip
	ServerIP = ipam.ConvertIPToInt(ip)

	// add routes
	r := [5]byte{}
	r[4] = 32
	copy(r[:], ip)
	Routes = [][5]byte{r}

	ips = &net.IPNet{
		IP:   ip,
		Mask: ips.Mask,
	}
	err = StartDevice(ips)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}

func AssignIPForConnection(conn *proto.Connection) (uint32, error) {
	ip, err := IPAddressManager.Assign()
	if err != nil {
		return 0, err
	}
	value := ipam.ConvertIPToInt(ip)
	connectionIPs[conn] = value
	return value, nil
}

func RetireIPByConnection(conn *proto.Connection) {
	ip, ok := connectionIPs[conn]
	if ok {
		IPAddressManager.Retire(ipam.ConvertIntToIP(ip))
		delete(connectionIPs, conn)
	}
}
