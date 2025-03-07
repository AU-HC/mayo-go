package mayo

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
