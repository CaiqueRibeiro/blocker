package types

import (
	"testing"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	// verifies if generated block hash is [32]byte
	assert.Equal(t, 32, len(hash))
}

func TestSignBlock(t *testing.T) {
	var (
		block   = util.RandomBlock()          // creates a dummy block
		privKey = crypto.GeneratePrivateKey() // generates a new private key
		pubKey  = privKey.Public()            // get its public key
	)
	sig := SignBlock(privKey, block) // sign the block using private key
	// signed hash must be [64]byte
	assert.Equal(t, 64, len(sig.Bytes()))
	// verify if, decrypting hashed block, it will be equal to sign.value
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))
}
