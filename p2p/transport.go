package p2p

import "net"

// Peer is the interface that represents the remote node.
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport is anything that can handles the communication
// between the nodes in the internet. This can be of the
// form (TCP, UDP, websockets, etc.)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
