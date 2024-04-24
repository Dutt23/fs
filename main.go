package main

import (
	"log"

	"github.com/sd/fs/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    "localhost.trifacta.net:3004",
		HandshakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       &p2p.NOPDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
