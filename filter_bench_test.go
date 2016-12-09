// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
