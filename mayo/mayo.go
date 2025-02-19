package mayo

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha3"
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

	p := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)

	p1 := toInt64(p[:mayo.p1Bytes])
	p2 := toInt64(p[mayo.p1Bytes : mayo.p1Bytes+mayo.p2Bytes])

	expectedp1Size := ((mayo.n - mayo.o) * (mayo.n - mayo.o) * mayo.m) / 8
	p1Size := len(p1)

	fmt.Println("Expectedp1Size: ", expectedp1Size)
	fmt.Println("Actual size: ", p1Size)

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

func (mayo *Mayo) Sign() {

}

func (mayo *Mayo) Verify() {

}
