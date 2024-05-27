package types

import (
	"testing"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/CaiqueRibeiro/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestNewtransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.Public().Address().Bytes()

	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivKey.Public().Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}

	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	sig := SignTransaction(fromPrivKey, tx) // hashes all the message (with input and output)
	input.Signature = sig.Bytes()           // uses the whole tx hashed as signature of input

	assert.True(t, VerifyTransaction(tx)) // verify if transaction was signed properly in flow
}
