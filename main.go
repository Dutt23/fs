package main

import (
	"log"
	"time"

	"github.com/sd/fs/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    "localhost.trifacta.net:3000",
		HandshakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       &p2p.NOPDecoder{},
		// OnPeer:        onPeer,
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	sOpts := FileServerOpts{
		ListenAddr:        "localhost.trifacta.net:3000",
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}

	s := NewFileServer(sOpts)

	go func() {
		time.Sleep(time.Second * 3)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
	// select {}
}
