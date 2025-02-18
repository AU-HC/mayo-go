package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha3"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Mayo TODO: fix this struct, also refactor
type Mayo struct {
	// MAYO is parameterized by the following (missing F, which is the polynomial)
	q, m, n, o, k, saltBytes, digestBytes, pkSeedBytes int
	// MAYO then has the following derived parameters (missing E, which is a matrix)
	skSeedBytes, oBytes, vBytes, p1Bytes, p2Bytes, p3Bytes, lBytes, cskBytes, eskBytes, cpkBytes, epkBytes, sigBytes int
}

func InitMayo(q, m, n, o, k, saltBytes, digestBytes, pkSeedBytes int) *Mayo {
	if q != 16 {
		panic("q is fixed to be 16, in this version of MAYO")
	} else if k >= n-o {
		panic("k should be smaller than n-o")
	}

	skSeedBytes := saltBytes
	oBytes := int(math.Ceil(float64((n-o)*o) / 2.0))
	vBytes := int(math.Ceil(float64(n-o) / 2.0))
	p1Bytes := m * ((n - o) * ((n - o) + 1) / 2) / 2
	p2Bytes := m * (n - o) * o / 2
	p3Bytes := m * ((o + 1) * o / 2) / 2 // TODO: is this correct?
	lBytes := m * (n - o) * o / 2
	eskBytes := skSeedBytes + oBytes + p1Bytes + lBytes
	cpkBytes := pkSeedBytes + p3Bytes
	epkBytes := p1Bytes + p2Bytes + p3Bytes
	sigBytes := int(math.Ceil(float64(n*k)/2.0)) + saltBytes

	return &Mayo{
		q:           q,
		m:           m,
		n:           n,
		o:           o,
		k:           k,
		saltBytes:   saltBytes,
		digestBytes: digestBytes,
		pkSeedBytes: pkSeedBytes,
		// derived parameters
		skSeedBytes: skSeedBytes,
		oBytes:      oBytes,
		vBytes:      vBytes,
		p1Bytes:     p1Bytes,
		p2Bytes:     p2Bytes,
		p3Bytes:     p3Bytes,
		lBytes:      lBytes,
		cskBytes:    skSeedBytes,
		eskBytes:    eskBytes,
		cpkBytes:    cpkBytes,
		epkBytes:    epkBytes,
		sigBytes:    sigBytes,
	}
}

type CompactPublicKey struct {
}

type CompactSecretKey struct {
}

func decodeVec(n int, bytes []byte) []byte {
	decoded := make([]byte, n)
	var i int
	for i = 0; i < n/2; i++ {
		decoded[i*2] = bytes[i] & 0xf // 0xf=00001111
		decoded[i*2+1] = bytes[i] >> 4
	}

	// if 'n' is odd, then fix last nibble
	if n%2 == 1 {
		decoded[i*2] = bytes[i] & 0xf // 0xf=00001111
	}

	return decoded
}

func encodeVec(bytes []byte) []byte {
	encoded := make([]byte, int(math.Ceil(float64(len(bytes))/2.0)))

	var i int
	for i = 0; i+1 < len(bytes); i += 2 {
		encoded[i/2] = (bytes[i+0] << 0) | (bytes[i+1] << 4)
	}

	// if 'n' is odd, then fix last nibble
	if len(bytes)%2 == 1 {
		encoded[i/2] = bytes[i+0] << 0
	}

	return encoded
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
