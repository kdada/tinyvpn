package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/kdada/tinyvpn/pkg/tun"
	"github.com/xtaci/kcp-go"
)

var server string
var local string
var remote string
var route string

func init() {
	flag.StringVar(&server, "s", "", "host:port e.g. 22.22.22.22:9989")
	flag.StringVar(&local, "l", "", "local ip e.g. 10.0.0.1")
	flag.StringVar(&remote, "r", "", "remote ip e.g. 10.0.0.2")
	flag.StringVar(&route, "d", "", "default route e.g. 10.0.0.0/24")
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate)
	log.Println("tinyvpn server started")
	log.Println(local, remote)
	device, err := tun.CreateDevice(net.ParseIP(local), net.ParseIP(remote))
	if err != nil {
		log.Fatalln(err)
	}
	defer device.Close()

	// add route
	_, r, err := net.ParseCIDR(route)
	if err != nil {
		log.Fatalln(err)
	}
	device.AddRoute(r)

	handle(device)
	listen(device)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig
	log.Println("tinyvpn server stoped")
}

var conns = make(map[uint32]net.Conn)

func handle(device *tun.Device) {
	go func() {
		buf := make([]byte, 4096)
		for true {
			rc, err := device.Read(buf)
			if err != nil {
				log.Println("tunnel", "read error", err)
				break
			}
			ipp := tun.IPPacket(buf[:rc])
			ip := convertIP(ipp.DestIP())
			conn, ok := conns[ip]
			if ok {
				wc, err := conn.Write(buf[:rc])
				if err == nil && wc != rc {
					err = fmt.Errorf("read count: %d write count: %d", rc, wc)
				}
				if err != nil {
					log.Println(conn.RemoteAddr(), "closed connection", err)
					conn.Close()
					delete(conns, ip)
				}
			}
		}
	}()
}

func listen(device *tun.Device) {
	listener, err := kcp.ListenWithOptions(server, nil, 10, 3)
	if err != nil {
		log.Fatalln(err)
	}
	listener.SetReadBuffer(4096 * 1024)
	listener.SetWriteBuffer(4096 * 1024)
	go func() {
		for true {
			conn, err := listener.AcceptKCP()
			if err != nil {
				log.Println("listen error", err)
			} else {
				log.Println("accept", conn.RemoteAddr())
				conn.SetNoDelay(1, 30, 2, 1)
				conn.SetReadBuffer(4096 * 1024)
				conn.SetWriteBuffer(4096 * 1024)
				conn.SetWindowSize(1024, 1024)
				conn.SetACKNoDelay(true)
				register(device, conn)
			}
		}
	}()

}

func register(device *tun.Device, conn net.Conn) {
	go func() {
		buf := make([]byte, 4096)
		addr := uint32(0)
		for true {
			rc, err := conn.Read(buf)
			if err != nil {
				log.Println(conn.RemoteAddr(), "read error", err)
				break
			}
			if addr == 0 {
				ipp := tun.IPPacket(buf[:rc])
				if err := ipp.Validate(); err != nil {
					log.Println(conn.RemoteAddr(), "ip packet error")
					break
				}
				sip := ipp.SrcIP()
				addr = convertIP(sip)
				_, ok := conns[addr]
				if !ok {
					log.Println("allow source ip", sip.String())
					conns[addr] = conn
				} else {
					log.Println("reject source ip", sip.String())
					conn.Close()
					break
				}
			}
			wc, err := device.Write(buf[:rc])
			if err != nil {
				log.Println(conn.RemoteAddr(), "write error", err)
				break
			}
			if rc != wc {
				log.Println(conn.RemoteAddr(), "broken connection", "read count:", rc, "write count:", wc)
				break
			}
		}
	}()
}

func convertIP(ip net.IP) uint32 {
	return binary.LittleEndian.Uint32(ip.To4())
}
