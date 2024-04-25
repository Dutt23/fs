package main

import (
	"log"

	"github.com/sd/fs/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       &p2p.NOPDecoder{},
		// OnPeer:        onPeer,
	}

	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	sOpts := FileServerOpts{
		ListenAddr:        listenAddr,
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	return NewFileServer(sOpts)

}

func main() {
	s1 := makeServer(":4000", "")
	s2 := makeServer(":5000", ":4000")
	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()
}
