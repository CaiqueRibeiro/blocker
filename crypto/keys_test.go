package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), privKeyLen)
	pubKey := privKey.Public()
	assert.Equal(t, len(pubKey.Bytes()), pubKeyLen)
}

func TestGeneratePrivateKeyFromString(t *testing.T) {
	var (
		seed         = "8e41a5878c3f70850588f6560c91048fa7d67743a148ddce23c1e47aeb149871"
		expectedAddr = "3579839bce98bc81030b0ab5068e155e55bf222b"
		privKey      = NewPrivateKeyFromString(seed)
	)
	assert.Equal(t, privKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	assert.Equal(t, expectedAddr, address.String())
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	msg := []byte("foo bar baz")

	sig := privKey.Sign(msg) // signed the message and encrypt it
	// the with valid message:
	assert.True(t, sig.Verify(pubKey, msg)) // use public key to verify if raw and signed msg are the same (true)\
	// the with invalid message:
	assert.False(t, sig.Verify(pubKey, []byte("foo"))) // use public key to verify if raw and signed msg are the same (false)

	// test with invalid pubKey
	otherPrivKey := GeneratePrivateKey()
	otherPubKey := otherPrivKey.Public()

	/*
		Tries to use another public key to unsign the "sig" signed message and compare it with the message
		But, as the sig.value was not signed using "otherPrivKey", using "otherPubKey" won't be able to unsign it
	*/
	assert.False(t, sig.Verify(otherPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()
	assert.Equal(t, addressLen, len(address.Bytes()))
}
