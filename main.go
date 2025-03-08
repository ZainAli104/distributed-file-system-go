package main

import (
	"bytes"
	"fmt"
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
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(1 * time.Second)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(1 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		//key := "personalPicture.jpg"
		key := fmt.Sprintf("picture_%d.jpg", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		err := s3.Store(key, data)
		if err != nil {
			log.Println(err)
			return
		}

		err = s3.store.Delete(s3.ID, key)
		if err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(string(b))
	}
}
