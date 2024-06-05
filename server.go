package main

import (
	"bytes"
	"encoding/binary"
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
	Key  string
	Size int64
}

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("[%s] serving file from (%s) local disk \n", s.ListenAddr, key)
		_, f, err := s.store.Read(key)
		return f, err
	}

	msg := &Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	fmt.Printf("don't have file (%s)locally , fetching from network ...\n", key)

	if err := s.broadcast(msg); err != nil {
		return nil, nil
	}

	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received bytes (%d) over the network : (%s)", s.ListenAddr, n, peer.LocalAddr().String())

		peer.CloseStream()
	}
	_, f, err := s.store.Read(key)

	return f, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store file on disk
	// 2. broadcast this file to all known peers

	buffer := new(bytes.Buffer)
	tee := io.TeeReader(r, buffer)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return nil
	}

	msg := &Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 3)

	// TODO: multiwriter
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		_, err := io.Copy(peer, buffer)
		if err != nil {
			return nil
		}
	}

	return nil
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
	case MessageGetFile:
		return s.handleMessageGetFile(msg.From.String(), v)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Println("need to get a file from disk and send it over the wire")
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("file (%s) does not exist on disk\n", msg.Key)
	}

	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not in map")
	}

	// Send incoming byte message
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)

	if err != nil {
		return err
	}

	fmt.Printf("[%s] written bytes (%d) over the network to %s\n", s.ListenAddr, n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in list", from)
	}

	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size))); err != nil {
		return nil
	}

	peer.CloseStream()

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
			fmt.Println("received msesage", rpc)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding err : ", err)
			}
			msg.From = rpc.From

			fmt.Printf("%+v\n", msg.Payload)

			if err := s.handleMessage(&msg); err != nil {
				log.Println("handling message error : ", err)
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

// func (s *FileServer) Store(key string, r io.Reader) (int64, error) {
// 	return s.store.Write(key, r)
// }
