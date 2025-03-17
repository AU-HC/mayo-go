package mayo

import (
	"mayo-go/field"
)

type Mayo struct {
	//tailF []byte
	field *field.Field
}

// InitMayo initializes mayo with the correct parameters according to the specification. Note that
// mayo has 4 levels: 1, 2, 3, and 5.
func InitMayo() *Mayo {
	//return initMayo(86, 78, 8, 10, 16, 24, 32, 16, []byte{8, 1, 1, 0})
	/*if securityLevel == 1 {
		return initMayo(86, 78, 8, 10, 16, 24, 32, 16, []byte{8, 1, 1, 0}), nil
	} else if securityLevel == 2 {
		return initMayo(81, 64, 17, 4, 16, 24, 32, 16, []byte{8, 0, 2, 8}), nil
	} else if securityLevel == 3 {
		return initMayo(118, 108, 10, 11, 16, 32, 48, 16, []byte{8, 0, 1, 7}), nil
	} else if securityLevel == 5 {
		return initMayo(154, 142, 12, 12, 16, 40, 64, 16, []byte{4, 0, 8, 1}), nil
	}

	*/

	return initMayo()
}

// n, m, o, k, q, saltBytes, digestBytes, pkSeedBytes int, tailF []byte
func initMayo() *Mayo {
	/*skSeedBytes := saltBytes
	oBytes := int(math.Ceil(float64((n-o)*o) / 2.0))
	//vBytes := int(math.Ceil(float64(n-o) / 2.0))
	p1Bytes := m * ((n - o) * ((n - o) + 1) / 2) / 2
	p2Bytes := m * (n - o) * o / 2
	p3Bytes := m * ((o + 1) * o / 2) / 2 //
	lBytes := m * (n - o) * o / 2
	eskBytes := skSeedBytes + oBytes + p1Bytes + lBytes
	cpkBytes := pkSeedBytes + p3Bytes
	epkBytes := p1Bytes + p2Bytes + p3Bytes
	sigBytes := int(math.Ceil(float64(n*k)/2.0)) + saltBytes
	//v := n - o

	fmt.Println(fmt.Sprintf("eskBytes: %d", eskBytes))
	fmt.Println(fmt.Sprintf("cpkBytes: %d", cpkBytes))
	fmt.Println(fmt.Sprintf("epkBytes: %d", epkBytes))
	fmt.Println(fmt.Sprintf("sigBytes: %d", sigBytes))
	fmt.Println(fmt.Sprintf("lBytes: %d", lBytes))

	*/

	return &Mayo{
		field: field.InitField(),
	}
}
