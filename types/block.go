package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/cbergoon/merkletree"
	pb "google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{
		hash: hash,
	}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

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
		1. Create Merkleroot if not exists
		2. Transform de block header in []byte
		3. Sign it with private key
	*/
	if len(b.Transactions) > 0 {
		tree, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}
		b.Header.RootHash = tree.MerkleRoot()
	}
	hash := HashBlock(b)              // block hashed in a [32]byte
	sig := pk.Sign(hash)              // returns a signature (64 bytes)
	b.PublicKey = pk.Public().Bytes() // public key used to unsign the pk.Sign(hash)
	b.Signature = sig.Bytes()         // hashed block signed with private key
	return sig
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.Transactions) > 0 {
		if !VerifyRootHash(b) {
			return false
		}
	}
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}
	var (
		sig    = crypto.SignatureFromBytes(b.Signature) // gets signature of the block (set in bytes)
		pubKey = crypto.PublicKeyFromBytes(b.PublicKey) // gets public key of the block (set in bytes)
		hash   = HashBlock(b)                           // hash the block
	)
	return sig.Verify(pubKey, hash) // verify if, when the sig.value is decrypted, it will be equal to hash
}

func VerifyRootHash(b *proto.Block) bool {
	tree, err := GetMerkleTree(b)
	if err != nil {
		return false
	}
	valid, err := tree.VerifyTree()
	if err != nil {
		return false
	}
	if !valid {
		return false
	}
	return bytes.Equal(b.Header.RootHash, tree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))
	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}

	// Create new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
}
