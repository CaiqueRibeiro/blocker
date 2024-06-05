package types

import (
	"crypto/sha256"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

// Returns a SHA256 of the block's header to be used as block's hash (general, not using priv key)
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

// Gets the header of a block and hashes it in a sha256 [32]byte (general, not using priv key)
func HashHeader(h *proto.Header) []byte {
	b, err := pb.Marshal(h)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

/*
Assigns the block after getting its hash and add private key and signature in it
 1. transform the block header in a hash
 2. Sign (encrypt) the hash with private key
 3. Add the public key in the block
 4. Add the signature in the block
 5. Return the signed hash that only can be decrypted using public key
*/
func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	/*
		1. Transform de block header in []byte
		2. Sign it with private key
	*/
	hash := HashBlock(b)              // block hashed in a [32]byte
	sig := pk.Sign(hash)              // returns a signature (64 bytes)
	b.PublicKey = pk.Public().Bytes() // public key used to unsign the pk.Sign(hash)
	b.Signature = sig.Bytes()         // hashed block signed with private key
	return sig
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)    // gets signature of the block (set in bytes)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey) // gets public key of the block (set in bytes)
	hash := HashBlock(b)                             // hash the block
	return sig.Verify(pubKey, hash)                  // verify if, when the sig.value is decrypted, it will be equal to hash
}
