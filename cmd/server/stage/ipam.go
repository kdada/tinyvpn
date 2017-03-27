package stage

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/kdada/tinyvpn/pkg/ipam"
	"github.com/kdada/tinyvpn/pkg/proto"
)

var IPAddressManager *ipam.IPAM
var ServerIP uint32
var Routes []byte
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
	buf := bytes.NewBuffer(make([]byte, 60000))
	buf.Write(ip)
	buf.WriteByte(32)

	data, err := ioutil.ReadFile("./out.txt")
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	reader := bufio.NewReader(bytes.NewReader(data))
	for true {
		l, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		arrs := strings.Split(l, "/")
		rip := net.ParseIP(arrs[0]).To4()
		x := arrs[1]
		m, err := strconv.Atoi(x[:len(x)-1])

		buf.Write(rip)
		buf.WriteByte(byte(m))
	}
	Routes = buf.Bytes()
	log.Println("routes:", len(Routes))

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
