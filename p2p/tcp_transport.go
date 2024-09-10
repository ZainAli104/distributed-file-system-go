package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// conn is the underlying connection to the remote node.
	conn net.Conn
	// if we dail and retrieve a conn => outbound = true
	// if we accept and retrieve a conn => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn,
		outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("Error TCP accepting connection: %v\n", err)
		}
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(conn); err != nil {
		_ = conn.Close()
		fmt.Printf("Error TCP shaking hands: %v\n", err)
		return
	}

	// Read loop
	msg := &Temp{}
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("Error TCP decoding message: %v\n", err)
			continue
		}

		fmt.Printf("Received message: %v\n", msg)
	}

	//buf := new(bytes.Buffer)
	//for {
	//	data := make([]byte, 1024)
	//	n, err := conn.Read(data)
	//	if err != nil {
	//		fmt.Printf("Error TCP reading data: %v\n", err)
	//		return
	//	}
	//
	//	buf.Write(data[:n])
	//
	//	if n < 1024 {
	//		break
	//	}
	//}

	fmt.Printf("New incoming connection from %v\n", peer)
}
