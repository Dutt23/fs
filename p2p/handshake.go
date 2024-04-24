package p2p

// HandshakeFunc is
type HandshakeFunc func(Peer) error

func NOPHandShakeFunc(Peer) error {
	return nil
}
