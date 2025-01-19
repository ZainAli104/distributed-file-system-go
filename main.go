package main

import (
	"bytes"
	"github.com/ZainAli104/distributed-file-system-go/p2p"
	"io"
	"log"
	"strings"
	"time"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	sanitizedAddr := strings.Replace(listenAddr, ":", "", -1) // remove the colon
	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       sanitizedAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go s2.Start()
	time.Sleep(2 * time.Second)

	key := "personalPicture.jpg"
	data := bytes.NewReader([]byte("my big data file here!"))
	err := s2.Store(key, data)
	if err != nil {
		log.Println(err)
		return
	}

	err = s2.store.Delete(key)
	if err != nil {
		log.Fatal(err)
	}

	r, err := s2.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(b))
}
