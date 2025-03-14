package mayo

import (
	"bytes"
	"fmt"
	"math"
	"mayo-go/rand"
)

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead
// return an error, if it fails to generate random bytes.
func (mayo *Mayo) CompactKeyGen() ([]byte, []byte, error) {
	// Pick seekSk at random
	seedSk := rand.SampleRandomBytes(skSeedBytes)

	// Derive seedPk and O from seekSk
	s := rand.SHAKE256(pkSeedBytes+OBytes, seedSk)
	seedPk := s[:pkSeedBytes]
	O := decodeMatrix(N-o, o, s[pkSeedBytes:pkSeedBytes+OBytes])

	// Derive P_i^1 and P_i^2 from seekPk
	P := rand.AES128CTR64(seedPk, P1Bytes+P2Bytes)
	P1 := P[:P1Bytes/8]                      // v x v upper triangular matrix
	P2 := P[P1Bytes/8 : (P1Bytes+P2Bytes)/8] // v x o matrix

	// Compute P3
	P3ByteArray := mayo.computeP3(P1, O, P2)

	// Encode the compact public/secret key
	var cpk [cpkBytes]byte
	copy(cpk[:pkSeedBytes], seedPk)
	copy(cpk[pkSeedBytes:], P3ByteArray)
	csk := seedSk

	// Output keys
	return cpk[:], csk, nil
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk []byte) []byte {
	// Parse csk
	seedSk := csk[:skSeedBytes]

	// Derive seedPk and O from seedSk
	S := rand.SHAKE256(pkSeedBytes+OBytes, seedSk)
	seedPk := S[:pkSeedBytes]
	oByteString := S[pkSeedBytes : pkSeedBytes+OBytes]
	O := decodeMatrix(N-o, o, oByteString)

	// Derive P1 and P2 from seedPk
	P := rand.AES128CTR64(seedPk, P1Bytes+P2Bytes)
	P1 := P[:P1Bytes/8]                      // v x v upper triangular matrix
	P2 := P[P1Bytes/8 : (P1Bytes+P2Bytes)/8] // v x o matrix

	// Compute L
	lByteArray := mayo.computeL(P1, O, P2)

	// Encode the SK and output esk
	var p1Bytes [P1Bytes]byte
	uint64SliceToBytes(p1Bytes[:], P1)

	var esk [eskBytes]byte
	copy(esk[:skSeedBytes], seedSk)
	copy(esk[skSeedBytes:], oByteString)
	copy(esk[skSeedBytes+OBytes:], p1Bytes[:])
	copy(esk[skSeedBytes+OBytes+P1Bytes:], lByteArray)
	return esk[:]
}

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk []byte) []byte {
	// Parse cpk
	seedPk := cpk[:pkSeedBytes]

	// Expand seedPk and return epk
	var epk [epkBytes]byte
	copy(epk[:P1Bytes+P2Bytes], rand.AES128CTR(seedPk, P1Bytes+P2Bytes))
	copy(epk[P1Bytes+P2Bytes:], cpk[pkSeedBytes:pkSeedBytes+P3Bytes])
	return epk[:]
}

// Sign (Algorithm 7) takes an expanded secret key esk and a message M and outputs a signature on the message M
func (mayo *Mayo) Sign(esk, message []byte) []byte {
	// Decode esk
	fmt.Println(fmt.Sprintf("lbytes: %d", lBytes))
	seedSk := esk[:skSeedBytes]
	O := decodeVec(v*o, esk[skSeedBytes:skSeedBytes+OBytes])
	P1ByteArray := esk[skSeedBytes+OBytes : skSeedBytes+OBytes+P1Bytes]
	LBytes := esk[skSeedBytes+OBytes+P1Bytes : eskBytes]
	var P1 [P1Bytes / 8]uint64
	var L [lBytes / 8]uint64
	bytesToUint64Slice(P1[:], P1ByteArray)
	bytesToUint64Slice(L[:], LBytes)

	// Hash the message, and derive salt and t
	mDigest := rand.SHAKE256(digestBytes, message)
	R := rand.SampleRandomBytes(rBytes)
	salt := rand.SHAKE256(saltBytes, mDigest, R, seedSk)
	t := decodeVec(M, rand.SHAKE256(mayo.intTimesLogQ(M), mDigest, salt))

	// Attempt to find a preimage for t
	var mTemp [K * o * mVecLimbs]uint64
	var x [K * N]byte
	var vDec []byte // TODO: set length
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := rand.SHAKE256(K*vBytes+mayo.intTimesLogQ(K, o), mDigest, salt, seedSk, []byte{byte(ctr)})
		for i := 0; i < K; i++ {
			vDec = append(vDec, decodeVec(N-o, V[i*vBytes:(i+1)*vBytes])...)
		}
		r := decodeVec(K*o, V[K*vBytes:K*vBytes+mayo.intTimesLogQ(K, o)])

		// Build linear system Ax = y
		A := make([]uint64, (((M+7)/8*8)*(K*o+1))/8)
		y := make([]byte, M)
		mayo.computeMAndVpv(vDec, L[:], P1[:], mTemp[:], A)
		mayo.computeRhs(A, t, y)

		aBytes := make([]byte, ((M+7)/8*8)*(K*o+1))
		uint64SliceToBytes(aBytes, A)
		mayo.computeA(mTemp[:], aBytes)

		for i := 0; i < M; i++ {
			aBytes[(1+i)*(K*o+1)-1] = 0
		}

		hasSolution := mayo.sampleSolution(aBytes, y, r, x[:])

		if hasSolution {
			break
		}
	}

	var s [K * N]byte
	var Ox [v]byte
	for i := 0; i < K; i++ {
		mayo.matMul(O, x[i*o:], Ox[:], o, N-o, 1) // TODO: N - o = v?
		mayo.matAdd(vDec[i*(N-o):], Ox[:], s[:], i*N, N-o, 1)
		copy(s[i*N+N-o:], x[i*o:])
	}

	// Finish and output the signature
	var sig []byte // TODO: preallocate the size
	sig = append(sig, encodeVec(s[:])...)
	sig = append(sig, salt...)
	return sig
}

// Verify (Algorithm 8) takes an expanded public key, message M, and signature sig and outputs an integer to indicate
// if the signature is valid on M. Specifically if the signature is valid it will output 0, if invalid < 0.
func (mayo *Mayo) Verify(epk, message, sig []byte) int {
	// Decode epk
	var P1 [P1Bytes / 8]uint64
	var P2 [P2Bytes / 8]uint64
	var P3 [P3Bytes / 8]uint64
	bytesToUint64Slice(P1[:], epk[:P1Bytes])
	bytesToUint64Slice(P2[:], epk[P1Bytes:P1Bytes+P2Bytes])
	bytesToUint64Slice(P3[:], epk[P1Bytes+P2Bytes:P1Bytes+P2Bytes+P3Bytes])

	// Decode sig
	nkHalf := int(math.Ceil(float64(N) * float64(K) / 2.0))
	salt := sig[nkHalf : nkHalf+saltBytes]
	s := decodeVec(K*N, sig)

	// Hash the message and derive t
	mDigest := rand.SHAKE256(digestBytes, message)
	t := decodeVec(M, rand.SHAKE256(mayo.intTimesLogQ(M), mDigest, salt))

	// Compute P^*(s)
	var y [2 * M]byte
	mayo.evalPublicMap(s, P1[:], P2[:], P3[:], y[:]) // TODO dont use all

	// Accept the signature if y = t
	if bytes.Equal(y[:M], t) {
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
	copy(result[:sigBytes], sig)
	copy(result[sigBytes:], M)
	return result
}

// APISignOpen (Algorithm 10) Takes a signed message sig || M as input and expands the public key, which then calls
// Verify to check if the signature is valid. It returns the result and message if the signature is valid
func (mayo *Mayo) APISignOpen(sm []byte, pk []byte) (int, []byte) {
	// Expand the PK
	epk := mayo.ExpandPK(pk)

	// Parse the signed message
	sig, M := sm[:sigBytes], sm[sigBytes:]

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

	return int(math.Ceil(float64(product) * math.Log2(float64(q)) / 8.0))
}
