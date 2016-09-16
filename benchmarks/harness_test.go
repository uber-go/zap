// +build go1.7

package benchmarks

import (
	"testing"
)

func BenchmarkAddingFields(b *testing.B) {
	for _, tt := range tests {
		if tt.addingFields != nil {
			b.Run(tt.name, tt.addingFields)
		}
	}
}

func BenchmarkWithoutFields(b *testing.B) {
	for _, tt := range tests {
		b.Run(tt.name, tt.withoutFields)
	}
}

func BenchmarkWithAccumulatedContext(b *testing.B) {
	for _, tt := range tests {
		if tt.accumulatedContext != nil {
			b.Run(tt.name, tt.accumulatedContext)
		}
	}
}
