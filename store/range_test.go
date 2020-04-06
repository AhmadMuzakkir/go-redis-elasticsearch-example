package store

import (
	"testing"
)

// Test each range constant should have description.
func TestRangeDescription(t *testing.T) {
	for i := 0; i < int(NumRange); i++ {
		if RangeDescription(Range(i)) == "" {
			t.Fatalf("unimplemented range description for range %d", i)
		}
	}
}
