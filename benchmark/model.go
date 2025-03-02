package benchmark

type Results struct {
	KeyGen, ExpandSK, ExpandPK, Sign, Verify []int64
}
