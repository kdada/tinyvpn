package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/kdada/tinyvpn/pkg/tun"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate)
	srcIP := net.ParseIP("10.0.0.2")
	destIP := net.ParseIP("10.0.0.1")
	device, err := tun.CreateDevice(srcIP, destIP)
	if err != nil {
		panic(err)
	}
	defer device.Close()

	// add route
	_, x, _ := net.ParseCIDR("10.0.0.0/24")
	device.AddRoute(x)

	log.Println("start listen tunnel")
	go Output(device)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	sig := <-ch
	log.Println("exit because of", sig.String())
}

func Output(device *tun.Device) {
	data := make([]byte, 4096)
	ipp := tun.IPPacket(data)
	for true {
		count, err := device.Read(data)
		if err != nil {
			log.Println("read error:", err)
			return
		}
		log.Println("data length:", count)
		log.Print("src ip: ")
		log.Println(ipp.SrcIP())
		log.Print("dest ip: ")
		log.Println(ipp.DestIP())
	}
}
