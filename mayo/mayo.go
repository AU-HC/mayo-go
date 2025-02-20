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

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead return an error, if
// it fails to generate random bytes.
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

	v := mayo.n - mayo.o
	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	p1 := decodeMatrixList(mayo.m, v, v, p[:mayo.p1Bytes], true)
	p2 := decodeMatrixList(mayo.m, v, mayo.o, p[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)

	p3 := make([][][]byte, mayo.m)
	for i := 0; i < mayo.m; i++ {
		p3[i] = multiplyMatrices(transposeMatrix(o), addMatrices(multiplyMatrices(p1[i], o), p2[i]))
	}

	var cpk []byte
	cpk = append(cpk, seedPk...)
	cpk = append(cpk, encodeMatrixList(mayo.o, mayo.o, p3, true)...)
	csk := seedSk

	return cpk, csk, nil
}

func (mayo *Mayo) ExpandSK(csk []byte) []byte {
	// Parse csk
	seedSk := csk[:mayo.skSeedBytes]

	// Derive seedPk and O from seedSk
	s := make([]byte, mayo.pkSeedBytes+mayo.oBytes)
	h := sha3.NewSHAKE256()
	_, _ = h.Write(seedSk[:mayo.pkSeedBytes])
	_, _ = h.Read(s[:])
	seedPk := s[:mayo.pkSeedBytes]
	oByteString := s[mayo.pkSeedBytes : mayo.pkSeedBytes+mayo.oBytes]
	o := decodeMatrix(mayo.n-mayo.o, mayo.o, oByteString)

	// Derive P1 and P2 from seedPk
	v := mayo.n - mayo.o
	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	p1 := decodeMatrixList(mayo.m, v, v, p[:mayo.p1Bytes], true)
	p2 := decodeMatrixList(mayo.m, v, mayo.o, p[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)

	// Compute the L
	l := make([][][]byte, mayo.m)
	for i := 0; i < mayo.m; i++ {
		l[i] = addMatrices(multiplyMatrices(addMatrices(p1[i], transposeMatrix(p1[i])), o), p2[i])
	}

	// Encode L and output esk
	var esk []byte
	esk = append(esk, seedSk...)
	esk = append(esk, oByteString...)
	esk = append(esk, p[:mayo.p1Bytes]...)
	esk = append(esk, encodeMatrixList(v, mayo.o, l, false)...)

	return esk
}

func (mayo *Mayo) Sign() {

}

func (mayo *Mayo) Verify() {

}
