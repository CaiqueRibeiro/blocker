package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/CaiqueRibeiro/blocker/types"
)

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: make([]*proto.Header, 0),
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		panic("index too high")
	}
	return list.headers[index]
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

type Chain struct {
	blockStore BlockStorer
	headers    HeaderList
}

func NewChain(bs BlockStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		headers:    *NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	if err := c.addBlock(b); err != nil {
		return err
	}
	return nil
}

// Add block with validation (to be used outside the chain scope)
func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)
	return c.blockStore.Put(b)
}

// Add block without validation (to be used internally, ex: creating genesis block)
func (c *Chain) GetBlockByHash(b []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(b)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if height > c.Height() {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

/*
Validates the incomin block to verify if it should be added to the chain
 1. Validates the signature of the block
 2. Validates if the previous hash of the block is equal to the hash of the last block in the chain
*/
func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validates the signature of the block
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalid block signature")
	}

	// validates if block to be validated (current) has the previous hash equal to hash of last chain block
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(hash, b.Header.PrevHash) {
		return fmt.Errorf("invalid previous block hash")
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.GeneratePrivateKey() // creates a private key by hand because it's a genesis block
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	types.SignBlock(privKey, block)
	return block
}
