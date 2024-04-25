package p2p

// Peer is an interface that represents the remote node
type Peer interface {
	Close() error
}

// Transport is anything that can handles the communication
// between the nodes in the network. this can be of the form (TCP, UDP, websocket, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
