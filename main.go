package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/sd/fs/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandShakeFunc,
		Decoder:       &p2p.NOPDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	sOpts := FileServerOpts{
		ListenAddr:        listenAddr,
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(sOpts)

	tcpTransport.OnPeer = s.onPeer
	return s

}

func main() {
	s1 := makeServer(":4000", "")
	s2 := makeServer(":5000", ":4000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)

	go s2.Start()
	time.Sleep(2 * time.Second)

	data := bytes.NewReader([]byte("my big data file !"))
	err := s2.StoreData("myprivatedata", data)
	if err != nil {
		panic(err)
	}
	select {}
}

func init() {
	gob.Register(MessageStoreFile{})
}
