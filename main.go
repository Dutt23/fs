package main

import (
	"log"

	"github.com/sd/fs/p2p"
)

func main() {
	tr := p2p.NewTCPTransport("localhost.trifacta.net:3004")

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
