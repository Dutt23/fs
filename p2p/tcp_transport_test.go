package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcpTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddr:    "localhost.trifacta.net:3004",
		HandshakeFunc: NOPHandShakeFunc,
		Decoder:       &NOPDecoder{},
	}
	tr := NewTCPTransport(tcpOpts)

	assert.Equal(t, tcpOpts.ListenAddr, "localhost.trifacta.net:3004")

	// Server

	assert.Nil(t, tr.ListenAndAccept())
}
