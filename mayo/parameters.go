package mayo

import (
	"errors"
	"fmt"
	"math"
	"mayo-go/field"
)

type Mayo struct {
	// MAYO is parameterized by the following (missing F, which is the polynomial)
	q, m, n, o, k, saltBytes, digestBytes, pkSeedBytes int
	tailF                                              []byte

	// MAYO then has the following derived parameters
	skSeedBytes, oBytes, vBytes, p1Bytes, p2Bytes, p3Bytes, lBytes, cskBytes, eskBytes, cpkBytes, epkBytes, sigBytes, rBytes int

	// Lastly we have variables that are not defined in the spec, but help make the code more readable
	v, shifts, mVecLimbs, P1Limbs, P2Limbs, P3Limbs int

	field *field.Field
}

// InitMayo initializes mayo with the correct parameters according to the specification. Note that
// mayo has 4 levels: 1, 2, 3, and 5.
func InitMayo(securityLevel int) (*Mayo, error) {
	if securityLevel == 1 {
		return initMayo(86, 78, 8, 10, 16, 24, 32, 16, []byte{8, 1, 1, 0}), nil
	} else if securityLevel == 2 {
		return initMayo(81, 64, 17, 4, 16, 24, 32, 16, []byte{8, 0, 2, 8}), nil
	} else if securityLevel == 3 {
		return initMayo(118, 108, 10, 11, 16, 32, 48, 16, []byte{8, 0, 1, 7}), nil
	} else if securityLevel == 5 {
		return initMayo(154, 142, 12, 12, 16, 40, 64, 16, []byte{4, 0, 8, 1}), nil
	}

	return nil, errors.New(
		fmt.Sprintf("Wrong security level supplied: '%d'. Must be either '1', '2', '3', or '5'.", securityLevel))
}

func initMayo(n, m, o, k, q, saltBytes, digestBytes, pkSeedBytes int, tailF []byte) *Mayo {
	if q != 16 {
		panic("q is fixed to be 16, in this version of MAYO")
	} else if k >= n-o {
		panic("k should be smaller than n-o")
	}

	v := n - o
	skSeedBytes := saltBytes
	oBytes := int(math.Ceil(float64((n-o)*o) / 2.0))
	vBytes := int(math.Ceil(float64(n-o) / 2.0))
	p1Bytes := m * ((n - o) * ((n - o) + 1) / 2) / 2
	p2Bytes := m * (n - o) * o / 2
	p3Bytes := m * ((o + 1) * o / 2) / 2 //
	lBytes := m * (n - o) * o / 2
	eskBytes := skSeedBytes + oBytes + p1Bytes + lBytes
	cpkBytes := pkSeedBytes + p3Bytes
	epkBytes := p1Bytes + p2Bytes + p3Bytes
	sigBytes := int(math.Ceil(float64(n*k)/2.0)) + saltBytes

	mVecLimbs := (m + 15) / 16
	P1Limbs := v * (v + 1) / 2 * mVecLimbs
	P2Limbs := v * o * mVecLimbs
	P3Limbs := o * (o + 1) / 2 * mVecLimbs

	shifts := k * (k + 1) / 2

	return &Mayo{
		q:           q,
		m:           m,
		n:           n,
		o:           o,
		k:           k,
		saltBytes:   saltBytes,
		digestBytes: digestBytes,
		pkSeedBytes: pkSeedBytes,
		tailF:       tailF,
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
		v:           v,
		shifts:      shifts,
		mVecLimbs:   mVecLimbs,
		P1Limbs:     P1Limbs,
		P2Limbs:     P2Limbs,
		P3Limbs:     P3Limbs,
		field:       field.InitField(),
	}
}
