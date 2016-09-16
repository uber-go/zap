// +build !go1.7

package benchmarks

import (
	"testing"
)

// This file is workaround for the lack of testing.B.Run() in versions of go previous to 1.7.
// It does not respect -bench filtering, but does respect other options e.g. -benchmem.

func TestHarness(_ *testing.T) {
	bm := make([]testing.InternalBenchmark, 0, len(tests)*3)
	for _, tt := range tests {
		if tt.addingFields != nil {
			bm = append(bm, testing.InternalBenchmark{tt.name + "/addingFields", tt.addingFields})
		}
		bm = append(bm, testing.InternalBenchmark{tt.name + "/withoutFields", tt.withoutFields})
		if tt.accumulatedContext != nil {
			bm = append(bm, testing.InternalBenchmark{tt.name + "/withAccumulatedContext", tt.accumulatedContext})
		}
	}
	testing.RunBenchmarks(func(_, _ string) (bool, error) { return true, nil }, bm)
}
