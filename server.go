package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"io"

	"github.com/sd/fs/p2p"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		store: NewStore(StoreOpts{
			Root:              opts.StorageRoot,
			PathTransformFunc: CASPathTransformFunc,
		}),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	From    net.Addr
	Payload any
}

type MessageStoreFile struct {
	Key string
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store file on disk
	// 2. broadcast this file to all known peers

	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(MessageStoreFile{
		Key: key,
	}); err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)
	payload := []byte("This large file")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := s.Store(key, tee); err != nil {
	// 	return err
	// }

	// _, err := io.Copy(buf, r)
	// if err != nil {
	// 	return err
	// }

	// dm := &Message{
	// 	Payload: DataMessage{
	// 		Key:  key,
	// 		Data: buf.Bytes(),
	// 	},
	// }

	// return s.broadcast(dm)
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) onPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s and peer %+v", p.RemoteAddr(), p)
	return nil
}

func (s *FileServer) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(msg.From.String(), v)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("receievd message store file %+v \n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in list", from)
	}

	if err := s.store.Write(msg.Key, peer); err != nil {
		return nil
	}

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			fmt.Println("received msesag", rpc)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}
			msg.From = rpc.From

			fmt.Printf("%+v\n", msg.Payload)

			if err := s.handleMessage(&msg); err != nil {
				log.Println(err)
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error :", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNodes) != 0 {
		s.bootstrapNetwork()
	}

	s.loop()
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	return s.store.Write(key, r)
}
