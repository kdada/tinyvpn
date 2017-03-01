package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"flag"

	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/tun"
	"github.com/xtaci/kcp-go"
)

var server string
var local string
var remote string
var route string

func init() {
	flag.StringVar(&server, "s", "", "host:port e.g. 22.22.22.22:9989")
	flag.StringVar(&local, "l", "", "local ip e.g. 10.0.0.2")
	flag.StringVar(&remote, "r", "", "remote ip e.g. 10.0.0.1")
	flag.StringVar(&route, "d", "", "default route e.g. 10.0.0.0/24")
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ldate)
	log.Println("tinyvpn client started")
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

	// connect
	conn, err := kcp.DialWithOptions(server, nil, 10, 3)
	conn.SetNoDelay(1, 30, 2, 1)
	conn.SetReadBuffer(4096 * 1024)
	conn.SetWriteBuffer(4096 * 1024)
	conn.SetWindowSize(1024, 1024)
	conn.SetACKNoDelay(true)
	if err != nil {
		panic(err)
	}
	log.Println("tunnel connected")

	running := true

	sender := proto.Pipe("sender", &running, device, conn)
	receiver := proto.Pipe("receiver", &running, conn, device)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	select {
	case <-sender:
		log.Println("exit because of sender failed")
	case <-receiver:
		log.Println("exit because of receiver failed")
	case <-sig:
		log.Println("exit because of user")
	}
	log.Println("tinyvpn client stoped")
}
