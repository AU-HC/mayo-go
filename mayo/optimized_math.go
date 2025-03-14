package mayo

import (
	"unsafe"
)

// Note that these optimizations are based on the spec C implementation: https://github.com/PQCMayo/MAYO-C/

func (mayo *Mayo) matMulAdd(bsMat []uint64, mat [][]byte, acc []uint64, bsMatRows, bsMatCols, matCols int, isUpperTriangular int) {
	bsMatEntriesUsed := 0

	for r := 0; r < bsMatRows; r++ {
		for c := r * isUpperTriangular; c < bsMatCols; c++ {
			for k := 0; k < matCols; k++ {
				bsMatStartIndex := bsMatEntriesUsed * mVecLimbs
				accStartIndex := (r*matCols + k) * mVecLimbs

				mayo.vecMulAdd(bsMat, bsMatStartIndex, mat[c][k], acc, accStartIndex)
			}
			bsMatEntriesUsed += 1
		}
	}
}

func (mayo *Mayo) mulAddMatTransMat(mat [][]byte, bsMat []uint64, acc []uint64, matRows, matCols, bsMatCols int) {
	for r := 0; r < matCols; r++ {
		for c := 0; c < matRows; c++ {
			for k := 0; k < bsMatCols; k++ {
				bsMatStartIndex := (c*bsMatCols + k) * mVecLimbs
				accStartIndex := (r*bsMatCols + k) * mVecLimbs

				mayo.vecMulAdd(bsMat, bsMatStartIndex, mat[c][r], acc, accStartIndex)
			}
		}
	}
}

func (mayo *Mayo) vecAdd(bsMat []uint64, bsMatStartIndex int, acc []uint64, accStartIndex int) {
	for i := 0; i < mVecLimbs; i++ {
		acc[accStartIndex+i] ^= bsMat[bsMatStartIndex+i]
	}
}

func (mayo *Mayo) vecMulAdd(in []uint64, inputStart int, nibble byte, acc []uint64, accStart int) {
	tab := mulTable(nibble)
	var lsbAsk uint64 = 0x1111111111111111

	for i := 0; i < mVecLimbs; i++ {
		acc[accStart+i] ^= (in[i+inputStart]&lsbAsk)*(tab&0xff) ^
			((in[i+inputStart]>>1)&lsbAsk)*((tab>>8)&0xf) ^
			((in[i+inputStart]>>2)&lsbAsk)*((tab>>16)&0xf) ^
			((in[i+inputStart]>>3)&lsbAsk)*((tab>>24)&0xf)
	}
}

func mulTable(b byte) uint64 {
	x := uint64(b) * 0x08040201

	highNibbleMask := uint64(0xf0f0f0f0)

	highHalf := x & highNibbleMask
	return x ^ (highHalf >> 4) ^ (highHalf >> 3)
}

func (mayo *Mayo) lincomb(a, b []byte, n, m int) byte {
	var ret byte = 0
	for i := 0; i < n; i++ {
		if i >= len(a) || i*m >= len(b) {
			continue // count as XOR'ing with zero
		}
		ret = mayo.field.Gf16Mul(a[i], b[i*m]) ^ ret
	}
	return ret
}

func (mayo *Mayo) matMul(a, b, c []byte, colrowAB, rowA, colB int) {
	for i := 0; i < rowA; i++ {
		aOffset := i * colrowAB
		for j := 0; j < colB; j++ {
			c[i*colB+j] = mayo.lincomb(a[aOffset:], b[j:], colrowAB, colB)
		}
	}
}

func (mayo *Mayo) matAdd(a, b, c []byte, cStartIdx, m, n int) {
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			c[cStartIdx+i*n+j] = (a[i*n+j]) ^ (b[i*n+j])
		}
	}
}

func (mayo *Mayo) upper(matrix []uint64, matrixUpper []uint64, rows, cols int) {
	entriesUsed := 0

	for r := 0; r < rows; r++ {
		for c := r; c < cols; c++ {
			for current := 0; current < mVecLimbs; current++ {
				matrixUpper[mVecLimbs*entriesUsed+current] = matrix[mVecLimbs*(r*cols+c)+current]
			}

			if r != c {
				for current := 0; current < mVecLimbs; current++ {
					matrixUpper[mVecLimbs*entriesUsed+current] ^= matrix[mVecLimbs*(c*cols+r)+current]
				}
			}

			entriesUsed += 1
		}
	}
}

func (mayo *Mayo) computeP3(P1 []uint64, O [][]byte, P2 []uint64) []byte {
	// Compute P3 = (−O^T P1 O ) − (−O^T  P2) as P3 = O^t (P1 O + P2)
	// First compute (P1 O + P2) and store in P2
	mayo.matMulAdd(P1, O, P2, v, v, o, 1)
	// Then compute P3 = O^t (P1 O + P2) and store in p3
	var P3 [o * o * M / 16]uint64
	mayo.mulAddMatTransMat(O, P2, P3[:], v, o, o)
	// Compute upper of P3
	var P3Upper [P3Bytes / 8]uint64
	mayo.upper(P3[:], P3Upper[:], v, o)
	// Serialize P3 to bytes TODO: Consider making a struct for PK and simply storing the uint32's
	P3ByteArray := make([]byte, P3Bytes)
	uint64SliceToBytes(P3ByteArray, P3Upper[:])
	return P3ByteArray
}

func (mayo *Mayo) computeL(P1 []uint64, O [][]byte, P2acc []uint64) []byte {
	bsMatEntriesUsed := 0

	for r := 0; r < v; r++ {
		for c := r; c < v; c++ {
			if c == r {
				bsMatEntriesUsed += 1
				continue
			}
			bsMatStartIndex := bsMatEntriesUsed * mVecLimbs
			for k := 0; k < o; k++ {
				mayo.vecMulAdd(P1, bsMatStartIndex, O[c][k], P2acc, (r*o+k)*mVecLimbs)
				mayo.vecMulAdd(P1, bsMatStartIndex, O[r][k], P2acc, (c*o+k)*mVecLimbs)
			}
			bsMatEntriesUsed += 1
		}
	}
	// Serialize L to bytes TODO: Consider making a struct for PK and simply storing the uint32's
	var lBytesArray [lBytes]byte
	uint64SliceToBytes(lBytesArray[:], P2acc)
	return lBytesArray[:]
}

func (mayo *Mayo) vecMulAddXInv(in []uint64, inStart int, acc []uint64, accStart int) {
	maskLsb := uint64(0x1111111111111111)
	for i := 0; i < mVecLimbs; i++ {
		t := in[i+inStart] & maskLsb
		acc[i+accStart] ^= ((in[i+inStart] ^ t) >> 1) ^ (t * 9)
	}
}

func (mayo *Mayo) vecMulAddX(in []uint64, inStart int, acc []uint64, accStart int) {
	maskMsb := uint64(0x8888888888888888)
	for i := 0; i < mVecLimbs; i++ {
		t := in[i+inStart] & maskMsb
		acc[i+accStart] ^= ((in[i+inStart] ^ t) << 1) ^ ((t >> 3) * 3)
	}
}

func (mayo *Mayo) vecCopy(in []uint64, inStart int, out []uint64, outStart int) {
	for i := 0; i < mVecLimbs; i++ {
		out[i+outStart] = in[i+inStart]
	}
}

func (mayo *Mayo) vecMultiplyBins(bins []uint64, binsStartIndex int, out []uint64, outStartIndex int) {
	mayo.vecMulAddXInv(bins, binsStartIndex+5*mVecLimbs, bins, binsStartIndex+10*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+11*mVecLimbs, bins, binsStartIndex+12*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+10*mVecLimbs, bins, binsStartIndex+7*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+12*mVecLimbs, bins, binsStartIndex+6*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+7*mVecLimbs, bins, binsStartIndex+14*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+6*mVecLimbs, bins, binsStartIndex+3*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+14*mVecLimbs, bins, binsStartIndex+15*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+3*mVecLimbs, bins, binsStartIndex+8*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+15*mVecLimbs, bins, binsStartIndex+13*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+8*mVecLimbs, bins, binsStartIndex+4*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+13*mVecLimbs, bins, binsStartIndex+9*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+4*mVecLimbs, bins, binsStartIndex+2*mVecLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+9*mVecLimbs, bins, binsStartIndex+1*mVecLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+2*mVecLimbs, bins, binsStartIndex+1*mVecLimbs)
	mayo.vecCopy(bins, binsStartIndex+mVecLimbs, out, outStartIndex)
}

func (mayo *Mayo) calculatePS(P1 []uint64, P2 []uint64, P3 []uint64, s []byte, PS []uint64) {
	acc := make([]uint64, 16*((M+15)/16)*K*N)

	p1Used := 0
	for row := 0; row < v; row++ {
		for j := row; j < v; j++ {
			for col := 0; col < K; col++ {
				bsMatStartIndex := p1Used * mVecLimbs
				accStartIndex := ((row*K+col)*16 + int(s[col*N+j])) * mVecLimbs
				mayo.vecAdd(P1, bsMatStartIndex, acc, accStartIndex)
			}
			p1Used += 1
		}

		for j := 0; j < o; j++ {
			for col := 0; col < K; col++ {
				bsMatStartIndex := (row*o + j) * mVecLimbs
				accStartIndex := ((row*K+col)*16 + int(s[(col*N)+j+v])) * mVecLimbs
				mayo.vecAdd(P2, bsMatStartIndex, acc, accStartIndex)
			}
		}
	}

	p3Used := 0
	for row := v; row < N; row++ {
		for j := row; j < N; j++ {
			for col := 0; col < K; col++ {
				bsMatStartIndex := p3Used * mVecLimbs
				accStartIndex := ((row*K+col)*16 + int(s[col*N+j])) * mVecLimbs
				mayo.vecAdd(P3, bsMatStartIndex, acc, accStartIndex)
			}
			p3Used += 1
		}
	}

	for i := 0; i < N*K; i++ {
		bsMatStartIndex := i * mVecLimbs
		accStartIndex := i * 16 * mVecLimbs
		mayo.vecMultiplyBins(acc, accStartIndex, PS, bsMatStartIndex)
	}
}

func (mayo *Mayo) calculateSPS(PS []uint64, s []byte, SPS []uint64) {
	var acc [16 * ((M + 15) / 16) * K * K]uint64

	for row := 0; row < K; row++ {
		for j := 0; j < N; j++ {
			for col := 0; col < K; col++ {
				bsMatStartIndex := (j*K + col) * mVecLimbs
				accStartIndex := ((row*K+col)*16 + int(s[row*N+j])) * mVecLimbs
				mayo.vecAdd(PS, bsMatStartIndex, acc[:], accStartIndex)
			}
		}
	}

	for i := 0; i < K*K; i++ {
		bsMatStartIndex := i * mVecLimbs
		accStartIndex := i * 16 * mVecLimbs
		mayo.vecMultiplyBins(acc[:], accStartIndex, SPS, bsMatStartIndex)
	}
}

func (mayo *Mayo) calculatePsSps(P1 []uint64, P2 []uint64, P3 []uint64, s []byte, SPS []uint64) {
	var PS [N * K * 4]uint64
	mayo.calculatePS(P1, P2, P3, s, PS[:])
	mayo.calculateSPS(PS[:], s, SPS)
}

func (mayo *Mayo) computeRhs(VPV []uint64, t, y []byte) {
	topPos := ((M - 1) % 16) * 4

	// TODO: zero out fails of m_vectors if necessary (not needed for mayo2 as 64 % 16 == 0)
	// here
	// here
	// here

	temp := make([]uint64, mVecLimbs)
	tempBytes := unsafe.Slice((*byte)(unsafe.Pointer(&temp[0])), len(temp)*8)
	for i := K - 1; i >= 0; i-- {
		for j := i; j < K; j++ {
			// multiply
			top := byte((temp[mVecLimbs-1] >> topPos) % 16)
			temp[mVecLimbs-1] <<= 4
			for k := mVecLimbs - 2; k >= 0; k-- {
				temp[k+1] ^= temp[k] >> 60
				temp[k] <<= 4
			}

			// reduce
			for jj := 0; jj < len(mayo.tailF); jj++ {
				if jj%2 == 0 {
					tempBytes[jj/2] ^= mayo.field.Gf16Mul(top, mayo.tailF[jj])
				} else {
					tempBytes[jj/2] ^= mayo.field.Gf16Mul(top, mayo.tailF[jj]) << 4
				}
			}

			// extract
			for k := 0; k < mVecLimbs; k++ {
				var ij uint64
				if i != j {
					ij = 1
				}

				temp[k] ^= VPV[(i*K+j)*mVecLimbs+k] ^ ((ij) * VPV[(j*K+i)*mVecLimbs+k])
			}
		}
	}

	// compute y
	for i := 0; i < M; i += 2 {
		y[i] = t[i] ^ (tempBytes[i/2] & 0xF)
		y[i+1] = t[i+1] ^ (tempBytes[i/2] >> 4)
	}
}

func (mayo *Mayo) evalPublicMap(s []byte, P1 []uint64, P2 []uint64, P3 []uint64, eval []byte) {
	var SPS [K * K * mVecLimbs]uint64
	mayo.calculatePsSps(P1, P2, P3, s, SPS[:])
	zero := make([]byte, M)
	mayo.computeRhs(SPS[:], zero, eval)
}

func (mayo *Mayo) mulAddMatXMMat(v []byte, L []uint64, acc []uint64, matRows, matCols, bsMatCols int) {
	for r := 0; r < matRows; r++ {
		for c := 0; c < matCols; c++ {
			for k := 0; k < bsMatCols; k++ {
				mayo.vecMulAdd(L, mVecLimbs*(c*bsMatCols+k), v[r*matCols+c], acc, mVecLimbs*(r*bsMatCols+k))
			}
		}
	}
}

func (mayo *Mayo) P1MulVt(P1 []uint64, vDec []byte, Pv []uint64) {
	bsMatEntriesUsed := 0
	for r := 0; r < v; r++ {
		for c := 1 * r; c < v; c++ {
			for k := 0; k < K; k++ {
				mayo.vecMulAdd(P1, mVecLimbs*bsMatEntriesUsed, vDec[k*v+c], Pv, mVecLimbs*(r*K+k))
			}
			bsMatEntriesUsed++
		}
	}
}

func (mayo *Mayo) computeMAndVpv(vDec []byte, L, P1, VL, A []uint64) {
	// Compute VL
	mayo.mulAddMatXMMat(vDec, L, VL, K, v, o)

	// Compute VP1V
	var Pv [v * K * mVecLimbs]uint64
	mayo.P1MulVt(P1, vDec, Pv[:])
	mayo.mulAddMatXMMat(vDec, Pv[:], A, K, v, K)
}

func (mayo *Mayo) Transpose16x16Nibbles(M []uint64, c int) {
	evenNibbles := uint64(0x0f0f0f0f0f0f0f0f)
	evenBytes := uint64(0x00ff00ff00ff00ff)
	even2bytes := uint64(0x0000ffff0000ffff)
	evenHalf := uint64(0x00000000ffffffff)

	for i := 0; i < 16; i += 2 {
		t := ((M[c+i] >> 4) ^ M[c+i+1]) & evenNibbles
		M[c+i] ^= t << 4
		M[c+i+1] ^= t
	}

	for i := 0; i < 16; i += 4 {
		t0 := ((M[c+i] >> 8) ^ M[c+i+2]) & evenBytes
		t1 := ((M[c+i+1] >> 8) ^ M[c+i+3]) & evenBytes
		M[c+i] ^= t0 << 8
		M[c+i+1] ^= t1 << 8
		M[c+i+2] ^= t0
		M[c+i+3] ^= t1
	}

	for i := 0; i < 4; i++ {
		t0 := ((M[c+i] >> 16) ^ M[c+i+4]) & even2bytes
		t1 := ((M[c+i+8] >> 16) ^ M[c+i+12]) & even2bytes

		M[c+i] ^= t0 << 16
		M[c+i+8] ^= t1 << 16
		M[c+i+4] ^= t0
		M[c+i+12] ^= t1
	}

	for i := 0; i < 8; i++ {
		t := ((M[c+i] >> 32) ^ M[c+i+8]) & evenHalf
		M[c+i] ^= t << 32
		M[c+i+8] ^= t
	}
}

func (mayo *Mayo) computeA(mTemp []uint64, AOut []byte) {
	mayoMOver8 := (M + 7) / 8
	bitsToShift := 0
	wordsToShift := 0
	AWidth := ((o*K + 15) / 16) * 16
	A := make([]uint64, (((o*K+15)/16)*16)*mayoMOver8)

	// TODO: zero out fails of m_vectors if necessary (not needed for mayo2 as 64 % 16 == 0)
	// here
	// here
	// here

	for i := 0; i < K; i++ {
		for j := K - 1; j >= i; j-- {
			for c := 0; c < o; c++ {
				for k := 0; k < mVecLimbs; k++ {
					A[o*i+c+(k+wordsToShift)*AWidth] ^= mTemp[j*mVecLimbs*o+k+c*mVecLimbs] << bitsToShift
					if bitsToShift > 0 {
						A[o*i+c+(k+wordsToShift+1)*AWidth] ^= mTemp[j*mVecLimbs*o+k+c*mVecLimbs] >> (64 - bitsToShift)
					}
				}
			}

			if i != j {
				for c := 0; c < o; c++ {
					for k := 0; k < mVecLimbs; k++ {
						A[o*j+c+(k+wordsToShift)*AWidth] ^= mTemp[i*mVecLimbs*o+k+c*mVecLimbs] << bitsToShift
						if bitsToShift > 0 {
							A[o*j+c+(k+wordsToShift+1)*AWidth] ^= mTemp[i*mVecLimbs*o+k+c*mVecLimbs] >> (64 - bitsToShift)
						}
					}
				}
			}

			bitsToShift += 4 // TODO is this mVectorLimbs
			if bitsToShift == 64 {
				bitsToShift = 0
				wordsToShift++
			}
		}
	}

	for c := 0; c < AWidth*((M+(K+1)*K/2+15)/16); c += 16 {
		mayo.Transpose16x16Nibbles(A, c)
	}

	tab := make([]byte, len(mayo.tailF)*4) // TODO: is this mVecLimbs
	for i := 0; i < len(mayo.tailF); i++ {
		tab[4*i] = mayo.field.Gf16Mul(mayo.tailF[i], 1)
		tab[4*i+1] = mayo.field.Gf16Mul(mayo.tailF[i], 2)
		tab[4*i+2] = mayo.field.Gf16Mul(mayo.tailF[i], 4)
		tab[4*i+3] = mayo.field.Gf16Mul(mayo.tailF[i], 8)
	}

	lowBitInNibble := uint64(0x1111111111111111)
	for c := 0; c < AWidth; c += 16 {
		for r := M; r < M+(K+1)*K/2; r++ {
			pos := (r/16)*AWidth + c + (r % 16)
			t0 := A[pos] & lowBitInNibble
			t1 := (A[pos] >> 1) & lowBitInNibble
			t2 := (A[pos] >> 2) & lowBitInNibble
			t3 := (A[pos] >> 3) & lowBitInNibble
			for t := 0; t < len(mayo.tailF); t++ {
				A[((r+t-M)/16)*AWidth+c+((r+t-M)%16)] ^= t0*uint64(tab[4*t+0]) ^ t1*uint64(tab[4*t+1]) ^ t2*uint64(tab[4*t+2]) ^ t3*uint64(tab[4*t+3])
			}
		}
	}

	aBytes := make([]byte, len(A)*8)
	uint64SliceToBytes(aBytes[:], A[:])

	OKpadded := (K*o + 15) / 16 * 16
	KO1 := K*o + 1
	for r := 0; r < M; r += 16 {
		for c := 0; c < KO1-1; c += 16 {
			for i := 0; i < 16; i++ {
				src := aBytes[(r/16*OKpadded+c+i)*8:]
				offset := KO1*(r+i) + c
				decoded := decodeVec(len(src), src)
				copy(AOut[offset:offset+min(16, KO1-1-c)], decoded)
			}
		}
	}
}
