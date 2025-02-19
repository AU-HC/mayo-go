package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha3"
	"encoding/binary"
	"fmt"
	"io"
)

type CompactPublicKey struct {
}

type CompactSecretKey struct {
}

func aes128ctr(seed []byte, l int) []byte {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:]) // TODO: should this just be all zeroes?
	return dst
}

func (mayo *Mayo) CompactKeyGen() (*CompactPublicKey, *CompactSecretKey, error) {
	seedSk := make([]byte, mayo.skSeedBytes)
	rand := cryptoRand.Reader // TODO: refactor this prob
	_, err := io.ReadFull(rand, seedSk[:])
	if err != nil {
		return nil, nil, err
	}

	s := make([]byte, mayo.pkSeedBytes+mayo.oBytes)
	h := sha3.NewSHAKE256()
	_, _ = h.Write(seedSk[:])
	_, _ = h.Read(s[:])

	seedPk := s[:mayo.pkSeedBytes]
	o := decodeVec((mayo.n-mayo.o)*mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes]) // TODO: fix this to be a matrix
	om := make([][]byte, mayo.n-mayo.o)
	for i := 0; i < len(om); i++ {
		om[i] = o[i*mayo.o : (i+1)*mayo.o]
	}

	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)

	p1 := slice(p[:mayo.p1Bytes])
	p2 := slice(p[mayo.p1Bytes : mayo.p1Bytes+mayo.p2Bytes])

	for i := 0; i < mayo.m; i++ {
		p1i := p1[i*mayo.p1Bytes : (i+1)*mayo.p1Bytes]
		p2i := p2[i*mayo.p2Bytes : (i+1)*mayo.p2Bytes]
		fmt.Println(p1i)
		fmt.Println(p2i)
	}

	fmt.Println(p1)
	fmt.Println(p2)

	return &CompactPublicKey{}, &CompactSecretKey{}, nil
}

func slice(src []byte) []uint64 {
	dst := make([]uint64, len(src)/8)

	for i := range dst {
		dst[i] = binary.LittleEndian.Uint64(src)
		src = src[8:]
	}

	fmt.Println(len(src))

	return dst
}

func (mayo *Mayo) Sign() {

}

func (mayo *Mayo) Verify() {

}
