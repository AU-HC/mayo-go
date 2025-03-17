package mayo

type CompactSecretKey struct {
	seedSk [skSeedBytes]byte
}

func (csk *CompactSecretKey) Bytes() []byte {
	return csk.seedSk[:]
}

type ExpandedSecretKey struct {
	seedSk [skSeedBytes]byte
	p1     [P1Limbs]uint64
	l      [P2Limbs]uint64
	o      [OBytes]byte
}

func (esk *ExpandedSecretKey) Bytes() []byte {
	var result [eskBytes]byte
	copy(result[:skSeedBytes], esk.seedSk[:])
	copy(result[skSeedBytes:], esk.o[:])
	uint64SliceToBytes(result[skSeedBytes+OBytes:], esk.p1[:])
	uint64SliceToBytes(result[skSeedBytes+OBytes+P1Bytes:], esk.l[:])
	return result[:]
}

type CompactPublicKey struct {
	seedPk [pkSeedBytes]byte
	p3     [P3Limbs]uint64
}

func (cpk *CompactPublicKey) Bytes() []byte {
	var result [cpkBytes]byte
	copy(result[:pkSeedBytes], cpk.seedPk[:])
	packMVecs(cpk.p3[:], result[pkSeedBytes:cpkBytes], P3Limbs/mVecLimbs)

	//uint64SliceToBytes(result[pkSeedBytes:cpkBytes], cpk.p3[:])
	return result[:]
}

type ExpandedPublicKey struct {
	p1 [P1Limbs]uint64
	p2 [P2Limbs]uint64
	p3 [P3Limbs]uint64
}

func (epk *ExpandedPublicKey) Bytes() []byte {
	var result [epkBytes]byte
	uint64SliceToBytes(result[:P1Bytes], epk.p1[:])
	uint64SliceToBytes(result[P1Bytes:P1Bytes+P2Bytes], epk.p2[:])
	uint64SliceToBytes(result[P1Bytes+P2Bytes:P1Bytes+P2Bytes+P3Bytes], epk.p3[:])
	return result[:]
}
