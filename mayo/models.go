package mayo

type CompactSecretKey struct {
	seedSk []byte
}

func (csk *CompactSecretKey) Bytes() []byte {
	return csk.seedSk[:]
}

type ExpandedSecretKey struct {
	seedSk []byte
	p1     []uint64
	l      []uint64
	o      []byte
}

func (cpk *CompactPublicKey) Bytes(mayo *Mayo) []byte {
	result := make([]byte, mayo.cpkBytes)
	copy(result[:mayo.pkSeedBytes], cpk.seedPk[:])
	mayo.packMVecs(cpk.p3[:], result[mayo.pkSeedBytes:mayo.cpkBytes], mayo.P3Limbs/mayo.mVecLimbs)
	return result[:]
}

type CompactPublicKey struct {
	seedPk []byte
	p3     []uint64
}

type ExpandedPublicKey struct {
	p1 []uint64
	p2 []uint64
	p3 []uint64
}
