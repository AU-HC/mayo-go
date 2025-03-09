package mayo

import (
	"bytes"
	"math"
	"mayo-go/field"
	"mayo-go/rand"
	"slices"
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
	O := decodeMatrix(mayo.v, mayo.o, esk[mayo.skSeedBytes:mayo.skSeedBytes+mayo.oBytes])
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
	var x []byte
	v := make([][]byte, mayo.k)
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := rand.SHAKE256(mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o), mDigest, salt, seedSk, []byte{byte(ctr)})
		for i := 0; i < mayo.k; i++ {
			v[i] = decodeVec(mayo.n-mayo.o, V[i*mayo.vBytes:(i+1)*mayo.vBytes])
		}
		r := decodeVec(mayo.k*mayo.o, V[mayo.k*mayo.vBytes:mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o)])

		// Build linear system Ax = y
		A := make([]byte, 0) // TODO: length
		y := make([]byte, mayo.m)

		// Compute M and vPv
		mayo.computeMAndVpv(v, L, P1, mTemp, A)
		mayo.computeRhs(A, t, y)
		mayo.computeA(mTemp, A) // TODO

		for i := 0; i < mayo.m; i++ {
			A[(1+i)*(mayo.k*mayo.o+1)-1] = 0
		}

		aCols := 0 // TODO: figure out
		hasSolution := mayo.sampleSolutionOpti(A, y, r, x, mayo.k, mayo.o, mayo.m, aCols)

		if hasSolution {
			break
		}
	}

	// Finish and output the signature
	var sig []byte
	sig = append(sig, t...)
	sig = append(sig, x...)
	sig = append(sig, salt...)
	return sig
}

// Verify (Algorithm 8) takes an expanded public key, message m, and signature sig and outputs an integer to indicate
// if the signature is valid on m. Specifically if the signature is valid it will output 0, if invalid < 0.
func (mayo *Mayo) Verify(epk, m, sig []byte) int {
	// Decode epk TODO: probably refactor this
	P1ByteString := epk[:mayo.p1Bytes]
	P2ByteString := epk[mayo.p1Bytes : mayo.p1Bytes+mayo.p2Bytes]
	P3ByteString := epk[mayo.p1Bytes+mayo.p2Bytes : mayo.p1Bytes+mayo.p2Bytes+mayo.p3Bytes]
	P1 := make([]uint64, mayo.p1Bytes/8)
	P2 := make([]uint64, mayo.p2Bytes/8)
	P3 := make([]uint64, mayo.p3Bytes/8)
	bytesToUint64Slice(P1, P1ByteString)
	bytesToUint64Slice(P2, P2ByteString)
	bytesToUint64Slice(P3, P3ByteString)

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

func (mayo *Mayo) echelonForm(B [][]byte) [][]byte {
	pivotColumn := 0
	pivotRow := 0

	for pivotRow < mayo.m && pivotColumn < mayo.o*mayo.k+1 {
		var possiblePivots []int
		for i := pivotRow; i < mayo.m; i++ {
			if B[i][pivotColumn] != 0 {
				possiblePivots = append(possiblePivots, i)
			}
		}

		if len(possiblePivots) == 0 {
			pivotColumn++
			continue
		}

		nextPivotRow := slices.Min(possiblePivots)
		B[pivotRow], B[nextPivotRow] = B[nextPivotRow], B[pivotRow]

		// Make the leading entry a 1
		B[pivotRow] = mayo.field.MultiplyVecConstant(mayo.field.Gf16Inv(B[pivotRow][pivotColumn]), B[pivotRow])

		// Eliminate entries below the pivot
		for row := nextPivotRow + 1; row < mayo.m; row++ {
			B[row] = field.AddVec(B[row], mayo.field.MultiplyVecConstant(B[row][pivotColumn], B[pivotRow]))
		}

		pivotRow++
		pivotColumn++
	}

	return B
}

func (mayo *Mayo) lincomb(a []byte, b []byte, aCounter, j, n, m int) byte {
	var ret byte
	bCounter := 0
	for i := 0; i < n; i++ {
		bCounter += m
		ret = mayo.field.Gf16Mul(a[aCounter+i], b[j+bCounter]) ^ ret
	}
	return ret
}

func (mayo *Mayo) matMul(a []byte, b []byte, c []byte, colRowAb, rowA, colB int) {
	cCounter := 0
	aCounter := 0
	for i := 0; i < rowA; i++ {
		aCounter += colRowAb
		for j := 0; j < colB; j++ {
			cCounter++
			c[cCounter] = mayo.lincomb(a, b, aCounter, j, colRowAb, colB)
		}
	}
}

func (mayo *Mayo) echelonForm2(A []byte, nRows int, nCols int) {
	rowLen := (nCols + 15) / 16

	for i := 0; i < nRows; i++ {
		mayo.ef_pack_m_vec(A, i*nCols, packedA, i*rowLen, nCols)
	}
}

func (mayo *Mayo) sampleSolutionOpti(A []byte, y, r, x []byte, k, o, m, aCols int) bool {
	// x <- r
	copy(x, r)

	// compute Ar;
	Ar := make([]byte, m)
	for i := 0; i < m; i++ {
		A[k*o+i*(k*o+1)] = 0 // clear last col of A
	}
	mayo.matMul(A, r, Ar, k*o+1, m, 1)

	// move y - Ar to last column of matrix A
	for i := 0; i < m; i++ {
		A[k*o+i*(k*o+1)] = y[i] ^ Ar[i]
	}

	mayo.echelonForm2(A, m, k*o+1)

	// check if last row of A (excluding the last entry of y) is zero
	var fullRank byte
	for i := 0; i < aCols-1; i++ {
		fullRank |= A[(m-1)*aCols+i]
	}

	if fullRank == 0 {
		return false
	}

	for row := m - 1; row >= 0; row-- {
		var finished byte
		colUpperBound := int(math.Min(float64(row+(32/(m-row))), float64(k*o)))

		for col := row; col <= colUpperBound; col++ {
			correctColumn := mayo.ct_compare_8(A[row*aCols+col], 0) & finished // TODO: implement

			u := correctColumn & A[row*aCols+aCols-1]
			x[col] ^= u

			for i := 0; i < row; i += 8 {
				tmp := (uint64(A[i*aCols+col]) << 0) ^
					(uint64(A[(i+1)*aCols+col]) << 8) ^
					(uint64(A[(i+2)*aCols+col]) << 16) ^
					(uint64(A[(i+3)*aCols+col]) << 24) ^
					(uint64(A[(i+4)*aCols+col]) << 32) ^
					(uint64(A[(i+5)*aCols+col]) << 40) ^
					(uint64(A[(i+6)*aCols+col]) << 48) ^
					(uint64(A[(i+7)*aCols+col]) << 56)

				tmp = mulFx8(u, tmp)

				A[i*aCols+aCols-1] ^= byte((tmp) & 0xf)
				A[(i+1)*aCols+aCols-1] ^= byte((tmp >> 8) & 0xf)
				A[(i+2)*aCols+aCols-1] ^= byte((tmp >> 16) & 0xf)
				A[(i+3)*aCols+aCols-1] ^= byte((tmp >> 24) & 0xf)
				A[(i+4)*aCols+aCols-1] ^= byte((tmp >> 32) & 0xf)
				A[(i+5)*aCols+aCols-1] ^= byte((tmp >> 40) & 0xf)
				A[(i+6)*aCols+aCols-1] ^= byte((tmp >> 48) & 0xf)
				A[(i+7)*aCols+aCols-1] ^= byte((tmp >> 56) & 0xf)
			}

			finished = finished | correctColumn
		}
	}

	return true
}

func mulFx8(a byte, b uint64) uint64 { // TODO: Move this
	// carry-less multiply
	var p uint64
	p = uint64(a&1) * b
	p ^= uint64(a&2) * b
	p ^= uint64(a&4) * b
	p ^= uint64(a&8) * b

	// reduce mod x^4 + x + 1
	topP := p & 0xf0f0f0f0f0f0f0f0
	out := (p ^ (topP >> 4) ^ (topP >> 3)) & 0x0f0f0f0f0f0f0f0f
	return out
}
