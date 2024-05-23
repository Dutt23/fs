package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// Message holds any arbitrary data that is being sent over
// each transport between two nodes
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
