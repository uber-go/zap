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

package zapcore

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/testutils"

	"github.com/stretchr/testify/assert"
)

func makeInt64Field(key string, val int) Field {
	return Field{Type: Int64Type, Integer: int64(val), Key: key}
}

func TestObserverWith(t *testing.T) {
	sf1, logs := NewObserver(InfoLevel, 10)

	// need to pad out enough initial fields so that the underlying slice cap()
	// gets ahead of its len() so that the sf3/4 With append's could choose
	// not to copy (if the implementation doesn't force them)
	sf1 = sf1.With([]Field{makeInt64Field("a", 1), makeInt64Field("b", 2)})

	sf2 := sf1.With([]Field{makeInt64Field("c", 3)})
	sf3 := sf2.With([]Field{makeInt64Field("d", 4)})
	sf4 := sf2.With([]Field{makeInt64Field("e", 5)})
	ent := Entry{Level: InfoLevel, Message: "hello"}

	for i, f := range []Facility{sf2, sf3, sf4} {
		if ce := f.Check(ent, nil); ce != nil {
			ce.Write(makeInt64Field("i", i))
		}
	}

	assert.Equal(t, []ObservedLog{
		{
			Entry: ent,
			Context: []Field{
				makeInt64Field("a", 1),
				makeInt64Field("b", 2),
				makeInt64Field("c", 3),
				makeInt64Field("i", 0),
			},
		},
		{
			Entry: ent,
			Context: []Field{
				makeInt64Field("a", 1),
				makeInt64Field("b", 2),
				makeInt64Field("c", 3),
				makeInt64Field("d", 4),
				makeInt64Field("i", 1),
			},
		},
		{
			Entry: ent,
			Context: []Field{
				makeInt64Field("a", 1),
				makeInt64Field("b", 2),
				makeInt64Field("c", 3),
				makeInt64Field("e", 5),
				makeInt64Field("i", 2),
			},
		},
	}, logs.All(), "expected no field sharing between With siblings")
}

func TestWriterFacilitySyncsOutput(t *testing.T) {
	tests := []struct {
		entry      Entry
		shouldSync bool
	}{
		{Entry{Level: DebugLevel}, false},
		{Entry{Level: InfoLevel}, false},
		{Entry{Level: WarnLevel}, false},
		{Entry{Level: ErrorLevel}, false},
		{Entry{Level: DPanicLevel}, true},
		{Entry{Level: PanicLevel}, true},
		{Entry{Level: FatalLevel}, true},
	}

	for _, tt := range tests {
		sink := &testutils.Discarder{}
		fac := WriterFacility(
			NewJSONEncoder(testJSONConfig()),
			sink,
			DebugLevel,
		)

		fac.Write(tt.entry, nil)
		assert.Equal(t, tt.shouldSync, sink.Called(), "Incorrect Sync behavior.")
	}
}

func TestWriterFacilityWriteFailure(t *testing.T) {
	fac := WriterFacility(
		NewJSONEncoder(testJSONConfig()),
		Lock(&testutils.FailWriter{}),
		DebugLevel,
	)
	err := fac.Write(Entry{}, nil)
	// Should log the error.
	assert.Error(t, err, "Expected writing Entry to fail.")
}

func TestWriterFacilityShortWrite(t *testing.T) {
	fac := WriterFacility(
		NewJSONEncoder(testJSONConfig()),
		Lock(&testutils.ShortWriter{}),
		DebugLevel,
	)
	err := fac.Write(Entry{}, nil)
	// Should log the error.
	assert.Error(t, err, "Expected writing Entry to fail.")
}

func writeIter(fac Facility, n int, lvl Level) {
	fac = fac.With([]Field{makeInt64Field("iter", n)})
	if ce := fac.Check(Entry{Level: lvl}, nil); ce != nil {
		ce.Write()
	}
}

func fakeSampler(lvl LevelEnabler, tick time.Duration, first, thereafter int) (Facility, *ObservedLogs) {
	fac, logs := NewObserver(lvl, 128)
	fac = Sample(fac, tick, first, thereafter)
	return fac, logs
}

func buildExpectation(level Level, nums ...int) []ObservedLog {
	var expected []ObservedLog
	for _, n := range nums {
		expected = append(expected, ObservedLog{
			Entry:   Entry{Level: level},
			Context: []Field{makeInt64Field("iter", n)},
		})
	}
	return expected
}

func TestSampler(t *testing.T) {
	for _, lvl := range []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel} {
		sampler, logs := fakeSampler(DebugLevel, time.Minute, 2, 3)
		for i := 1; i < 10; i++ {
			writeIter(sampler, i, lvl)
		}
		assert.Equal(t, buildExpectation(lvl, 1, 2, 5, 8), logs.All(), "Unexpected output from sampled logger.")
	}
}

func TestSamplerDisabledLevels(t *testing.T) {
	sampler, logs := fakeSampler(InfoLevel, time.Minute, 1, 100)

	// Shouldn't be counted, because debug logging isn't enabled.
	writeIter(sampler, 1, DebugLevel)
	writeIter(sampler, 2, InfoLevel)
	expected := buildExpectation(InfoLevel, 2)
	assert.Equal(t, expected, logs.All(), "Expected to disregard disabled log levels.")
}

func TestSamplerWithSharesCounters(t *testing.T) {
	sampler, logs := fakeSampler(DebugLevel, time.Minute, 1, 100)

	first := sampler.With([]Field{makeInt64Field("child", 1)})
	for i := 1; i < 10; i++ {
		writeIter(first, i, InfoLevel)
	}
	second := sampler.With([]Field{makeInt64Field("child", 2)})
	// The new child logger should share the same counters, so we don't expect to
	// write these logs.
	for i := 10; i < 20; i++ {
		writeIter(second, i, InfoLevel)
	}

	expected := []ObservedLog{{
		Entry:   Entry{Level: InfoLevel},
		Context: []Field{makeInt64Field("child", 1), makeInt64Field("iter", 1)},
	}}
	assert.Equal(t, expected, logs.All(), "Expected child loggers to share counters.")
}

func TestSamplerTicks(t *testing.T) {
	// Ensure that we're resetting the sampler's counter every tick.
	sampler, logs := fakeSampler(DebugLevel, time.Millisecond, 1, 1000)

	// The first statement should be logged, the second should be skipped but
	// start the reset timer, and then we sleep. After sleeping for more than a
	// tick, the third statement should be logged.
	for i := 1; i < 4; i++ {
		if i == 3 {
			testutils.Sleep(5 * time.Millisecond)
		}
		writeIter(sampler, i, InfoLevel)
	}

	expected := buildExpectation(InfoLevel, 1, 3)
	assert.Equal(t, expected, logs.All(), "Expected sleeping for a tick to reset sampler.")
}

func TestSamplerCheck(t *testing.T) {
	sampler, logs := fakeSampler(InfoLevel, time.Millisecond, 1, 10)

	assert.Nil(t, sampler.Check(Entry{Level: DebugLevel}, nil), "Expected a nil CheckedMessage at disabled log levels.")

	for i := 1; i < 12; i++ {
		if cm := sampler.Check(Entry{Level: InfoLevel}, nil); cm != nil {
			cm.Write(makeInt64Field("iter", i))
		}
	}

	expected := buildExpectation(InfoLevel, 1, 11)
	assert.Equal(t, expected, logs.All(), "Unexpected output when sampling with Check.")
}

// TODO: restore this test, now that panic and fatal are actually terminal
// func TestSamplerCheckPanicFatal(t *testing.T) {
// 	for _, level := range []zap.Level{zap.FatalLevel, zap.PanicLevel} {
// 		sampler, sink := fakeSampler(zap.FatalLevel+1, time.Millisecond, 1, 10, false)

// 		assert.Nil(t, sampler.Check(zap.DebugLevel, "foo"), "Expected a nil CheckedMessage at disabled log levels.")
// 		for i := 0; i < 5; i++ {
// 			if cm := sampler.Check(level, "sample"); assert.True(t, cm.OK(), "expected %v level to always be OK", level) {
// 				cm.Write(zap.Int("iter", i))
// 			}
// 		}

// 		assert.Equal(t, []spy.Log(nil), sink.Logs(), "Unexpected output when sampling with Check.")
// 	}
// }

func TestSamplerRaces(t *testing.T) {
	sampler, _ := fakeSampler(DebugLevel, time.Minute, 1, 1000)

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-start
			for j := 0; j < 100; j++ {
				writeIter(sampler, j, InfoLevel)
			}
			wg.Done()
		}()
	}

	close(start)
	wg.Wait()
}
