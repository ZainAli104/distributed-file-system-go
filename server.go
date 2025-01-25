package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/ZainAli104/distributed-file-system-go/p2p"
	"io"
	"log"
	"sync"
	"time"
)

type FileServerOpts struct {
	EncKey      []byte
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

func (s *FileServer) stream(msg *Message) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			log.Println("Failed to send message to peer: ", err)
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("%s serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)
		return r, err
	}

	fmt.Printf("%s File not found (%s) locally, fetching from the network\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		// First read the file size so we can limit the amount of bytes that we read
		// from the connection, so it will not keep hanging.
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		// Store the file to disk and then return the reader.
		n, err := s.store.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_, r, err := s.store.Read(key)
	return r, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
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
			Size: size + 16,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)

	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	mw.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

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
		log.Println("File server stopped due to error or user quit action.")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("[%s] file (%s) does not exist on disk\n", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		fmt.Println("Closing the file reader")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not found: %s", from)
	}

	// First send the "incomingStream" byte to the peer, and then we can send
	// the file size as an int64.
	peer.Send([]byte{p2p.IncomingStream})
	//var fileSize int64 = 22
	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	log.Printf("[%s] written %d bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("Received file store message: %+v\n", msg.Key)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not found: %s", from)
	}

	defer peer.CloseStream()

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	log.Printf("%s written %d bytes to disk", s.Transport.Addr(), n)
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
	gob.Register(MessageGetFile{})
}
