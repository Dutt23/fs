package p2p

import "errors"

// ErrInvalidHandshake is returned if the handshake between
// the local and remote node could not be established
var ErrInvalidHandshake = errors.New("invalid handshake")

// HandshakeFunc is
type HandshakeFunc func(Peer) error

func NOPHandShakeFunc(Peer) error {
	return nil
}
