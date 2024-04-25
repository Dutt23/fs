package p2p

import (
	"fmt"
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

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("TCP handshake error: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

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
		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("TCP Error: %s\n", err)
			continue
			// lenDecoderError++
			// if lenDecoderError == 5 {

			// }
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
		// msg := buff[:n]
	}
}
