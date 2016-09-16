package benchmarks

import (
	"testing"
)

var (
	tests = []struct {
		name string

		addingFields, withoutFields, accumulatedContext func(*testing.B)
	}{
		{"apexLog", apexLogAddingFields, apexLogWithoutFields, apexLogWithAccumulatedContext},
		{"goKit", goKitAddingFields, goKitWithoutFields, goKitWithAccumulatedContext},
		{"log15", log15AddingFields, log15WithoutFields, log15WithAccumulatedContext},
		{"logrus", logrusAddingFields, logrusWithoutFields, logrusWithAccumulatedContext},
		{"standardLibrary", nil, standardLibraryWithoutFields, nil},
		{"zapBarkify", zapBarkifyAddingFields, zapBarkifyWithoutFields, zapBarkifyWithAccumulatedContext},
		{"zapDisabledLevels", zapDisabledLevelsAddingFields, zapDisabledLevelsWithoutFields, zapDisabledLevelsWithAccumulatedContext},
		{"zapStandardize", nil, zapStandardizeWithoutFields, nil},
		{"zapDisabledLevelsCheck", zapDisabledLevelsCheckAddingFields, zapBarkifyWithoutFields, nil},
		{"zapSample", zapSampleAddingFields, zapBarkifyWithoutFields, nil},
		{"zapSampleCheck", zapSampleCheckAddingFields, zapSampleCheckWithoutFields, nil},
		{"zap", zapAddingFields, zapWithoutFields, zapWithAccumulatedContext},
	}
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
