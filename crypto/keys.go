package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	privKeyLen   = 64 // 32 from private key + 32 from appending public key
	signatureLen = 64
	pubKeyLen    = 32
	seedLen      = 32
	addressLen   = 20
)

// Private Key
func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != seedLen {
		panic("invalid seed length. Must be 32")
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromString(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

type PrivateKey struct {
	key ed25519.PrivateKey
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

/*
Uses the generated private key bytes to sign/encrypt the message.
Will only be decrypted using public key.
This will generate a hash that, using public key, can be reverted to original message
*/
func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, msg), // signed message
	}
}

/*
Returns the public key. Public key is extracted from last 32 bytes of private key
*/
func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, pubKeyLen)
	copy(b, p.key[pubKeyLen:]) // get last 32 bytes of private key

	return &PublicKey{
		key: b,
	}
}

// Public Key
type PublicKey struct {
	key ed25519.PublicKey
}

// Converts a public key in bytes to the proper struct (do not change value)
func PublicKeyFromBytes(b []byte) *PublicKey {
	if len(b) != pubKeyLen {
		panic("invalid public key length")
	}
	return &PublicKey{
		key: ed25519.PublicKey(b),
	}
}

func (p *PublicKey) Address() Address {
	return Address{
		value: p.key[len(p.key)-addressLen:], // same as p.key[12:]. Ignores first 12 bytes and get last 20 to be address
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

// Signature: a message hashed (signed) with the private key
type Signature struct {
	value []byte
}

func (s *Signature) Bytes() []byte {
	return s.value
}

// Converts a signature in bytes to the proper struct (do not change value)
func SignatureFromBytes(b []byte) *Signature {
	if len(b) != signatureLen {
		panic(fmt.Sprintf("length of the bytes not equal to %d", signatureLen))
	}
	return &Signature{
		value: b,
	}
}

/*
Verify if the the message informed is the same os the signature.
This can be verified using the public key + s.value
(if s.value can be descrypted and it's equal to msg, it means that msg informed is valid)
pubKey: the public key that should be used
msg: the message you want to compare with
*/
func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	/*
		pubKey.key: the public key to extract data from signed message (3rd param)
		msg: the raw msg that is to be verified
		s.value: the signed message that will be compared with raw message
	*/
	return ed25519.Verify(pubKey.key, msg, s.value)
}

// Address
type Address struct {
	value []byte
}

func (a Address) Bytes() []byte {
	return a.value
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}
