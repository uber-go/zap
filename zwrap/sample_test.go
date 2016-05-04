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

package zwrap

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/uber-common/zap"
	"github.com/uber-common/zap/spy"

	"github.com/stretchr/testify/assert"
)

func fakeSampler(tick time.Duration, first, thereafter int, development bool) (zap.Logger, *spy.Sink) {
	base, sink := spy.New()
	base.SetLevel(zap.All)
	base.SetDevelopment(development)
	sampler := Sample(base, tick, first, thereafter)
	return sampler, sink
}

func buildExpectation(level zap.Level, messages ...string) []spy.Log {
	var expected []spy.Log
	for _, m := range messages {
		expected = append(expected, spy.Log{
			Level:  level,
			Msg:    m,
			Fields: make([]zap.Field, 0),
		})
	}
	return expected
}

func TestSampler(t *testing.T) {
	tests := []struct {
		level       zap.Level
		logFunc     func(zap.Logger, string)
		development bool
	}{
		{
			level:   zap.Debug,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Debug(msg) },
		},
		{
			level:   zap.Info,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Info(msg) },
		},
		{
			level:   zap.Warn,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Warn(msg) },
		},
		{
			level:   zap.Error,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Error(msg) },
		},
		{
			level:   zap.Panic,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Panic(msg) },
		},
		{
			level:   zap.Fatal,
			logFunc: func(sampler zap.Logger, msg string) { sampler.Fatal(msg) },
		},
		{
			level:   zap.Error,
			logFunc: func(sampler zap.Logger, msg string) { sampler.DFatal(msg) },
		},
		{
			level:       zap.Fatal,
			logFunc:     func(sampler zap.Logger, msg string) { sampler.DFatal(msg) },
			development: true,
		},
	}

	for _, tt := range tests {
		sampler, sink := fakeSampler(time.Minute, 2, 3, tt.development)
		for i := 1; i < 10; i++ {
			tt.logFunc(sampler, strconv.Itoa(i))
		}
		expected := buildExpectation(tt.level, "1", "2", "5", "8")
		assert.Equal(t, expected, sink.Logs(), "Unexpected output from sampled logger.")
	}
}

func TestSampledDisabledLevels(t *testing.T) {
	sampler, sink := fakeSampler(time.Minute, 1, 100, false)
	sampler.SetLevel(zap.Info)

	// Shouldn't be counted, because debug logging isn't enabled.
	sampler.Debug("1")
	sampler.Info("2")
	expected := buildExpectation(zap.Info, "2")
	assert.Equal(t, expected, sink.Logs(), "Expected to disregard disabled log levels.")
}

func TestSamplerWith(t *testing.T) {
	// Check that child loggers are sampled and independent.
	sampler, sink := fakeSampler(time.Minute, 1, 100, false)

	expected := []spy.Log{
		{
			Level:  zap.Info,
			Msg:    "1",
			Fields: []zap.Field{zap.String("child", "first")},
		},
		{
			Level:  zap.Info,
			Msg:    "20",
			Fields: []zap.Field{zap.String("child", "second")},
		},
	}

	first := sampler.With(zap.String("child", "first"))
	for i := 1; i < 100; i++ {
		first.Info(strconv.Itoa(i))
	}
	second := sampler.With(zap.String("child", "second"))
	// Even though the first child already logged 20 messages, we should see the
	// first message from this child.
	for i := 20; i < 40; i++ {
		second.Info(strconv.Itoa(i))
	}

	assert.Equal(t, expected, sink.Logs(), "Expected child loggers to maintain separate counters.")
}

func TestSamplerTicks(t *testing.T) {
	// Ensure that we're resetting the sampler's counter every tick.
	sampler, sink := fakeSampler(time.Millisecond, 1, 1000, false)

	// The first statement should be logged, the second should be skipped but
	// start the reset timer, and then we sleep. After sleeping for more than a
	// tick, the third statement should be logged.
	for i := 1; i < 4; i++ {
		if i == 3 {
			time.Sleep(2 * time.Millisecond)
		}
		sampler.Info(strconv.Itoa(i))
	}

	expected := buildExpectation(zap.Info, "1", "3")
	assert.Equal(t, expected, sink.Logs(), "Expected sleeping for a tick to reset sampler.")
}

func TestSamplerBucketsByCaller(t *testing.T) {
	sampler, sink := fakeSampler(time.Minute, 1, 1000, false)
	for i := 0; i < 5; i++ {
		sampler.Info("First call site.")
	}
	for i := 0; i < 5; i++ {
		sampler.Info("Second call site.")
	}
	expected := buildExpectation(zap.Info, "First call site.", "Second call site.")
	assert.Equal(t, expected, sink.Logs(), "Expected to sample each call site separately.")
}

func TestSamplerRaces(t *testing.T) {
	sampler, _ := fakeSampler(time.Minute, 1, 1000, false)

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-start
			for j := 0; j < 100; j++ {
				sampler.Info("Testing for races.")
			}
			wg.Done()
		}()
	}

	close(start)
	wg.Wait()
}
