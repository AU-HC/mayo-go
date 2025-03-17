package mayo

import (
	"bytes"
	"math"
	"mayo-go/rand"
	"unsafe"
)

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead
// return an error, if it fails to generate random bytes.
func (mayo *Mayo) CompactKeyGen() (CompactPublicKey, CompactSecretKey) {
	// Pick seekSk at random
	var seedSk [skSeedBytes]byte
	rand.SampleRandomBytes(seedSk[:])

	// Derive seedPk and O from seekSk
	var s [pkSeedBytes + OBytes]byte
	rand.SHAKE256(s[:], seedSk[:])
	var seedPk [pkSeedBytes]byte
	copy(seedPk[:], s[:pkSeedBytes])
	O := decodeMatrix(v, o, s[pkSeedBytes:pkSeedBytes+OBytes])

	// Derive P_i^1 and P_i^2 from seekPk
	PBytes := make([]byte, P1Bytes+P2Bytes)
	rand.AES128CTR(seedPk[:], PBytes[:])
	var P [P1Limbs + P2Limbs]uint64
	unpackMVecs(PBytes, P[:], (P1Limbs+P2Limbs)/mVecLimbs)

	P1 := P[:P1Limbs]                  // v x v upper triangular matrix
	P2 := P[P1Limbs : P1Limbs+P2Limbs] // v x o matrix

	// Compute P3
	P3 := mayo.computeP3(P1, O, P2)

	// Output keys
	return CompactPublicKey{seedPk: seedPk, p3: P3}, CompactSecretKey{seedSk: seedSk}
}

func unpackMVecs(in []byte, out []uint64, vecs int) {
	tmp := make([]byte, M/2) // Temporary buffer for a single vector

	for i := vecs - 1; i >= 0; i-- {
		// Copy packed vector from `in` to `tmp`
		copy(tmp, in[i*M/2:i*M/2+M/2])

		// Copy `tmp` into the appropriate location in `out`
		outBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&out[0])))[:]
		copy(outBytes[i*mVecLimbs*8:], tmp)
	}
}

func packMVecs(in []uint64, out []byte, vecs int) {
	// Treat `in` as a byte slice for copying
	inBytes := (*(*[1 << 30]byte)(unsafe.Pointer(&in[0])))[:]

	for i := 0; i < vecs; i++ {
		copy(out[i*M/2:], inBytes[i*mVecLimbs*8:i*mVecLimbs*8+M/2])
	}
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk CompactSecretKey) ExpandedSecretKey {
	// Derive seedPk and O from seedSk
	var s [pkSeedBytes + OBytes]byte
	rand.SHAKE256(s[:], csk.seedSk[:])
	seedPk := s[:pkSeedBytes]
	var oByteString [OBytes]byte
	copy(oByteString[:], s[pkSeedBytes:pkSeedBytes+OBytes])
	O := decodeMatrix(v, o, oByteString[:])

	// Derive P1 and P2 from seedPk
	PBytes := make([]byte, P1Bytes+P2Bytes)
	rand.AES128CTR(seedPk[:], PBytes[:])
	var P [P1Limbs + P2Limbs]uint64
	unpackMVecs(PBytes, P[:], (P1Limbs+P2Limbs)/mVecLimbs)

	var P1 [P1Limbs]uint64
	var P2 [P2Limbs]uint64
	copy(P1[:], P[:P1Limbs])                // v x v upper triangular matrix
	copy(P2[:], P[P1Limbs:P1Limbs+P2Limbs]) // v x o matrix

	// Compute L and store in P2
	mayo.computeL(P1[:], O, P2[:])

	return ExpandedSecretKey{
		seedSk: csk.seedSk,
		p1:     P1,
		l:      P2,
		o:      oByteString,
	}
}

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk CompactPublicKey) ExpandedPublicKey {
	// Parse cpk
	seedPk := cpk.seedPk

	// Derive P_i^1 and P_i^2 from seekPk
	PBytes := make([]byte, P1Bytes+P2Bytes)
	rand.AES128CTR(seedPk[:], PBytes[:])
	var P [P1Limbs + P2Limbs]uint64
	unpackMVecs(PBytes, P[:], (P1Limbs+P2Limbs)/mVecLimbs)

	P1 := P[:P1Limbs]                  // v x v upper triangular matrix
	P2 := P[P1Limbs : P1Limbs+P2Limbs] // v x o matrix

	return ExpandedPublicKey{
		p1: [P1Limbs]uint64(P1),
		p2: [P2Limbs]uint64(P2),
		p3: cpk.p3,
	}
}

// Sign (Algorithm 7) takes an expanded secret key esk and a message M and outputs a signature on the message M
func (mayo *Mayo) Sign(esk ExpandedSecretKey, message []byte) []byte {
	// Decode O
	var O [v * o]byte
	decodeVec(O[:], esk.o[:])

	// Hash the message, and derive salt and t
	var mDigest [digestBytes]byte
	rand.SHAKE256(mDigest[:], message)
	var R [rBytes]byte
	rand.SampleRandomBytes(R[:])
	var salt [saltBytes]byte
	rand.SHAKE256(salt[:], mDigest[:], R[:], esk.seedSk[:])
	var t [M]byte
	decodeVec(t[:], rand.SHAKE256Slow(mayo.intTimesLogQ(M), mDigest[:], salt[:]))

	// Attempt to find a preimage for t
	var mTemp [K * o * mVecLimbs]uint64
	var x [K * N]byte
	var vDec [K * v]byte
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := rand.SHAKE256Slow(K*vBytes+mayo.intTimesLogQ(K, o), mDigest[:], salt[:], esk.seedSk[:], []byte{byte(ctr)})
		for i := 0; i < K; i++ {
			offset := i * v
			decodeVec(vDec[offset:offset+v], V[i*vBytes:(i+1)*vBytes])
		}
		var r [K * o]byte
		decodeVec(r[:], V[K*vBytes:K*vBytes+mayo.intTimesLogQ(K, o)])

		// Build linear system Ax = y
		var A [(((M + 7) / 8 * 8) * (K*o + 1)) / 8]uint64
		var y [M]byte
		mayo.computeMAndVpv(vDec[:], esk.l[:], esk.p1[:], mTemp[:], A[:])
		mayo.computeRhs(A[:], t[:], y[:])

		var aBytes [((M + 7) / 8 * 8) * (K*o + 1)]byte
		uint64SliceToBytes(aBytes[:], A[:]) // todo: pack or not?
		mayo.computeA(mTemp[:], aBytes[:])

		for i := 0; i < M; i++ {
			aBytes[(1+i)*(K*o+1)-1] = 0
		}

		hasSolution := mayo.sampleSolution(aBytes[:], y[:], r[:], x[:])

		if hasSolution {
			break
		}
	}

	var s [K * N]byte
	var Ox [v]byte
	for i := 0; i < K; i++ {
		mayo.matMul(O[:], x[i*o:], Ox[:], o, v, 1)
		mayo.matAdd(vDec[i*(v):], Ox[:], s[:], i*N, v, 1)
		copy(s[i*N+v:], x[i*o:])
	}

	// Finish and output the signature
	var sig [sigBytes]byte
	copy(sig[:], encodeVec(s[:]))
	copy(sig[sigBytes-saltBytes:], salt[:])
	return sig[:]
}

// Verify (Algorithm 8) takes an expanded public key, message M, and signature sig and outputs an integer to indicate
// if the signature is valid on M. Specifically if the signature is valid it will output 0, if invalid < 0.
func (mayo *Mayo) Verify(epk ExpandedPublicKey, message, sig []byte) int {
	// Decode sig
	nkHalf := int(math.Ceil(float64(N) * float64(K) / 2.0))
	salt := sig[nkHalf : nkHalf+saltBytes]
	var s [K * N]byte
	decodeVec(s[:], sig)

	// Hash the message and derive t
	var mDigest [digestBytes]byte
	rand.SHAKE256(mDigest[:], message)
	var t [M]byte
	decodeVec(t[:], rand.SHAKE256Slow(mayo.intTimesLogQ(M), mDigest[:], salt))

	// Compute P^*(s)
	var y [2 * M]byte
	mayo.evalPublicMap(s[:], epk.p1[:], epk.p2[:], epk.p3[:], y[:])

	// Accept the signature if y = t
	if bytes.Equal(y[:M], t[:]) {
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
	result := make([]byte, sigBytes+len(M))
	copy(result[:sigBytes], sig)
	copy(result[sigBytes:], M)
	return result
}

// APISignOpen (Algorithm 10) Takes a signed message sig || M as input and expands the public key, which then calls
// Verify to check if the signature is valid. It returns the result and message if the signature is valid
func (mayo *Mayo) APISignOpen(sm []byte, pk CompactPublicKey) (int, []byte) {
	// Expand the PK
	epk := mayo.ExpandPK(pk)

	// Parse the signed message
	sig, message := sm[:sigBytes], sm[sigBytes:]

	// Verify the signature
	result := mayo.Verify(epk, message, sig)

	// Return result and message
	if result < 0 {
		return result, nil
	}
	return result, message
}

func (mayo *Mayo) intTimesLogQ(ints ...int) int {
	product := 1
	for _, number := range ints {
		product *= number
	}

	return int(math.Ceil(float64(product) * math.Log2(float64(q)) / 8.0))
}
