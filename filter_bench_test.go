package zap

import (
	"testing"
)

func BenchmarkFilter(b *testing.B) {
	log := New(NewTextEncoder(), DiscardOutput, DebugLevel)

	logger := Filter(
		LeveledLogger{DebugLevel, log},
		LeveledLogger{InfoLevel, log},
		LeveledLogger{WarnLevel, log},
		LeveledLogger{ErrorLevel, log},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("filter")
	}
}

func BenchmarkFilterUsingTeeWithLevelEnabler(b *testing.B) {
	log := New(NewTextEncoder(), DiscardOutput, DebugLevel)

	logger := Tee(log, log, log, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("filter")
	}
}
