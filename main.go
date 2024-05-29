package main

import (
	"context"
	"log"
	"time"

	"github.com/CaiqueRibeiro/blocker/node"
	"github.com/CaiqueRibeiro/blocker/proto"
	"google.golang.org/grpc"
)

func main() {
	makeNode(":3000", []string{}) // creates a genesis node
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}) // creates a node that connects to the genesis node
	time.Sleep(4 * time.Second)
	makeNode(":6000", []string{":4000"})

	select {} // just to block
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode()                    // creates a new node
	go n.Start(listenAddr, bootstrapNodes) // starts the node
	return n
}

// temporary: just to test gRPC calls
func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(client)
	version := &proto.Version{
		Version:    "blocker-0.1",
		Height:     1,
		ListenAddr: ":4000",
	}

	/* sends own version to another node an receives its version from it */
	_, err = c.Handshake(context.TODO(), version)
	if err != nil {
		log.Fatal(err)
	}
}
