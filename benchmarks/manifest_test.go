package benchmarks

import (
	"testing"
)

var tests = []struct {
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
