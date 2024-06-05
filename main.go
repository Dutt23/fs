package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
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

	// for i := 0; i < 1; i++ {
	// 	data := bytes.NewReader([]byte(fmt.Sprintf("my big data file ! (%d)", i)))
	// 	s2.Store(fmt.Sprintf("myprivatedata_%d", i), data)
	// 	time.Sleep(5 * time.Millisecond)
	// }
	// data := bytes.NewReader([]byte("my big data file !"))
	// err := s2.Store("myprivatedata", data)
	// if err != nil {
	// 	panic(err)
	// }

	r, err := s1.Get("myprivatedata_0")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
