package enrichment

import (
	"testing"
)

func TestFuelAnomalyDetectorCheck(t *testing.T) {
	fd := &FuelAnomalyDetector{}
	if fd.logger == nil {
		// Just a dummy test
	}
}
