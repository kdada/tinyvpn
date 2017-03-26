package main

import (
	"log"

	"github.com/kdada/tinyvpn/cmd/server/stage"
	"github.com/kdada/tinyvpn/pkg/proto"
	"github.com/kdada/tinyvpn/pkg/proto/types"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	server, err := proto.NewServer(":10999", types.DefaultConverter)
	if err != nil {
		log.Println(err)
		return
	}
	server.Run(&proto.ServerHandler{
		Accepted: accepted,
		Closed:   closed,
	})
	select {}
}

func accepted(conn *proto.Connection) {
	_, err := stage.Start(conn)
	if err != nil {
		log.Println(err)
		conn.Close(err.Error())
	}
}

func closed(conn *proto.Connection, reason string) {
	log.Println("closed", conn.Session.RemoteAddr().String(), reason)
}
