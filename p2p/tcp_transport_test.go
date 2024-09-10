package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       GOBDecoder{},
	}

	tr := NewTCPTransport(tcpOpts)
	assert.Equal(t, ":3000", tr.ListenAddr)

	// Server
	assert.Nil(t, tr.ListenAndAccept())

	select {}
}
