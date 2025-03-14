package mayo

import "unsafe"

func (mayo *Mayo) sampleSolution(A, y, r, x []byte) bool {
	// x <- r
	copy(x, r)

	// compute Ar;
	var Ar [M]byte
	for i := 0; i < M; i++ {
		A[K*o+i*(K*o+1)] = 0 // clear last col of A
	}
	mayo.matMul(A, r, Ar[:], K*o+1, M, 1)

	// move y - Ar to last column of matrix A
	for i := 0; i < M; i++ {
		A[K*o+i*(K*o+1)] = y[i] ^ Ar[i]
	}

	mayo.echelonForm(A, M, K*o+1)

	// check if last row of A (excluding the last entry of y) is zero
	var fullRank byte
	for i := 0; i < aCols-1; i++ {
		fullRank |= A[(M-1)*aCols+i]
	}

	if fullRank == 0 {
		return false
	}

	for row := M - 1; row >= 0; row-- {
		var finished byte
		colUpperBound := min(row+(32/(M-row)), K*o)

		for col := row; col <= colUpperBound; col++ {
			correctColumn := mayo.ctCompare8(A[row*aCols+col], 0) & ^finished

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

func (mayo *Mayo) echelonForm(A []byte, nRows int, nCols int) {
	pivotRowData := make([]uint64, (K*o+1+15)/16)
	pivotRowData2 := make([]uint64, (K*o+1+15)/16)
	packedA := make([]uint64, (K*o+1+15)/16*M)

	rowLen := (nCols + 15) / 16

	for i := 0; i < nRows; i++ {
		mayo.efPackMVec(A, i*nCols, packedA, i*rowLen, nCols)
	}

	var inverse byte
	var pivotRow int
	for pivotCol := 0; pivotCol < nCols; pivotCol++ {
		pivotRowLowerBound := max(0, pivotCol+nRows-nCols)
		pivotRowUpperBound := min(nRows-1, pivotCol)

		for i := 0; i < rowLen; i++ { // TODO: Check if needed
			pivotRowData[i] = 0
			pivotRowData2[i] = 0
		}

		var pivot byte
		var pivotIsZero uint64 = 0xffffffffffffffff
		for row := pivotRowLowerBound; row <= min(nRows-1, pivotRowUpperBound+32); row++ {
			isPivotRow := ^mayo.ctCompare(row, pivotRow)
			belowPivotRow := mayo.ctIsGreaterThan(row, pivotRow)

			for j := 0; j < rowLen; j++ {
				mask := isPivotRow | (belowPivotRow & pivotIsZero)
				pivotRowData[j] ^= mask & packedA[row*rowLen+j]
			}

			pivot = extractElement(pivotRowData[:], pivotCol)
			pivotIsZero = ^mayo.ctCompare(int(pivot), 0)
		}

		inverse = mayo.field.Gf16Inv(pivot)
		mayo.vecMulAddUint64(rowLen, pivotRowData, inverse, pivotRowData2, 0)

		for row := pivotRowLowerBound; row <= pivotRowUpperBound; row++ {
			doCopy := ^mayo.ctCompare(row, pivotRow) & ^pivotIsZero
			doNotCopy := ^doCopy
			for col := 0; col < rowLen; col++ {
				packedA[row*rowLen+col] = (doNotCopy & packedA[row*rowLen+col]) +
					(doCopy & pivotRowData2[col])
			}
		}

		for row := pivotRowLowerBound; row < nRows; row++ {
			belowPivot := byte(0)
			if row > pivotRow {
				belowPivot = 1
			}
			eltToElim := extractElement(packedA[row*rowLen:], pivotCol)
			mayo.vecMulAddUint64(rowLen, pivotRowData2, belowPivot*eltToElim, packedA, row*rowLen)
		}

		pivotRow += -int(^pivotIsZero)
	}

	var temp [o*K + 1 + 15]byte
	// unbitslice the matrix A
	for i := 0; i < nRows; i++ {
		efUnpackMVec(rowLen, packedA, i*rowLen, temp[:])
		for j := 0; j < nCols; j++ {
			A[i*nCols+j] = temp[j]
		}
	}
}

func efUnpackMVec(legs int, in []uint64, inStart int, out []byte) {
	inBytes := unsafe.Slice((*byte)(unsafe.Pointer(&in[0])), len(in)*8)
	for i := 0; i < legs*16; i += 2 {
		out[i] = (inBytes[inStart*8+i/2]) & 0xF
		out[i+1] = (inBytes[inStart*8+i/2]) >> 4
	}
}

func (mayo *Mayo) ctCompare8(a, b byte) byte {
	if a == b {
		return 0
	}
	return 0xff
}

func (mayo *Mayo) efPackMVec(in []byte, inStart int, out []uint64, outStart int, nCols int) {
	outBytes := unsafe.Slice((*byte)(unsafe.Pointer(&out[0])), len(out)*8)
	i := 0
	for ; i+1 < nCols; i += 2 {
		outBytes[outStart*8+i/2] = (in[inStart+i] << 0) | (in[inStart+i+1] << 4)
	}

	if nCols%2 == 1 {
		outBytes[outStart*8+i/2] = in[inStart+i] << 0
	}
}

func extractElement(in []uint64, index int) byte {
	leg := index / 16
	offset := index & 15

	return byte((in[leg] >> (offset * 4)) & 0xF)
}

func (mayo *Mayo) ctCompare(a, b int) uint64 { // TODO: Dont use branching here
	if a == b {
		return 0
	}
	return 0xFFFFFFFFFFFFFFFF
}

func (mayo *Mayo) ctIsGreaterThan(a, b int) uint64 { // TODO: Dont use branching here
	if a > b {
		return 0xFFFFFFFFFFFFFFFF
	}
	return 0
}

func (mayo *Mayo) vecMulAddUint64(legs int, in []uint64, a byte, acc []uint64, accStartIdx int) {
	tab := mulTable(a)
	var lsbAsk uint64 = 0x1111111111111111

	for i := 0; i < legs; i++ {
		acc[accStartIdx+i] ^= (in[i]&lsbAsk)*(tab&0xff) ^
			((in[i]>>1)&lsbAsk)*((tab>>8)&0xf) ^
			((in[i]>>2)&lsbAsk)*((tab>>16)&0xf) ^
			((in[i]>>3)&lsbAsk)*((tab>>24)&0xf)
	}
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
