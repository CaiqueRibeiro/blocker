package types

import (
	"crypto/sha256"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

// Returns a SHA256 of the block's header to be used as block's hash
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

// Gets the header of a block and hashes it in a sha256 [32]byte
func HashHeader(h *proto.Header) []byte {
	b, err := pb.Marshal(h)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

/*
Assigns the block after getting its hash
 1. transform the block header in a hash
 2. Sign (encrypt) the hash with private key
 3. Return the signed hash that only can be decrypted using public key
*/
func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	/*
		1. Transform de block header in []byte
		2. Sign it with private key
	*/
	hb := HashBlock(b) // block hashed in a [32]byte
	return pk.Sign(hb) // returns a signature (64 bytes)
}
