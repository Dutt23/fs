package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
		StorageRoot:       fmt.Sprintf("%s_network", listenAddr),
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
	err := s2.Store("myprivatedata", data)
	if err != nil {
		panic(err)
	}

	// r, err := s2.Get("myprivatedata")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// b, err := ioutil.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))
	select {}
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
