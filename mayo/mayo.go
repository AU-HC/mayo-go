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
	P := rand.AES128CTR32(seedPk, mayo.p1Bytes+mayo.p2Bytes)
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
	P := rand.AES128CTR32(seedPk, mayo.p1Bytes+mayo.p2Bytes)
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
	P1 := decodeMatrices(mayo.m, mayo.v, mayo.v, esk[mayo.skSeedBytes+mayo.oBytes:mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes], true)
	L := decodeMatrices(mayo.m, mayo.v, mayo.o, esk[mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes:mayo.eskBytes], false)

	// Hash the message, and derive salt and t
	mDigest := rand.SHAKE256(mayo.digestBytes, m)
	R := rand.SampleRandomBytes(mayo.rBytes)
	salt := rand.SHAKE256(mayo.saltBytes, mDigest, R, seedSk)
	t := decodeVec(mayo.m, rand.SHAKE256(mayo.intTimesLogQ(mayo.m), mDigest, salt))

	// Attempt to find a preimage for t
	var x []byte
	var hasSolution bool
	v := make([][]byte, mayo.k)
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := rand.SHAKE256(mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o), mDigest, salt, seedSk, []byte{byte(ctr)})
		for i := 0; i < mayo.k; i++ {
			v[i] = decodeVec(mayo.n-mayo.o, V[i*mayo.vBytes:(i+1)*mayo.vBytes])
		}
		r := decodeVec(mayo.k*mayo.o, V[mayo.k*mayo.vBytes:mayo.k*mayo.vBytes+mayo.intTimesLogQ(mayo.k, mayo.o)])

		// Build linear system Ax = y
		A := generateZeroMatrix(mayo.m+mayo.shifts, mayo.k*mayo.o)
		y := make([]byte, mayo.m+mayo.shifts)
		copy(y[:mayo.m], t)
		ell := 0
		M := make([][][]byte, mayo.k)
		for i := 0; i < mayo.k; i++ {
			mi := generateZeroMatrix(mayo.m, mayo.o)

			for j := 0; j < mayo.m; j++ {
				mi[j] = mayo.field.MultiplyMatrices(transposeVector(v[i]), L[j])[0]
			}

			M[i] = mi
		}

		for i := 0; i < mayo.k; i++ {
			// Calculate v_i P1 and v_i P1 v_i
			viP := make([][]byte, mayo.m)
			viPvi := make([]byte, mayo.m)
			for a := 0; a < mayo.m; a++ {
				viP[a] = mayo.field.VectorTransposedMatrixMul(v[i], P1[a])
				viPvi[a] = mayo.field.VecInnerProduct(viP[a], v[i])
			}

			for j := mayo.k - 1; j >= i; j-- {
				u := make([]byte, mayo.m)
				if i == j {
					for a := 0; a < mayo.m; a++ {
						u[a] = viPvi[a]
					}
				} else {
					for a := 0; a < mayo.m; a++ {
						u[a] = mayo.field.VecInnerProduct(viP[a], v[j]) ^
							mayo.field.VecInnerProduct(mayo.field.VectorTransposedMatrixMul(v[j], P1[a]), v[i])
					}
				}

				// Calculate y = y - z^l * u
				for d := 0; d < mayo.m; d++ {
					y[d+ell] ^= u[d]
				}

				for row := 0; row < mayo.m; row++ {
					for column := i * mayo.o; column < (i+1)*mayo.o; column++ {
						A[row+ell][column] ^= M[j][row][column%mayo.o]
					}

					if i != j {
						for column := j * mayo.o; column < (j+1)*mayo.o; column++ {
							A[row+ell][column] ^= M[i][row][column%mayo.o]
						}
					}
				}

				ell += 1
			}
		}

		// Reduce y and the columns of A mod f(x)
		y = mayo.reduceVecModF(y)
		A = mayo.reduceAModF(A)

		// Try to solve the system
		x, hasSolution = mayo.sampleSolution(A, y, r)
		if hasSolution {
			break
		}
	}

	// Finish and output the signature
	var s []byte
	for i := 0; i < mayo.k; i++ {
		xIndexed := x[i*mayo.o : (i+1)*mayo.o]
		OX := field.AddVec(v[i], mayo.field.MatrixVectorMul(O, xIndexed))
		s = append(s, OX...)
		s = append(s, xIndexed...)
	}
	var sig []byte
	sig = append(sig, encodeVec(s)...)
	sig = append(sig, salt...)
	return sig
}

// Verify (Algorithm 8) takes an expanded public key, message m, and signature sig and outputs an integer to indicate
// if the signature is valid on m. Specifically if the signature is valid it will output 0, if invalid < 0.
func (mayo *Mayo) Verify(epk, m, sig []byte) int {
	// Decode epk
	P1ByteString := epk[:mayo.p1Bytes]
	P2ByteString := epk[mayo.p1Bytes : mayo.p1Bytes+mayo.p2Bytes]
	P3ByteString := epk[mayo.p1Bytes+mayo.p2Bytes : mayo.p1Bytes+mayo.p2Bytes+mayo.p3Bytes]
	P1 := decodeMatrices(mayo.m, mayo.v, mayo.v, P1ByteString, true)
	P2 := decodeMatrices(mayo.m, mayo.v, mayo.o, P2ByteString, false)
	P3 := decodeMatrices(mayo.m, mayo.o, mayo.o, P3ByteString, true)

	// Decode sig
	nkHalf := int(math.Ceil(float64(mayo.n) * float64(mayo.k) / 2.0))
	salt := sig[nkHalf : nkHalf+mayo.saltBytes]
	s := decodeVec(mayo.k*mayo.n, sig)
	sVector := make([][]byte, mayo.k)
	for i := 0; i < mayo.k; i++ {
		sVector[i] = make([]byte, mayo.n)
		copy(sVector[i], s[i*mayo.n:(i+1)*mayo.n])
	}

	// Hash the message and derive t
	mDigest := rand.SHAKE256(mayo.digestBytes, m)
	t := decodeVec(mayo.m, rand.SHAKE256(mayo.intTimesLogQ(mayo.m), mDigest, salt))

	// Compute P^*(s)
	P := mayo.calculateP(P1, P2, P3)
	y := make([]byte, mayo.m+(mayo.k*(mayo.k+1)/2))
	ell := 0
	for i := 0; i < mayo.k; i++ {
		// Calculate s_i P and s_i P s_i
		siP := make([][]byte, mayo.m)
		siPsi := make([]byte, mayo.m)
		for a := 0; a < mayo.m; a++ {
			siP[a] = mayo.field.VectorTransposedMatrixMul(sVector[i], P[a])
			siPsi[a] = mayo.field.VecInnerProduct(siP[a], sVector[i])
		}

		for j := mayo.k - 1; j >= i; j-- {
			u := make([]byte, mayo.m)
			if i == j {
				for a := 0; a < mayo.m; a++ {
					u[a] = siPsi[a]
				}
			} else {
				for a := 0; a < mayo.m; a++ {
					u[a] = mayo.field.VecInnerProduct(siP[a], sVector[j]) ^
						mayo.field.VecInnerProduct(mayo.field.VectorTransposedMatrixMul(sVector[j], P[a]), sVector[i])
				}
			}

			// Calculate y = y - z^l * u
			for d := 0; d < mayo.m; d++ {
				y[d+ell] ^= u[d]
			}

			ell += 1
		}
	}

	// Reduce y mod f(x)
	y = mayo.reduceVecModF(y)

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

func (mayo *Mayo) reduceVecModF(y []byte) []byte {
	for i := mayo.m + mayo.shifts - 1; i >= mayo.m; i-- {
		for shift, coefficient := range mayo.tailF {
			y[i-mayo.m+shift] ^= mayo.field.Gf16Mul(y[i], coefficient)
		}
		y[i] = 0
	}
	y = y[:mayo.m]

	return y
}

func (mayo *Mayo) reduceAModF(A [][]byte) [][]byte {
	for row := mayo.m + mayo.shifts - 1; row >= mayo.m; row-- {
		for column := 0; column < mayo.k*mayo.o; column++ {
			for shift := 0; shift < len(mayo.tailF); shift++ {
				A[row-mayo.m+shift][column] ^= mayo.field.Gf16Mul(A[row][column], mayo.tailF[shift])
			}
			A[row][column] = 0
		}
	}
	A = A[:mayo.m]

	return A
}

func (mayo *Mayo) calculateP(P1, P2, P3 [][][]byte) [][][]byte {
	P := make([][][]byte, mayo.m)
	for i := 0; i < mayo.m; i++ {
		P[i] = make([][]byte, mayo.n)
		for j := 0; j < mayo.n; j++ {
			P[i][j] = make([]byte, mayo.n)
		}
	}

	for i := 0; i < mayo.m; i++ {
		// Set P1
		for row := 0; row < mayo.v; row++ {
			for column := 0; column < mayo.v; column++ {
				P[i][row][column] = P1[i][row][column]
			}
		}
		// Set P2
		for row := 0; row < mayo.v; row++ {
			for column := 0; column < mayo.o; column++ {
				P[i][row][column+mayo.v] = P2[i][row][column]
			}
		}
		// Set P3
		for row := 0; row < mayo.o; row++ {
			for column := 0; column < mayo.o; column++ {
				P[i][row+mayo.v][column+mayo.v] = P3[i][row][column]
			}
		}
	}

	return P
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

func (mayo *Mayo) sampleSolution(A [][]byte, y []byte, R []byte) ([]byte, bool) {
	// Randomize the system using r
	x := make([]byte, len(R))
	copy(x, R)

	yMatrix := field.AddVec(y, mayo.field.MatrixVectorMul(A, R))

	// Put (A y) in echelon form with leading 1's
	AyMatrix := appendVecToMatrix(A, yMatrix)
	AyMatrix = mayo.echelonForm(AyMatrix)
	A, y = extractVecFromMatrix(AyMatrix)

	// Check if A has rank m
	zeroVector := make([]byte, mayo.k*mayo.o)
	if bytes.Equal(A[mayo.m-1], zeroVector) {
		return nil, false
	}

	// Back-substitution
	for r := mayo.m - 1; r >= 0; r-- {
		// Let c be the index of first non-zero element of A[r,:]
		for c := 0; c < len(A[r]); c++ {
			if A[r][c] != 0 {
				x[c] ^= y[r]

				for i := 0; i < mayo.m; i++ {
					y[i] ^= mayo.field.Gf16Mul(y[r], A[i][c])
				}

				break
			}
		}
	}

	return x, true
}
