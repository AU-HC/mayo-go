package mayo

import (
	"unsafe"
)

// Note that these optimizations are based on the spec C implementation: https://github.com/PQCMayo/MAYO-C/

func (mayo *Mayo) matMulAdd(bsMat []uint64, mat [][]byte, acc []uint64, bsMatRows, bsMatCols, matCols int, isUpperTriangular int) {
	bsMatEntriesUsed := 0
	mVectorLimbs := (mayo.m + 15) / 16

	for r := 0; r < bsMatRows; r++ {
		for c := r * isUpperTriangular; c < bsMatCols; c++ {
			for k := 0; k < matCols; k++ {
				bsMatStartIndex := bsMatEntriesUsed * mVectorLimbs
				accStartIndex := (r*matCols + k) * mVectorLimbs

				mayo.vecMulAdd(bsMat, bsMatStartIndex, mat[c][k], acc, accStartIndex)
			}
			bsMatEntriesUsed += 1
		}
	}
}

func (mayo *Mayo) mulAddMatTransMat(mat [][]byte, bsMat []uint64, acc []uint64, matRows, matCols, bsMatCols int) {
	mVectorLimbs := (mayo.m + 15) / 16

	for r := 0; r < matCols; r++ {
		for c := 0; c < matRows; c++ {
			for k := 0; k < bsMatCols; k++ {
				bsMatStartIndex := (c*bsMatCols + k) * mVectorLimbs
				accStartIndex := (r*bsMatCols + k) * mVectorLimbs

				mayo.vecMulAdd(bsMat, bsMatStartIndex, mat[c][r], acc, accStartIndex)
			}
		}
	}
}

func (mayo *Mayo) vecAdd(bsMat []uint64, bsMatStartIndex int, acc []uint64, accStartIndex int) {
	mVectorLimbs := (mayo.m + 15) / 16
	for i := 0; i < mVectorLimbs; i++ {
		acc[accStartIndex+i] ^= bsMat[bsMatStartIndex+i]
	}
}

func (mayo *Mayo) vecMulAdd(in []uint64, inputStart int, nibble byte, acc []uint64, accStart int) {
	tab := mulTable(nibble)
	var lsbAsk uint64 = 0x1111111111111111

	mVectorLimbs := (mayo.m + 15) / 16

	for i := 0; i < mVectorLimbs; i++ {
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
	u32PerIndex := (mayo.m + 15) / 16

	for r := 0; r < rows; r++ {
		for c := r; c < cols; c++ {
			for current := 0; current < u32PerIndex; current++ {
				matrixUpper[u32PerIndex*entriesUsed+current] = matrix[u32PerIndex*(r*cols+c)+current]
			}

			if r != c {
				for current := 0; current < u32PerIndex; current++ {
					matrixUpper[u32PerIndex*entriesUsed+current] ^= matrix[u32PerIndex*(c*cols+r)+current]
				}
			}

			entriesUsed += 1
		}
	}
}

func (mayo *Mayo) computeP3(P1 []uint64, O [][]byte, P2 []uint64) []byte {
	// Compute P3 = (−O^T P1 O ) − (−O^T  P2) as P3 = O^t (P1 O + P2)
	// First compute (P1 O + P2) and store in P2
	mayo.matMulAdd(P1, O, P2, mayo.v, mayo.v, mayo.o, 1)
	// Then compute P3 = O^t (P1 O + P2) and store in p3
	P3 := make([]uint64, mayo.o*mayo.o*mayo.m/16)
	mayo.mulAddMatTransMat(O, P2, P3, mayo.v, mayo.o, mayo.o)
	// Compute upper of P3
	P3Upper := make([]uint64, mayo.p3Bytes/8)
	mayo.upper(P3, P3Upper, mayo.v, mayo.o)
	// Serialize P3 to bytes TODO: Consider making a struct for PK and simply storing the uint32's
	P3Bytes := make([]byte, mayo.p3Bytes)
	uint64SliceToBytes(P3Bytes, P3Upper)
	return P3Bytes
}

func (mayo *Mayo) computeL(P1 []uint64, O [][]byte, P2acc []uint64) []byte {
	bsMatEntriesUsed := 0
	mVectorLimbs := (mayo.m + 15) / 16

	for r := 0; r < mayo.v; r++ {
		for c := r; c < mayo.v; c++ {
			if c == r {
				bsMatEntriesUsed += 1
				continue
			}
			bsMatStartIndex := bsMatEntriesUsed * mVectorLimbs
			for k := 0; k < mayo.o; k++ {
				mayo.vecMulAdd(P1, bsMatStartIndex, O[c][k], P2acc, (r*mayo.o+k)*mVectorLimbs)
				mayo.vecMulAdd(P1, bsMatStartIndex, O[r][k], P2acc, (c*mayo.o+k)*mVectorLimbs)
			}
			bsMatEntriesUsed += 1
		}
	}
	// Serialize L to bytes TODO: Consider making a struct for PK and simply storing the uint32's
	lBytes := make([]byte, mayo.lBytes)
	uint64SliceToBytes(lBytes, P2acc)
	return lBytes
}

func (mayo *Mayo) vecMulAddXInv(in []uint64, inStart int, acc []uint64, accStart int) {
	maskLsb := uint64(0x1111111111111111)
	for i := 0; i < 4; i++ { // TODO: fix m vector limbs
		t := in[i+inStart] & maskLsb
		acc[i+accStart] ^= ((in[i+inStart] ^ t) >> 1) ^ (t * 9)
	}
}

func (mayo *Mayo) vecMulAddX(in []uint64, inStart int, acc []uint64, accStart int) {
	maskMsb := uint64(0x8888888888888888)
	for i := 0; i < 4; i++ { // TODO: fix m vector limbs
		t := in[i+inStart] & maskMsb
		acc[i+accStart] ^= ((in[i+inStart] ^ t) << 1) ^ ((t >> 3) * 3)
	}
}

func (mayo *Mayo) vecCopy(in []uint64, inStart int, out []uint64, outStart int) {
	mVectorLimbs := 4 // TODO: fix m vector limbs
	for i := 0; i < mVectorLimbs; i++ {
		out[i+outStart] = in[i+inStart]
	}
}

func (mayo *Mayo) vecMultiplyBins(bins []uint64, binsStartIndex int, out []uint64, outStartIndex int) {
	mVectorLimbs := 4 // TODO: fix
	mayo.vecMulAddXInv(bins, binsStartIndex+5*mVectorLimbs, bins, binsStartIndex+10*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+11*mVectorLimbs, bins, binsStartIndex+12*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+10*mVectorLimbs, bins, binsStartIndex+7*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+12*mVectorLimbs, bins, binsStartIndex+6*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+7*mVectorLimbs, bins, binsStartIndex+14*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+6*mVectorLimbs, bins, binsStartIndex+3*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+14*mVectorLimbs, bins, binsStartIndex+15*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+3*mVectorLimbs, bins, binsStartIndex+8*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+15*mVectorLimbs, bins, binsStartIndex+13*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+8*mVectorLimbs, bins, binsStartIndex+4*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+13*mVectorLimbs, bins, binsStartIndex+9*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+4*mVectorLimbs, bins, binsStartIndex+2*mVectorLimbs)
	mayo.vecMulAddXInv(bins, binsStartIndex+9*mVectorLimbs, bins, binsStartIndex+1*mVectorLimbs)
	mayo.vecMulAddX(bins, binsStartIndex+2*mVectorLimbs, bins, binsStartIndex+1*mVectorLimbs)
	mayo.vecCopy(bins, binsStartIndex+mVectorLimbs, out, outStartIndex)
}

func (mayo *Mayo) calculatePS(P1 []uint64, P2 []uint64, P3 []uint64, s []byte, m int, v int, o int, k int, PS []uint64) {
	n := o + v
	mVectorLimbs := (mayo.m + 15) / 16
	acc := make([]uint64, 16*((m+15)/16)*k*n)

	p1Used := 0
	for row := 0; row < v; row++ {
		for j := row; j < v; j++ {
			for col := 0; col < k; col++ {
				bsMatStartIndex := p1Used * mVectorLimbs
				accStartIndex := ((row*k+col)*16 + int(s[col*n+j])) * mVectorLimbs
				mayo.vecAdd(P1, bsMatStartIndex, acc, accStartIndex)
			}
			p1Used += 1
		}

		for j := 0; j < o; j++ {
			for col := 0; col < k; col++ {
				bsMatStartIndex := (row*o + j) * mVectorLimbs
				accStartIndex := ((row*k+col)*16 + int(s[(col*n)+j+v])) * mVectorLimbs
				mayo.vecAdd(P2, bsMatStartIndex, acc, accStartIndex)
			}
		}
	}

	p3Used := 0
	for row := v; row < n; row++ {
		for j := row; j < n; j++ {
			for col := 0; col < k; col++ {
				bsMatStartIndex := p3Used * mVectorLimbs
				accStartIndex := ((row*k+col)*16 + int(s[col*n+j])) * mVectorLimbs
				mayo.vecAdd(P3, bsMatStartIndex, acc, accStartIndex)
			}
			p3Used += 1
		}
	}

	for i := 0; i < n*k; i++ {
		bsMatStartIndex := i * mVectorLimbs
		accStartIndex := i * 16 * mVectorLimbs
		mayo.vecMultiplyBins(acc, accStartIndex, PS, bsMatStartIndex)
	}
}

func (mayo *Mayo) calculateSPS(PS []uint64, s []byte, m int, k int, n int, SPS []uint64) {
	mVectorLimbs := 4 // TODO: 4 = mVectorLimbs
	acc := make([]uint64, 16*((m+15)/16)*k*k)

	for row := 0; row < k; row++ {
		for j := 0; j < n; j++ {
			for col := 0; col < k; col++ {
				bsMatStartIndex := (j*k + col) * mVectorLimbs
				accStartIndex := ((row*k+col)*16 + int(s[row*n+j])) * mVectorLimbs
				mayo.vecAdd(PS, bsMatStartIndex, acc, accStartIndex)
			}
		}
	}

	for i := 0; i < k*k; i++ {
		bsMatStartIndex := i * mVectorLimbs
		accStartIndex := i * 16 * mVectorLimbs
		mayo.vecMultiplyBins(acc, accStartIndex, SPS, bsMatStartIndex)
	}
}

func (mayo *Mayo) calculatePsSps(P1 []uint64, P2 []uint64, P3 []uint64, s []byte, SPS []uint64) {
	PS := make([]uint64, mayo.n*mayo.k*4) // TODO: 4 = mVectorLimbs
	mayo.calculatePS(P1, P2, P3, s, mayo.m, mayo.v, mayo.o, mayo.k, PS)
	mayo.calculateSPS(PS, s, mayo.m, mayo.k, mayo.n, SPS)
}

func (mayo *Mayo) computeRhs(VPV []uint64, t, y []byte) {
	topPos := ((mayo.m - 1) % 16) * 4
	mVectorLimbs := 4 // TODO: 4 = mVectorLimbs

	// TODO: zero out fails of m_vectors if necessary (not needed for mayo2 as 64 % 16 == 0)
	// here
	// here
	// here

	temp := make([]uint64, mVectorLimbs)
	tempBytes := unsafe.Slice((*byte)(unsafe.Pointer(&temp[0])), len(temp)*8)
	for i := mayo.k - 1; i >= 0; i-- {
		for j := i; j < mayo.k; j++ {
			// multiply
			top := byte((temp[mVectorLimbs-1] >> topPos) % 16)
			temp[mVectorLimbs-1] <<= 4
			for k := mVectorLimbs - 2; k >= 0; k-- {
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
			for k := 0; k < mVectorLimbs; k++ {
				var ij uint64
				if i != j {
					ij = 1
				}

				temp[k] ^= VPV[(i*mayo.k+j)*mVectorLimbs+k] ^ ((ij) * VPV[(j*mayo.k+i)*mVectorLimbs+k])
			}
		}
	}

	// compute y
	for i := 0; i < mayo.m; i += 2 {
		y[i] = t[i] ^ (tempBytes[i/2] & 0xF)
		y[i+1] = t[i+1] ^ (tempBytes[i/2] >> 4)
	}
}

func (mayo *Mayo) evalPublicMap(s []byte, P1 []uint64, P2 []uint64, P3 []uint64, eval []byte) {
	SPS := make([]uint64, mayo.k*mayo.k*4) // TODO: 4 = mVectorLimbs
	mayo.calculatePsSps(P1, P2, P3, s, SPS)
	zero := make([]byte, mayo.m)
	mayo.computeRhs(SPS, zero, eval)
}

func (mayo *Mayo) mulAddMatXMMat(v []byte, L []uint64, acc []uint64, matRows, matCols, bsMatCols int) {
	for r := 0; r < matRows; r++ {
		for c := 0; c < matCols; c++ {
			for k := 0; k < bsMatCols; k++ {
				mayo.vecMulAdd(L, 4*(c*bsMatCols+k), v[r*matCols+c], acc, 4*(r*bsMatCols+k)) // TODO: mVectorLimbs = 4 also are we indexing correct in v
			}
		}
	}
}

func (mayo *Mayo) P1MulVt(P1 []uint64, v []byte, Pv []uint64) {
	bsMatEntriesUsed := 0
	for r := 0; r < mayo.v; r++ {
		for c := 1 * r; c < mayo.v; c++ {
			for k := 0; k < mayo.k; k++ {
				mayo.vecMulAdd(P1, 4*bsMatEntriesUsed, v[k*mayo.v+c], Pv, 4*(r*mayo.k+k)) // TODO: mVectorLimbs = 4 also are we indexing correct in v
			}
			bsMatEntriesUsed++
		}
	}
}

func (mayo *Mayo) computeMAndVpv(v []byte, L, P1, VL, A []uint64) {
	// Compute VL
	mayo.mulAddMatXMMat(v, L, VL, mayo.k, mayo.v, mayo.o)

	// Compute VP1V
	Pv := make([]uint64, mayo.v*mayo.k*4) // TODO: 4 = mVectorLimbs
	mayo.P1MulVt(P1, v, Pv)
	mayo.mulAddMatXMMat(v, Pv, A, mayo.k, mayo.v, mayo.k) // TODO: Cast A to uint64* type
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
	mayoMOver8 := (mayo.m + 7) / 8
	bitsToShift := 0
	wordsToShift := 0
	AWidth := ((mayo.o*mayo.k + 15) / 16) * 16
	A := make([]uint64, (((mayo.o*mayo.k+15)/16)*16)*mayoMOver8)

	// TODO: zero out fails of m_vectors if necessary (not needed for mayo2 as 64 % 16 == 0)
	// here
	// here
	// here

	for i := 0; i < mayo.k; i++ {
		for j := mayo.k - 1; j >= i; j-- {
			for c := 0; c < mayo.o; c++ {
				for k := 0; k < 4; k++ { //TODO: mVectorLimbs = 4
					A[mayo.o*i+c+(k+wordsToShift)*AWidth] ^= mTemp[j*4*mayo.o+k+c*4] << bitsToShift // TODO: mVectorLimbs = 4
					if bitsToShift > 0 {
						A[mayo.o*i+c+(k+wordsToShift+1)*AWidth] ^= mTemp[j*4*mayo.o+k+c*4] >> (64 - bitsToShift) // TODO: mVectorLimbs = 4
					}
				}
			}

			if i != j {
				for c := 0; c < mayo.o; c++ {
					for k := 0; k < 4; k++ { //TODO: mVectorLimbs = 4
						A[mayo.o*j+c+(k+wordsToShift)*AWidth] ^= mTemp[i*4*mayo.o+k+c*4] << bitsToShift //TODO: mVectorLimbs = 4
						if bitsToShift > 0 {
							A[mayo.o*j+c+(k+wordsToShift+1)*AWidth] ^= mTemp[i*4*mayo.o+k+c*4] >> (64 - bitsToShift) //TODO: mVectorLimbs = 4
						}
					}
				}
			}

			bitsToShift += 4
			if bitsToShift == 64 {
				bitsToShift = 0
				wordsToShift++
			}
		}
	}

	for c := 0; c < AWidth*((mayo.m+(mayo.k+1)*mayo.k/2+15)/16); c += 16 {
		mayo.Transpose16x16Nibbles(A, c)
	}

	tab := make([]byte, len(mayo.tailF)*4)
	for i := 0; i < len(mayo.tailF); i++ {
		tab[4*i] = mayo.field.Gf16Mul(mayo.tailF[i], 1)
		tab[4*i+1] = mayo.field.Gf16Mul(mayo.tailF[i], 2)
		tab[4*i+2] = mayo.field.Gf16Mul(mayo.tailF[i], 4)
		tab[4*i+3] = mayo.field.Gf16Mul(mayo.tailF[i], 8)
	}

	lowBitInNibble := uint64(0x1111111111111111)
	for c := 0; c < AWidth; c += 16 {
		for r := mayo.m; r < mayo.m+(mayo.k+1)*mayo.k/2; r++ {
			pos := (r/16)*AWidth + c + (r % 16)
			t0 := A[pos] & lowBitInNibble
			t1 := (A[pos] >> 1) & lowBitInNibble
			t2 := (A[pos] >> 2) & lowBitInNibble
			t3 := (A[pos] >> 3) & lowBitInNibble
			for t := 0; t < len(mayo.tailF); t++ {
				A[((r+t-mayo.m)/16)*AWidth+c+((r+t-mayo.m)%16)] ^= t0*uint64(tab[4*t+0]) ^ t1*uint64(tab[4*t+1]) ^ t2*uint64(tab[4*t+2]) ^ t3*uint64(tab[4*t+3])
			}
		}
	}
	/*
		aCols := mayo.k*mayo.o + 1
		aBytes := make([]byte, len(A)*8)
		uint64SliceToBytes(aBytes, A)
		for r := 0; r < mayo.m; r += 16 {
			for c := 0; c < aCols-1; c += 16 {
				for i := 0; i+r < mayo.m; i++ {
					col := decodeVec(int(math.Min(16, float64(aCols-1-c))), aBytes[r*AWidth/16+c+i:])
					copy(AOut[aCols*(r+i)+c:aCols*(r+(i+1))+c], col[:]) // TODO Check this
				}
			}
		}

	*/

	aBytes := make([]byte, len(A)*8)
	uint64SliceToBytes(aBytes[:], A[:])

	OKpadded := (mayo.k*mayo.o + 15) / 16 * 16
	KO1 := mayo.k*mayo.o + 1
	for r := 0; r < mayo.m; r += 16 {
		for c := 0; c < KO1-1; c += 16 {
			for i := 0; i < 16; i++ {
				src := aBytes[(r/16*OKpadded+c+i)*8:]
				offset := KO1*(r+i) + c
				decoded := decodeVec(len(src), src) // TODO: Fix
				copy(AOut[offset:offset+min(16, KO1-1-c)], decoded)
			}
		}
	}
}
