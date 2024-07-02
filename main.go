package main

import (
	"context"
	"log"
	"time"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/node"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/CaiqueRibeiro/blocker/util"
	"google.golang.org/grpc"
)

func main() {
	makeNode(":3000", []string{}, true) // creates a genesis node
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}, false) // creates a node that connects to the genesis node
	time.Sleep(time.Second)
	makeNode(":6000", []string{":4000"}, false) // creates a node that connects to the genesis node

	for {
		time.Sleep(time.Second)
		makeTransaction()
	}
}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {
	cfg := node.ServerConfig{
		Version:    "Blocker-1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.NewNode(cfg)                 // creates a new node
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
	privKey := crypto.GeneratePrivateKey()
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	_, err = c.HandleTransaction(context.TODO(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
