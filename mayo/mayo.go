package mayo

import (
	"bytes"
	cryptoRand "crypto/rand"
	"io"
	"math"
)

// CompactKeyGen (Algorithm 4) outputs compact representation of a secret key csk and public key cpk. Will instead
// return an error, if it fails to generate random bytes.
func (mayo *Mayo) CompactKeyGen() ([]byte, []byte, error) {
	seedSk := make([]byte, mayo.skSeedBytes)
	rand := cryptoRand.Reader
	_, err := io.ReadFull(rand, seedSk[:])
	if err != nil {
		return nil, nil, err
	}

	s := shake256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
	seedPk := s[:mayo.pkSeedBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, s[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.oBytes])

	P := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	P1 := decodeMatrices(mayo.m, mayo.v, mayo.v, P[:mayo.p1Bytes], true)
	P2 := decodeMatrices(mayo.m, mayo.v, mayo.o, P[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)
	P3 := make([][][]byte, mayo.m)
	for i := 0; i < mayo.m; i++ {
		P3[i] = upper(multiplyMatrices(transposeMatrix(O), addMatrices(multiplyMatrices(P1[i], O), P2[i])))
	}

	// Return the encoded cpk and csk
	cpk := make([]byte, mayo.cpkBytes)
	copy(cpk[:mayo.pkSeedBytes], seedPk)
	copy(cpk[mayo.pkSeedBytes:], encodeMatrices(mayo.o, mayo.o, P3, true))
	csk := seedSk

	return cpk, csk, nil
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk []byte) []byte {
	// Parse csk
	seedSk := csk[:mayo.skSeedBytes]

	// Derive seedPk and O from seedSk
	S := shake256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
	seedPk := S[:mayo.pkSeedBytes]
	oByteString := S[mayo.pkSeedBytes : mayo.pkSeedBytes+mayo.oBytes]
	O := decodeMatrix(mayo.n-mayo.o, mayo.o, oByteString)

	// Derive P1 and P2 from seedPk
	P := aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)
	p1Bytes := P[:mayo.p1Bytes]
	P1 := decodeMatrices(mayo.m, mayo.v, mayo.v, p1Bytes, true)
	P2 := decodeMatrices(mayo.m, mayo.v, mayo.o, P[mayo.p1Bytes:mayo.p1Bytes+mayo.p2Bytes], false)

	// Compute the L
	L := make([][][]byte, mayo.m)
	for i := 0; i < mayo.m; i++ {
		L[i] = addMatrices(multiplyMatrices(addMatrices(P1[i], transposeMatrix(P1[i])), O), P2[i])
	}

	// Encode L and output esk
	esk := make([]byte, mayo.eskBytes)
	copy(esk[:mayo.skSeedBytes], seedSk)
	copy(esk[mayo.skSeedBytes:], oByteString)
	copy(esk[mayo.skSeedBytes+mayo.oBytes:], p1Bytes)
	copy(esk[mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes:], encodeMatrices(mayo.v, mayo.o, L, false))
	return esk
}

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk []byte) []byte {
	// Parse cpk
	seedPk := cpk[:mayo.pkSeedBytes]

	// Expand seedPk and return epk
	epk := make([]byte, mayo.epkBytes)
	copy(epk[:mayo.p1Bytes+mayo.p2Bytes], aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes))
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
	mDigest := shake256(mayo.digestBytes, m)
	R := make([]byte, mayo.rBytes) // TODO: add randomization
	salt := shake256(mayo.saltBytes, mDigest, R, seedSk)
	t := decodeVec(mayo.m, shake256(int(math.Ceil(float64(mayo.m)*math.Log2(float64(mayo.q))/8.0)), mDigest, salt)) // TODO: refactor this length

	// Attempt to find a preimage for t
	var x []byte
	var hasSolution bool
	v := make([][]byte, mayo.k)
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := shake256(mayo.k*mayo.vBytes+int(math.Ceil(float64(mayo.k)*float64(mayo.o)*math.Log2(float64(mayo.q))/8)),
			mDigest, salt, seedSk, []byte{byte(ctr)})
		for i := 0; i < mayo.k; i++ {
			v[i] = decodeVec(mayo.n-mayo.o, V[i*mayo.vBytes:(i+1)*mayo.vBytes])
		}
		r := decodeVec(mayo.k*mayo.o, V[mayo.k*mayo.vBytes:mayo.k*mayo.vBytes+int(math.Ceil(float64(mayo.k)*float64(mayo.o)*math.Log2(float64(mayo.q))/8))])

		// Build linear system Ax = y
		A := generateZeroMatrix(mayo.m+mayo.shifts, mayo.k*mayo.o)
		y := make([]byte, mayo.m+mayo.shifts)
		copy(y[:mayo.m], t)
		ell := 0
		M := make([][][]byte, mayo.k)
		for i := 0; i < mayo.k; i++ {
			mi := generateZeroMatrix(mayo.m, mayo.o)

			for j := 0; j < mayo.m; j++ {
				mi[j] = multiplyMatrices(transposeVector(v[i]), L[j])[0]
			}

			M[i] = mi
		}

		for i := 0; i < mayo.k; i++ {
			for j := mayo.k - 1; j >= i; j-- {
				u := make([]byte, mayo.m)
				if i == j {
					for a := 0; a < mayo.m; a++ {
						vMatrix := vecToMatrix(v[i])
						u[a] = multiplyMatrices(multiplyMatrices(transposeVector(v[i]), P1[a]), vMatrix)[0][0]
					}
				} else {
					for a := 0; a < mayo.m; a++ {
						viMatrix := vecToMatrix(v[i])
						vjMatrix := vecToMatrix(v[j])
						u[a] = addMatrices(
							multiplyMatrices(multiplyMatrices(transposeVector(v[i]), P1[a]), vjMatrix),
							multiplyMatrices(multiplyMatrices(transposeVector(v[j]), P1[a]), viMatrix),
						)[0][0]
					}
				}

				// Calculate y = y - z^l * u
				for d := 0; d < mayo.m; d++ {
					y[d+ell] ^= u[d]
				}

				// TODO: Make this one for loop?
				for row := 0; row < mayo.m; row++ {
					for column := i * mayo.o; column < (i+1)*mayo.o; column++ {
						A[row+ell][column] ^= M[j][row][column%mayo.o]
					}
				}
				if i != j {
					for row := 0; row < mayo.m; row++ {
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
		x, hasSolution = mayo.SampleSolution(A, y, r)
		if hasSolution {
			break
		}
	}

	// Finish and output the signature
	var s []byte
	for i := 0; i < mayo.k; i++ {
		xIndexed := x[i*mayo.o : (i+1)*mayo.o]
		OX := transposeMatrix(addMatrices(vecToMatrix(v[i]), multiplyMatrices(O, vecToMatrix(xIndexed))))[0]

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
	mDigest := shake256(mayo.digestBytes, m)
	t := decodeVec(mayo.m, shake256(int(math.Ceil(float64(mayo.m)*math.Log2(float64(mayo.q))/8)), mDigest, salt))

	// Compute P^*(s)
	P := mayo.calculateP(P1, P2, P3)
	y := make([]byte, mayo.m+(mayo.k*(mayo.k+1)/2))
	ell := 0
	for i := 0; i < mayo.k; i++ {
		for j := mayo.k - 1; j >= i; j-- {
			u := make([]byte, mayo.m)
			if i == j {
				for a := 0; a < mayo.m; a++ {
					siMatrix := vecToMatrix(sVector[i])
					u[a] = multiplyMatrices(multiplyMatrices(transposeVector(sVector[i]), P[a]), siMatrix)[0][0]
				}
			} else {
				for a := 0; a < mayo.m; a++ {
					siMatrix := vecToMatrix(sVector[i])
					sjMatrix := vecToMatrix(sVector[j])
					u[a] = addMatrices(
						multiplyMatrices(multiplyMatrices(transposeVector(sVector[i]), P[a]), sjMatrix),
						multiplyMatrices(multiplyMatrices(transposeVector(sVector[j]), P[a]), siMatrix),
					)[0][0]
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

func (mayo *Mayo) reduceVecModF(y []byte) []byte {
	tailF := []byte{8, 0, 2, 8, 0}
	for i := mayo.m + mayo.shifts - 1; i >= mayo.m; i-- {
		for shift, coefficient := range tailF {
			y[i-mayo.m+shift] ^= gf16Mul(y[i], coefficient)
		}
		y[i] = 0
	}
	y = y[:mayo.m]

	return y
}

func (mayo *Mayo) reduceAModF(A [][]byte) [][]byte {
	tailF := []byte{8, 0, 2, 8, 0}
	for row := mayo.m + mayo.shifts - 1; row >= mayo.m; row-- {
		for column := 0; column < mayo.k*mayo.o; column++ {
			for shift := 0; shift < len(tailF); shift++ {
				A[row-mayo.m+shift][column] ^= gf16Mul(A[row][column], tailF[shift])
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
