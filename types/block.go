package types

import (
	"crypto/sha256"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

// Returns a SHA256 of the block to be used as block's hash
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

/*
Assigns the block after getting its hash
 1. transform the block in a hash
 2. Sign (encrypt) the hash with private key
 3. Return the signed hash that only can be decrypted using public key
*/
func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	/*
		1. Transform de block in []byte
		2. Sign it with private key
	*/
	hb := HashBlock(b) // block hashed in a [32]byte
	return pk.Sign(hb) // returns a signature (64 bytes)
}
