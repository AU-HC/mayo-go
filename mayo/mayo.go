package mayo

import (
	"bytes"
	"math"
	"mayo-go/rand"
)

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead
// return an error, if it fails to generate random bytes.
func (mayo *Mayo) CompactKeyGen() ([]byte, []byte, error) {
	// Pick seekSk at random
	seedSk := rand.SampleRandomBytes(mayo.skSeedBytes)

	// Derive seedPk and O from seekSk
	s := rand.SHAKE256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
	seedPk := s[:mayo.pkSeedBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes])

	// Derive P_i^1 and P_i^2 from seekPk
	P := rand.AES128CTR64(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	P1 := P[:mayo.p1Bytes/8]                                // v x v upper triangular matrix
	P2 := P[mayo.p1Bytes/8 : (mayo.p1Bytes+mayo.p2Bytes)/8] // v x o matrix

	// Compute P3
	P3Bytes := mayo.computeP3(P1, O, P2)

	// Encode the compact public/secret key
	cpk := make([]byte, mayo.cpkBytes)
	copy(cpk[:mayo.pkSeedBytes], seedPk)
	copy(cpk[mayo.pkSeedBytes:], P3Bytes)
	csk := seedSk

	// Output keys
	return cpk, csk, nil
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk []byte) []byte {
	// Parse csk
	seedSk := csk[:mayo.skSeedBytes]

	// Derive seedPk and O from seedSk
	S := rand.SHAKE256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
	seedPk := S[:mayo.pkSeedBytes]
	oByteString := S[mayo.pkSeedBytes : mayo.pkSeedBytes+mayo.oBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, oByteString)

	// Derive P1 and P2 from seedPk
	P := rand.AES128CTR64(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	P1 := P[:mayo.p1Bytes/8]                                // v x v upper triangular matrix
	P2 := P[mayo.p1Bytes/8 : (mayo.p1Bytes+mayo.p2Bytes)/8] // v x o matrix

	// Compute L
	lBytes := mayo.computeL(P1, O, P2)

	// Encode the SK and output esk
	p1Bytes := make([]byte, mayo.p1Bytes)
	uint64SliceToBytes(p1Bytes, P1)

	esk := make([]byte, mayo.eskBytes)
	copy(esk[:mayo.skSeedBytes], seedSk)
	copy(esk[mayo.skSeedBytes:], oByteString)
	copy(esk[mayo.skSeedBytes+mayo.oBytes:], p1Bytes)
	copy(esk[mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes:], lBytes)
	return esk
}

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk []byte) []byte {
	// Parse cpk
	seedPk := cpk[:mayo.pkSeedBytes]

	// Expand seedPk and return epk
	epk := make([]byte, mayo.epkBytes)
	copy(epk[:mayo.p1Bytes+mayo.p2Bytes], rand.AES128CTR(seedPk, mayo.p1Bytes+mayo.p2Bytes))
	copy(epk[mayo.p1Bytes+mayo.p2Bytes:], cpk[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.p3Bytes])
	return epk
}

// Sign (Algorithm 7) takes an expanded secret key esk and a message m and outputs a signature on the message m
func (mayo *Mayo) Sign(esk, m []byte) []byte {
	// Decode esk
	seedSk := esk[:mayo.skSeedBytes]
	O := decodeVec(mayo.v*mayo.o, esk[mayo.skSeedBytes:mayo.skSeedBytes+mayo.oBytes])
	P1Bytes := esk[mayo.skSeedBytes+mayo.oBytes : mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes]
	LBytes := esk[mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes : mayo.eskBytes]
	P1 := make([]uint64, mayo.p1Bytes/8)
	L := make([]uint64, mayo.lBytes/8)
	bytesToUint64Slice(P1, P1Bytes)
	bytesToUint64Slice(L, LBytes)

	// Hash the message, and derive salt and t
	mDigest := rand.SHAKE256(mayo.digestBytes, m)
	R := rand.SampleRandomBytes(mayo.rBytes)
	salt := rand.SHAKE256(mayo.saltBytes, mDigest, R, seedSk)
	t := decodeVec(mayo.m, rand.SHAKE256(mayo.intTimesLogQ(mayo.m), mDigest, salt))

	mTemp := make([]uint64, mayo.k*mayo.o*4) // TODO: mVecLimbs = 4

	// Attempt to find a preimage for t
	x := make([]byte, mayo.k*mayo.n)
	var v []byte
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := rand.SHAKE256(mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o), mDigest, salt, seedSk, []byte{byte(ctr)})
		for i := 0; i < mayo.k; i++ {
			v = append(v, decodeVec(mayo.n-mayo.o, V[i*mayo.vBytes:(i+1)*mayo.vBytes])...)
		}
		r := decodeVec(mayo.k*mayo.o, V[mayo.k*mayo.vBytes:mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o)])

		// Build linear system Ax = y
		A := make([]uint64, (((mayo.m+7)/8*8)*(mayo.k*mayo.o+1))/8)
		y := make([]byte, mayo.m)
		mayo.computeMAndVpv(v, L, P1, mTemp, A)
		mayo.computeRhs(A, t, y)

		aBytes := make([]byte, ((mayo.m+7)/8*8)*(mayo.k*mayo.o+1))
		uint64SliceToBytes(aBytes, A)
		mayo.computeA(mTemp, aBytes)

		for i := 0; i < mayo.m; i++ {
			aBytes[(1+i)*(mayo.k*mayo.o+1)-1] = 0
		}

		aCols := mayo.k*mayo.o + 1
		hasSolution := mayo.sampleSolution(aBytes, y, r, x, mayo.k, mayo.o, mayo.m, aCols)

		if hasSolution {
			break
		}
	}

	s := make([]byte, mayo.k*mayo.n)
	Ox := make([]byte, mayo.v)
	for i := 0; i < mayo.k; i++ {
		mayo.matMul(O, x[i*mayo.o:], Ox, mayo.o, mayo.n-mayo.o, 1)
		mayo.matAdd(v[i*(mayo.n-mayo.o):], Ox, s, i*mayo.n, mayo.n-mayo.o, 1)
		copy(s[i*mayo.n+mayo.n-mayo.o:], x[i*mayo.o:])
	}

	// Finish and output the signature
	var sig []byte
	sig = append(sig, encodeVec(s)...)
	sig = append(sig, salt...)
	return sig
}

// Verify (Algorithm 8) takes an expanded public key, message m, and signature sig and outputs an integer to indicate
// if the signature is valid on m. Specifically if the signature is valid it will output 0, if invalid < 0.
func (mayo *Mayo) Verify(epk, m, sig []byte) int {
	// Decode epk
	P1 := make([]uint64, mayo.p1Bytes/8)
	P2 := make([]uint64, mayo.p2Bytes/8)
	P3 := make([]uint64, mayo.p3Bytes/8)
	bytesToUint64Slice(P1, epk[:mayo.p1Bytes])
	bytesToUint64Slice(P2, epk[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes])
	bytesToUint64Slice(P3, epk[mayo.p1Bytes+mayo.p2Bytes:mayo.p1Bytes+mayo.p2Bytes+mayo.p3Bytes])

	// Decode sig
	nkHalf := int(math.Ceil(float64(mayo.n) * float64(mayo.k) / 2.0))
	salt := sig[nkHalf : nkHalf+mayo.saltBytes]
	s := decodeVec(mayo.k*mayo.n, sig)

	// Hash the message and derive t
	mDigest := rand.SHAKE256(mayo.digestBytes, m)
	t := decodeVec(mayo.m, rand.SHAKE256(mayo.intTimesLogQ(mayo.m), mDigest, salt))

	// Compute P^*(s)
	y := make([]byte, 2*mayo.m)
	mayo.evalPublicMap(s, P1, P2, P3, y)
	y = y[:mayo.m] // TODO: handle this differently?

	// Accept the signature if y = t
	if bytes.Equal(y, t) {
		return 0
	}
	return -1
}

// APISign (Algorithm 9) Takes a secret sk and message, it then expands the SK and calls Sign with the expanded secret key
// to produce the signature. It then outputs sig || M
func (mayo *Mayo) APISign(M, sk []byte) []byte {
	// Expand the SK
	esk := mayo.ExpandSK(sk)

	// Produce signature
	sig := mayo.Sign(esk, M)

	// Return signed message
	result := make([]byte, mayo.sigBytes+len(M))
	copy(result[:mayo.sigBytes], sig)
	copy(result[mayo.sigBytes:], M)
	return result
}

// APISignOpen (Algorithm 10) Takes a signed message sig || m as input and expands the public key, which then calls
// Verify to check if the signature is valid. It returns the result and message if the signature is valid
func (mayo *Mayo) APISignOpen(sm, pk []byte) (int, []byte) {
	// Expand the PK
	epk := mayo.ExpandPK(pk)

	// Parse the signed message
	sig, M := sm[:mayo.sigBytes], sm[mayo.sigBytes:]

	// Verify the signature
	result := mayo.Verify(epk, M, sig)

	// Return result and message
	if result < 0 {
		return result, nil
	}
	return result, M
}

func (mayo *Mayo) intTimesLogQ(ints ...int) int {
	product := 1
	for _, number := range ints {
		product *= number
	}

	return int(math.Ceil(float64(product) * math.Log2(float64(mayo.q)) / 8.0))
}
