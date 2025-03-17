package mayo

import (
	"bytes"
	"math"
	"mayo-go/rand"
)

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead
// return an error, if it fails to generate random bytes.
func (mayo *Mayo) CompactKeyGen() (CompactPublicKey, CompactSecretKey, error) {
	// Pick seekSk at random
	seedSk := rand.SampleRandomBytes(mayo.skSeedBytes)

	// Derive seedPk and O from seekSk
	s := rand.SHAKE256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
	seedPk := s[:mayo.pkSeedBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes])

	// Derive P_i^1 and P_i^2 from seekPk
	P1, P2 := mayo.expandP1P2(seedPk)

	// Compute P3
	P3 := mayo.computeP3(P1, O, P2)

	// Output keys
	return CompactPublicKey{seedPk: seedPk, p3: P3}, CompactSecretKey{seedSk: seedSk}, nil
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk CompactSecretKey) ExpandedSecretKey {
	// Derive seedPk and O from seedSk
	S := rand.SHAKE256(mayo.pkSeedBytes+mayo.oBytes, csk.seedSk)
	seedPk := S[:mayo.pkSeedBytes]
	oByteString := S[mayo.pkSeedBytes : mayo.pkSeedBytes+mayo.oBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, oByteString)

	// Derive P1 and P2 from seedPk
	P1, P2 := mayo.expandP1P2(seedPk)

	// Compute L and store in P2
	mayo.computeL(P1, O, P2)

	return ExpandedSecretKey{
		seedSk: csk.seedSk,
		p1:     P1,
		l:      P2, // l is stored in P2
		o:      oByteString,
	}
}

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk CompactPublicKey) ExpandedPublicKey {
	// Derive P_i^1 and P_i^2 from seekPk
	seedPk := cpk.seedPk
	P1, P2 := mayo.expandP1P2(seedPk)

	return ExpandedPublicKey{
		p1: P1,
		p2: P2,
		p3: cpk.p3,
	}
}

// Sign (Algorithm 7) takes an expanded secret key esk and a message m and outputs a signature on the message m
func (mayo *Mayo) Sign(esk ExpandedSecretKey, m []byte) []byte {
	// Decode esk
	seedSk := esk.seedSk
	O := decodeVec(mayo.v*mayo.o, esk.o)
	P1 := esk.p1
	L := esk.l

	// Hash the message, and derive salt and t
	mDigest := rand.SHAKE256(mayo.digestBytes, m)
	R := rand.SampleRandomBytes(mayo.rBytes)
	salt := rand.SHAKE256(mayo.saltBytes, mDigest, R, seedSk)
	t := decodeVec(mayo.m, rand.SHAKE256(mayo.intTimesLogQ(mayo.m), mDigest, salt))

	mTemp := make([]uint64, mayo.k*mayo.o*mayo.mVecLimbs)

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
func (mayo *Mayo) Verify(epk ExpandedPublicKey, m, sig []byte) int {
	// Decode epk
	P1 := epk.p1
	P2 := epk.p2
	P3 := epk.p3

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

	// Accept the signature if y = t
	if bytes.Equal(y[:mayo.m], t) {
		return 0
	}
	return -1
}

// APISign (Algorithm 9) Takes a secret sk and message, it then expands the SK and calls Sign with the expanded secret key
// to produce the signature. It then outputs sig || M
func (mayo *Mayo) APISign(M []byte, sk CompactSecretKey) []byte {
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
func (mayo *Mayo) APISignOpen(sm []byte, pk CompactPublicKey) (int, []byte) {
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

func (mayo *Mayo) expandP1P2(seedPk []byte) ([]uint64, []uint64) {
	// First define array and fill it with bytes
	p1p2Bytes := rand.AES128CTR(seedPk[:], mayo.p1Bytes+mayo.p2Bytes)

	// Unpack the bytes
	P := make([]uint64, mayo.P1Limbs+mayo.P2Limbs)
	mayo.unpackMVecs(p1p2Bytes[:], P[:], (mayo.P1Limbs+mayo.P2Limbs)/mayo.mVecLimbs)

	return P[:mayo.P1Limbs], P[mayo.P1Limbs : mayo.P1Limbs+mayo.P2Limbs]
}

func (mayo *Mayo) intTimesLogQ(ints ...int) int {
	product := 1
	for _, number := range ints {
		product *= number
	}

	return int(math.Ceil(float64(product) * math.Log2(float64(mayo.q)) / 8.0))
}
