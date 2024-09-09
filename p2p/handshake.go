package p2p

type Handshaker interface {
	Handshake() error
}
