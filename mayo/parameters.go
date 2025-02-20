package mayo

import (
	"errors"
	"fmt"
	"math"
)

type Mayo struct {
	// MAYO is parameterized by the following (missing F, which is the polynomial)
	q, m, n, o, k, saltBytes, digestBytes, pkSeedBytes int
	// MAYO then has the following derived parameters
	skSeedBytes, oBytes, vBytes, p1Bytes, p2Bytes, p3Bytes, lBytes, cskBytes, eskBytes, cpkBytes, epkBytes, sigBytes, rBytes int
	E                                                                                                                        [][]byte
	// Lastly we have variables that are not defined in the spec, but help make the code more readable
	v int
}

// InitMayo initializes mayo with the correct parameters according to the specification. Note that
// mayo has 4 levels: 1, 2, 3, and 5.
func InitMayo(securityLevel int) (*Mayo, error) {
	if securityLevel == 1 {
		return initMayo(86, 78, 8, 10, 16, 24, 32, 16), nil
	} else if securityLevel == 2 {
		return initMayo(81, 64, 17, 4, 16, 24, 32, 16), nil
	} else if securityLevel == 3 {
		return initMayo(118, 108, 10, 11, 16, 32, 48, 16), nil
	} else if securityLevel == 5 {
		return initMayo(154, 142, 12, 12, 16, 40, 64, 16), nil
	}

	return nil, errors.New(
		fmt.Sprintf("Wrong security level supplied: '%d'. Must be either '1', '2', '3', or '5'.", securityLevel))
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
	E := make([][]byte, q) // TODO: Generate this multiplication table

	v := n - o

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
		rBytes:      skSeedBytes,
		E:           E,
		v:           v,
	}
}
