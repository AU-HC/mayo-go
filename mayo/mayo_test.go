package mayo

import "testing"

func TestTheTest(t *testing.T) {
	b := true
	if !b {
		t.Error("error testing")
	}
}
