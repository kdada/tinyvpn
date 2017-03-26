package main

import (
	"log"

	"github.com/kdada/tinyvpn/cmd/client/stage"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_, err := stage.Start("45.32.34.206:10999", "admin", "123456")
	if err != nil {
		log.Println(err)
		return
	}
	select {}
}
