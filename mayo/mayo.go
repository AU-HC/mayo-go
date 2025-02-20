package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha3"
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
	o := decodeMatrix(mayo.n-mayo.o, mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes])

	fmt.Println(o)

	v := mayo.n - mayo.o
	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	p1 := decodeMatrixList(mayo.m, v, v, p[:mayo.p1Bytes], true)
	p2 := decodeMatrixList(mayo.m, v, mayo.o, p[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)

	for i := 0; i < mayo.m; i++ {
		p1i := p1[i*mayo.p1Bytes : (i+1)*mayo.p1Bytes]
		p2i := p2[i*mayo.p2Bytes : (i+1)*mayo.p2Bytes]
		fmt.Println(p1i)
		fmt.Println(p2i)
	}

	return &CompactPublicKey{}, &CompactSecretKey{}, nil
}

func (mayo *Mayo) Sign() {

}

func (mayo *Mayo) Verify() {

}
