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

	key := newEncryptionKey()
	fmt.Printf("key (%+v) \n", key)
	sOpts := FileServerOpts{
		EncKey:            key,
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
	data := bytes.NewReader([]byte("my big data file !"))
	err := s2.Store("myprivatedata", data)
	if err != nil {
		panic(err)
	}

	err = s2.store.Delete("myprivatedata")
	if err != nil {
		log.Fatal(err)
	}

	_, err = s2.Get("myprivatedata")
	if err != nil {
		log.Fatal("error here %+v", err)
	}

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
