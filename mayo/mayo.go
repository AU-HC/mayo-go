package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha3"
	"io"
)

func aes128ctr(seed []byte, l int) []byte {
	var nonce [16]byte
	block, _ := aes.NewCipher(seed[:])
	ctr := cipher.NewCTR(block, nonce[:])
	dst := make([]byte, l)
	ctr.XORKeyStream(dst[:], dst[:]) // TODO: should this just be all zeroes?
	return dst
}

func (mayo *Mayo) CompactKeyGen() ([]byte, []byte, error) {
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
	ot := decodeMatrix(mayo.o, mayo.n-mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes])

	v := mayo.n - mayo.o
	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	p1 := decodeMatrixList(mayo.m, v, v, p[:mayo.p1Bytes], true)
	p2 := decodeMatrixList(mayo.m, v, mayo.o, p[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)
	p3 := make([][][]byte, mayo.m)

	for i := 0; i < mayo.m; i++ {
		p3[i] = multiplyMatrices(ot, addMatrices(multiplyMatrices(p1[i], o), p2[i])) // TODO: Calculate ot properly
	}

	var cpk []byte
	cpk = append(cpk, seedPk...)
	cpk = append(cpk, encodeMatrixList(mayo.o, mayo.o, p3, true)...)
	csk := seedPk

	return cpk, csk, nil
}

func (mayo *Mayo) Sign() {

}

func (mayo *Mayo) Verify() {

}
