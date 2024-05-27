package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/CaiqueRibeiro/blocker/node"
	"github.com/CaiqueRibeiro/blocker/proto"
	"google.golang.org/grpc"
)

func main() {
	node := node.NewNode()
	options := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(options...)
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	proto.RegisterNodeServer(grpcServer, node)
	fmt.Println("node running on port:", ":3000")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			makeTransaction()
		}
	}()

	grpcServer.Serve(ln)
}

// temporary: just to test gRPC calls
func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(client)
	version := &proto.Version{
		Version: "blocker-0.1",
		Height:  1,
	}

	_, err = c.Handshake(context.TODO(), version)
	if err != nil {
		log.Fatal(err)
	}
}
