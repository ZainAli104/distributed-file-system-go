package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {
	listenAdd := ":3000"

	tcpOpts := TCPTransportOpts{
		ListenAddr:    listenAdd,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       GOBDecoder{},
	}

	tr := NewTCPTransport(tcpOpts)
	assert.Equal(t, tr.ListenAddr, listenAdd)

	// Server
	assert.Nil(t, tr.ListenAndAccept())
}
