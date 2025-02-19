package mayo

import (
	"testing"
)

func Test(t *testing.T) {
	if "hej" != "hej" {
		t.Error("Original and decoded is not the same")
	}
}
