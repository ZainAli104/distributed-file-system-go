package main

import (
	"github.com/ZainAli104/distributed-file-system-go/p2p"
	"log"
)

func main() {
	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}

// TODO: Add Readme.md
