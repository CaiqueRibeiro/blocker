package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/CaiqueRibeiro/blocker/types"
)

const seed = "33c3e6749d95d5e9611c3f8e6ebcfe10d840226c46c4df18b7026b64be73a13f"

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

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	blockStore BlockStorer
	utxoStore  UTXOStorer
	headers    HeaderList
}

func NewChain(bs BlockStorer, txs TXStorer) *Chain {
	chain := &Chain{
		txStore:    txs,
		blockStore: bs,
		utxoStore:  NewMemoryUTXOStore(),
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
	for _, tx := range b.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}
		hash := hex.EncodeToString(types.HashTransaction(tx))
		for it, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				Amount:   output.Amount,
				OutIndex: it,
				Spent:    false,
			}
			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}
		for _, input := range tx.Inputs {
			key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PrevTxHash), input.PrevOutIndex)
			utxo, err := c.utxoStore.Get(key)
			if err != nil {
				return err
			}
			utxo.Spent = true
			if err = c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}
	}
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

	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid transaction signature")
	}
	// Check if all the outputs are unspent
	var (
		nInputs = len(tx.Inputs)
		hash    = hex.EncodeToString(types.HashTransaction(tx))
	)
	sumInputs := 0
	for i := 0; i < nInputs; i++ {
		prevHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
		key := fmt.Sprintf("%s_%d", prevHash, i)
		utxo, err := c.utxoStore.Get(key)
		sumInputs += int(utxo.Amount)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("input %d of tx %s is already spent", i, hash)
		}
	}
	sumOutputs := 0
	for _, output := range tx.Outputs {
		sumOutputs += int(output.Amount)
	}
	if sumInputs < sumOutputs {
		return fmt.Errorf("insufficient balance got (%d) spending (%d)", sumInputs, sumOutputs)
	}
	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromString(seed) // creates a private key by hand because it's a genesis block
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	types.SignBlock(privKey, block)

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	return block
}
