package mayo

const W int = 32 / 4

func (mayo *Mayo) matMulAdd(bsMat []uint32, mat [][]byte, acc []uint32, bsMatRows, bsMatCols, matCols int, isUpperTriangular int) {
	bsMatEntriesUsed := 0
	for r := 0; r < bsMatRows; r++ {
		for c := r * isUpperTriangular; c < bsMatCols; c++ {
			for k := 0; k < matCols; k++ {
				bsMatStartIndex := bsMatEntriesUsed * (mayo.m / 32) * 4
				accStartIndex := (r*matCols + k) * (mayo.m / 32) * 4

				vecMulAdd(bsMat, bsMatStartIndex, mat[c][k], acc, accStartIndex)
			}
			bsMatEntriesUsed += 1
		}
	}
}

func (mayo *Mayo) mulAddMatTransMat(mat [][]byte, bsMat []uint32, acc []uint32, matRows, matCols, bsMatCols int) {
	for r := 0; r < matCols; r++ {
		for c := 0; c < matRows; c++ {
			for k := 0; k < bsMatCols; k++ {
				bsMatStartIndex := (c*bsMatCols + k) * mayo.m / 32 * 4
				accStartIndex := (r*bsMatCols + k) * mayo.m / 32 * 4

				vecMulAdd(bsMat, bsMatStartIndex, mat[c][r], acc, accStartIndex)
			}
		}
	}
}

func vecMulAdd(in []uint32, inputStart int, nibble byte, acc []uint32, accStart int) {
	tab := mulTable(nibble)
	var lsbAsk uint32 = 0x11111111 //11111111

	for i := 0; i < W; i++ {
		acc[accStart+i] ^= (in[i+inputStart]&lsbAsk)*(tab&0xff) ^
			((in[i+inputStart]>>1)&lsbAsk)*((tab>>8)&0xf) ^
			((in[i+inputStart]>>2)&lsbAsk)*((tab>>16)&0xf) ^
			((in[i+inputStart]>>3)&lsbAsk)*((tab>>24)&0xf)
	}
}

func mulTable(b byte) uint32 {
	x := uint32(b) * 0x08040201

	highNibbleMask := uint32(0xf0f0f0f0)

	highHalf := x & highNibbleMask
	return x ^ (highHalf >> 4) ^ (highHalf >> 3)
}

func (mayo *Mayo) upper(matrix []uint32, matrixUpper []uint32, rows, cols int) {
	entriesUsed := 0
	u32PerIndex := mayo.m / 32 * 4

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

func (mayo *Mayo) computeP3(P1 []uint32, O [][]byte, P2 []uint32) []byte {
	// Compute P3 = (−O^T P1 O ) − (−O^T  P2) as P3 = O^t (P1 O + P2)
	// First compute (P1 O + P2) and store in P2
	mayo.matMulAdd(P1, O, P2, mayo.v, mayo.v, mayo.o, 1)
	// Then compute P3 = O^t (P1 O + P2) and store in p3
	P3 := make([]uint32, mayo.o*mayo.o*mayo.m/8)
	mayo.mulAddMatTransMat(O, P2, P3, mayo.v, mayo.o, mayo.o)
	// Compute upper of P3
	P3Upper := make([]uint32, mayo.p3Bytes/4)
	mayo.upper(P3, P3Upper, mayo.v, mayo.o)
	// Serialize P3 to bytes TODO: Consider making a struct for PK and simply storing the uint32's
	P3Bytes := make([]byte, mayo.p3Bytes)
	uint32SliceToBytes(P3Bytes, P3Upper)
	return P3Bytes
}
