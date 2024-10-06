package main

import (
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

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(p Payload) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store this file to disk
	// 2. Broadcast this file to all the peers

	if err := s.store.Write(key, r); err != nil {
		return err
	}

	return nil
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
		case msg := <-s.Transport.Consume():
			fmt.Println("Received message: ", msg)
		case <-s.quitch:
			return
		}
	}
}

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
