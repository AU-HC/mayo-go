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
	rand := cryptoRand.Reader // TODO: refactor this prob
	_, err := io.ReadFull(rand, seedSk[:])
	if err != nil {
		return nil, nil, err
	}

	s := shake256(mayo.pkSeedBytes+mayo.oBytes, seedSk)
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

	var cpk []byte // TODO: is this a slow way of appended bytes? (General for entire file)
	cpk = append(cpk, seedPk...)
	cpk = append(cpk, encodeMatrixList(mayo.o, mayo.o, p3, true)...)
	csk := seedSk

	return cpk, csk, nil
}

// ExpandSK (Algorithm 5) takes the compacted secret key csk and outputs an expanded secret key esk
func (mayo *Mayo) ExpandSK(csk []byte) []byte {
	// Parse csk
	seedSk := csk[:mayo.skSeedBytes]

	// Derive seedPk and O from seedSk
	s := shake256(mayo.pkSeedBytes+mayo.oBytes, seedSk[:mayo.pkSeedBytes])
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

// ExpandPK (Algorithm 6) takes the compacted public key csk and outputs an expanded public key epk
func (mayo *Mayo) ExpandPK(cpk []byte) []byte {
	// Parse cpk
	seedPk := cpk[:mayo.pkSeedBytes]

	// Expand seedPk and return epk
	var epk []byte
	epk = append(epk, aes128ctr(seedPk, mayo.p1Bytes+mayo.p2Bytes)...)
	epk = append(epk, cpk[mayo.pkSeedBytes:mayo.pkSeedBytes+mayo.p3Bytes]...)
	return epk
}

// Sign (Algorithm 7) takes an expanded secret key esk and a message m and outputs a signature on the message m
func (mayo *Mayo) Sign(esk, m []byte) []byte {
	// Decode esk
	seedSk := esk[:mayo.skSeedBytes]
	O := decodeMatrix(mayo.v, mayo.o, esk[mayo.skSeedBytes:mayo.skSeedBytes+mayo.oBytes])
	P1 := decodeMatrixList(mayo.m, mayo.v, mayo.v, esk[mayo.skSeedBytes+mayo.oBytes:mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes], true)
	L := decodeMatrixList(mayo.m, mayo.v, mayo.o, esk[mayo.skSeedBytes+mayo.oBytes+mayo.p1Bytes:mayo.eskBytes], false)

	// Hash the message, and derive salt and t
	mDigest := shake256(mayo.digestBytes, m)
	R := make([]byte, mayo.rBytes) // TODO: add randomization?
	salt := shake256(mayo.saltBytes, mDigest, R, seedSk)
	t := decodeVec(mayo.m, shake256(int(math.Ceil(float64(mayo.m)*math.Log2(float64(mayo.q))/8.0)), mDigest, salt))

	// Attempt to find a preimage for t
	var x []byte
	var hasSolution bool
	v := make([][]byte, mayo.k)
	for ctr := 0; ctr < 256; ctr++ {
		// Derive v_i and r
		V := shake256(mayo.k*mayo.vBytes + int(math.Ceil(float64(mayo.k)*float64(mayo.o)*math.Log2(float64(mayo.q))/8)))
		for i := 0; i < mayo.k; i++ {
			v[i] = decodeVec(mayo.n-mayo.o, V[i*mayo.vBytes:(i+1)*mayo.vBytes])
		}
		r := decodeVec(mayo.k*mayo.o, V[mayo.k*mayo.vBytes:mayo.k*mayo.vBytes+int(math.Ceil(float64(mayo.k)*float64(mayo.o)*math.Log2(float64(mayo.q))/8))])

		// Build linear system Ax = y
		A := generateZeroMatrix(mayo.m, mayo.k*mayo.o)
		y := t
		l := 0
		M := make([][][]byte, mayo.k)
		for i := 0; i < mayo.k; i++ {
			mi := generateZeroMatrix(mayo.m, mayo.o)

			for j := 0; j < mayo.m; j++ {
				mi[j] = multiplyMatrices(transposeMatrix(vecToMatrix(v[i])), L[j])[0]
			}

			M[i] = mi
		}

		for i := 0; i < mayo.k; i++ {
			for j := mayo.k - 1; j >= i; j-- {
				u := make([]byte, mayo.m)
				if i == j {
					for a := 0; a < mayo.m; a++ {
						vMatrix := vecToMatrix(v[i])
						u[a] = multiplyMatrices(multiplyMatrices(transposeMatrix(vMatrix), P1[a]), vMatrix)[0][0]
					}
				} else {
					for a := 0; a < mayo.m; a++ {
						viMatrix := vecToMatrix(v[i])
						vjMatrix := vecToMatrix(v[j])
						u[a] = addMatrices(
							multiplyMatrices(multiplyMatrices(transposeMatrix(viMatrix), P1[a]), vjMatrix),
							multiplyMatrices(multiplyMatrices(transposeMatrix(vjMatrix), P1[a]), viMatrix),
						)[0][0] // TODO: Unnecessary computation?
					}
				}

				// TODO: Check how to use l in relation to E^l
				y = subVec(y, multiplyVecConstant(byte(l), u))

				for row := 0; row < mayo.m; row++ {
					for column := i * mayo.o; column < (i+1)*mayo.o; column++ {
						A[row][column] = (A[row][column] + gf16Mul(byte(l), M[j][row][column-i*mayo.o])) % 0x10 // TODO: Is this good enough, also should it be XOR?
					}
				}
				if i != j {
					for row := 0; row < mayo.m; row++ {
						for column := j * mayo.o; column < (j+1)*mayo.o; column++ {
							A[row][column] = (A[row][column] + gf16Mul(byte(l), M[i][row][column-j*mayo.o])) % 0x10 // TODO: Is this good enough, also should it be XOR?
						}
					}
				}

				l++
			}
		}

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
		viOx := transposeMatrix(addMatrices(vecToMatrix(v[i]), multiplyMatrices(O, vecToMatrix(xIndexed))))[0]
		s = append(s, viOx...)
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
	v := mayo.n - mayo.o // TODO: Remove this and use mayo.v
	P1ByteString := epk[:mayo.p1Bytes]
	P2ByteString := epk[mayo.p1Bytes : mayo.p1Bytes+mayo.p2Bytes]
	P3ByteString := epk[mayo.p1Bytes+mayo.p2Bytes : mayo.p1Bytes+mayo.p2Bytes+mayo.p3Bytes]
	P1 := decodeMatrixList(mayo.m, v, v, P1ByteString, true)
	P2 := decodeMatrixList(mayo.m, v, mayo.o, P2ByteString, false)
	P3 := decodeMatrixList(mayo.m, mayo.o, mayo.o, P3ByteString, true)

	// Decode sig
	nkHalf := int(math.Ceil(float64(mayo.n) * float64(mayo.k) / 2.0))
	salt := sig[nkHalf : nkHalf+mayo.saltBytes]
	s := decodeVec(mayo.k*mayo.n, sig)
	si := make([][]byte, mayo.k)
	for i := 0; i < mayo.k; i++ {
		si[i] = make([]byte, mayo.n)
		copy(si[i], s[i*mayo.n:(i+1)*mayo.n])
	}

	// Hash the message and derive t
	mDigest := shake256(mayo.digestBytes, m)
	t := decodeVec(mayo.m, shake256(int(math.Ceil(float64(mayo.m)*math.Log2(float64(mayo.q))/8)), mDigest, salt))

	// Compute P^*(s)
	P := mayo.calculateP(P1, P2, P3) // TODO: Is this calculated correctly?
	/*
		for _, row := range P[0] {
			for _, elem := range row {
				fmt.Printf("%2d ", elem)
			}
			fmt.Println()
		}
	*/

	y := make([]byte, mayo.m)
	l := 0
	for i := 0; i < mayo.k; i++ {
		for j := mayo.k - 1; j >= i; j-- {
			u := make([]byte, mayo.m)
			if i == j {
				for a := 0; a < mayo.m; a++ {
					sMatrix := vecToMatrix(si[i])
					u[a] = multiplyMatrices(multiplyMatrices(transposeMatrix(sMatrix), P[a]), sMatrix)[0][0] % 0x10 // TODO: Reduce mod like this?
				}
			} else {
				for a := 0; a < mayo.m; a++ {
					siMatrix := vecToMatrix(si[i])
					sjMatrix := vecToMatrix(si[j])
					u[a] = addMatrices(
						multiplyMatrices(multiplyMatrices(transposeMatrix(siMatrix), P[a]), sjMatrix),
						multiplyMatrices(multiplyMatrices(transposeMatrix(sjMatrix), P[a]), siMatrix),
					)[0][0]

				}
			}
			// TODO: Check how to use l in relation to E^l
			y = addVectors(y, multiplyVecConstant(byte(l), u))
			l++
		}
	}

	// Accept the signature if y = t
	if bytes.Equal(y, t) {
		return 0
	}
	return -1
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
