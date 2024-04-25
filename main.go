package main

import (
	"fmt"
	"log"

	"github.com/sd/fs/p2p"
)

func onPeer(peer p2p.Peer) error {
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    "localhost.trifacta.net:3000",
		HandshakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       &p2p.NOPDecoder{},
		OnPeer:        onPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("message : %+v\n", msg)
		}
	}()
	select {}
}
