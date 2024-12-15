package main

import (
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

	//data := bytes.NewReader([]byte("my big data file here!"))
	//err := s2.Store("personalPicture.jpg", data)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//time.Sleep(time.Millisecond * 5)

	r, err := s2.Get("personalPicture.jpg")
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(b))
}
