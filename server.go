package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ZainAli104/distributed-file-system-go/p2p"
	"io"
	"log"
	"sync"
	"time"
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
	Key  string
	Size int64
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	// 1. Store this file to disk
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	// 2. Broadcast this file to all the peers
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			log.Println("Failed to send message to peer: ", err)
			return err
		}
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			log.Println("Failed to send message to peer: ", err)
		}

		log.Println("Sent ", n, " bytes to ", peer.RemoteAddr())
	}

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

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Failed to handle the message: ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch p := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, p)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("Received file store message: %+v\n", msg.Key)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not found: %s", from)
	}

	defer peer.(*p2p.TCPPeer).Wg.Done()

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	log.Printf("Wrote %d bytes to disk", n)
	return nil
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

func init() {
	gob.Register(MessageStoreFile{})
}
