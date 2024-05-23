package util

import (
	randc "crypto/rand"
	"io"
	"math/rand"
	"time"

	"github.com/CaiqueRibeiro/blocker/proto"
)

// This utils functions will be used to generate random data for tests only

// Generates a random hash to be used in tests
func RandomHash() []byte {
	hash := make([]byte, 32)
	io.ReadFull(randc.Reader, hash)
	return hash
}

// Generates a random block to be used in tests
func RandomBlock() *proto.Block {
	header := &proto.Header{
		Version:   1,
		Height:    int32(rand.Intn(1000)),
		PrevHash:  RandomHash(),
		RootHash:  RandomHash(),
		Timestamp: time.Now().UnixNano(),
	}
	return &proto.Block{
		Header: header,
	}
}
