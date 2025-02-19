package mayo

import "math"

type SecurityLevel int

const (
	LevelOne SecurityLevel = iota
	LevelTwo
	LevelThree
	LevelFive
)

type Mayo struct {
	// MAYO is parameterized by the following (missing F, which is the polynomial)
	q, m, n, o, k, saltBytes, digestBytes, pkSeedBytes int
	// MAYO then has the following derived parameters (missing E, which is a matrix)
	skSeedBytes, oBytes, vBytes, p1Bytes, p2Bytes, p3Bytes, lBytes, cskBytes, eskBytes, cpkBytes, epkBytes, sigBytes int
}

func InitMayo(securityLevel SecurityLevel) *Mayo {
	if securityLevel == LevelOne {
		return initMayo(86, 78, 8, 10, 16, 24, 32, 16)
	} else if securityLevel == LevelTwo {
		return initMayo(81, 64, 17, 4, 16, 24, 32, 16)
	} else if securityLevel == LevelThree {
		return initMayo(118, 108, 10, 11, 16, 32, 48, 16)
	} else { // level five
		return initMayo(154, 142, 12, 12, 16, 40, 64, 16)
	}
}

func initMayo(n, m, o, k, q, saltBytes, digestBytes, pkSeedBytes int) *Mayo {
	if q != 16 {
		panic("q is fixed to be 16, in this version of MAYO")
	} else if k >= n-o {
		panic("k should be smaller than n-o")
	}

	skSeedBytes := saltBytes
	oBytes := int(math.Ceil(float64((n-o)*o) / 2.0))
	vBytes := int(math.Ceil(float64(n-o) / 2.0))
	p1Bytes := m * ((n - o) * ((n - o) + 1) / 2) / 2
	p2Bytes := m * (n - o) * o / 2
	p3Bytes := m * ((o + 1) * o / 2) / 2 // TODO: is this correct?
	lBytes := m * (n - o) * o / 2
	eskBytes := skSeedBytes + oBytes + p1Bytes + lBytes
	cpkBytes := pkSeedBytes + p3Bytes
	epkBytes := p1Bytes + p2Bytes + p3Bytes
	sigBytes := int(math.Ceil(float64(n*k)/2.0)) + saltBytes

	return &Mayo{
		q:           q,
		m:           m,
		n:           n,
		o:           o,
		k:           k,
		saltBytes:   saltBytes,
		digestBytes: digestBytes,
		pkSeedBytes: pkSeedBytes,
		// derived parameters
		skSeedBytes: skSeedBytes,
		oBytes:      oBytes,
		vBytes:      vBytes,
		p1Bytes:     p1Bytes,
		p2Bytes:     p2Bytes,
		p3Bytes:     p3Bytes,
		lBytes:      lBytes,
		cskBytes:    skSeedBytes,
		eskBytes:    eskBytes,
		cpkBytes:    cpkBytes,
		epkBytes:    epkBytes,
		sigBytes:    sigBytes,
	}
}
