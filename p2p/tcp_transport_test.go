package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcpTransport(t *testing.T) {
	listenAddrr := ":4000"
	tr := NewTCPTransport(listenAddrr)

	assert.Equal(t, tr.listenAddress, listenAddrr)

	// Server

	assert.Nil(t, tr.ListenAndAccept())
}
