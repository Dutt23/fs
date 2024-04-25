package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan *RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// TCPPeer represents the remote node over a tcp established connection
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn,
		outbound,
	}
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan *RPC),
	}
}

// Consume implements the transport interface, which will return a read-only channel.
func (t *TCPTransport) Consume() <-chan *RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	log.Println("listening on : ", t.ListenAddr)
	t.listener = ln

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP Transport error %v", err)
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) Close() error {
	close(t.rpcch)
	return t.listener.Close()
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("TCP handshake error: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	fmt.Printf("new incoming connection %+v\n", peer)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return
		}
	}

	// lenDecoderError := 0
	// Read Loop
	rpc := &RPC{}
	for {
		err := t.Decoder.Decode(conn, rpc)
		if err == net.ErrClosed {
			fmt.Printf("dropping con %s\n", err)
			return
		}
		if err != nil {
			fmt.Printf("TCP Error: %s\n", err)
			return
			// lenDecoderError++
			// if lenDecoderError == 5 {

			// }
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
		// msg := buff[:n]
	}
}
