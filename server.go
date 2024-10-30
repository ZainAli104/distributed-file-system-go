package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ZainAli104/distributed-file-system-go/p2p"
	"io"
	"log"
	"sync"
)

type FileServerOpts struct {
	StorageRoot string
	PathTransformFunc
	Transport      p2p.Transport
	BootstrapNodes []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store this file to disk
	// 2. Broadcast this file to all the peers

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			log.Println("Failed to send message to peer: ", err)
			return err
		}
	}

	//time.Sleep(time.Second * 1)
	//
	//payload := []byte("THIS IS A LARGE FILE")
	//for _, peer := range s.peers {
	//	if err := peer.Send(payload); err != nil {
	//		log.Println("Failed to send message to peer: ", err)
	//	}
	//}

	return nil

	//buf := new(bytes.Buffer)
	//tee := io.TeeReader(r, buf)
	//
	//if err := s.store.Write(key, tee); err != nil {
	//	return err
	//}
	//
	//p := &DataMessage{
	//	Key:  key,
	//	Data: buf.Bytes(),
	//}
	//
	//msg := &Message{
	//	From:    "todo",
	//	Payload: p,
	//}
	//
	//return s.broadcast(msg)
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Println("New peer connected: ", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped due to user quit action.")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("Failed to decode the payload: ", err)
			}

			fmt.Printf("Payload: %+v\n", msg.Payload)

			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("peer not found")
			}

			b := make([]byte, 1000)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}
			fmt.Printf("1 %s\n", string(b))

			peer.(*p2p.TCPPeer).Wg.Done()

			//if err := s.handleMessage(&m); err != nil {
			//	log.Println("Failed to handle the message: ", err)
			//}

		case <-s.quitch:
			return
		}
	}
}

//func (s *FileServer) handleMessage(msg *Message) error {
//	switch p := msg.Payload.(type) {
//	case *Message:
//		fmt.Println("Received data message: ", p)
//	}
//
//	return nil
//}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Println("Attempt to connect with remote: ", addr)

			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Failed to dial to bootstrap node: ", addr)
				log.Println("Error: ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if err := s.bootstrapNetwork(); err != nil {
		return err
	}

	s.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}
