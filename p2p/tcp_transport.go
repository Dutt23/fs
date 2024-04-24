package p2p

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// TCPPeer represents the remote node over a tcp established connection
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn,
		outbound,
	}
}

func NewTCPTransport(listenAddress string) *TCPTransport {
	return &TCPTransport{
		shakeHands:    NOPHandShakeFunc,
		listenAddress: listenAddress,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	t.listener = ln

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP Transport error %v", err)
		}
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil {
	}

	buf := new(bytes.Buffer)

	// lenDecoderError := 0
	// Read Loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP Error: %s\n", err)
			continue
			// lenDecoderError++
			// if lenDecoderError == 5 {

			// }
		}
		// msg := buff[:n]
	}

	fmt.Printf("new incoming connection %+v\n", peer)
}
