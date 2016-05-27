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
	"sync"
	"testing"
	"time"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/spy"
	"github.com/uber-go/zap/testutils"

	"github.com/stretchr/testify/assert"
)

func WithIter(l zap.Logger, n int) zap.Logger {
	return l.With(zap.Int("iter", n))
}

func fakeSampler(tick time.Duration, first, thereafter int, development bool) (zap.Logger, *spy.Sink) {
	base, sink := spy.New()
	base.SetLevel(zap.All)
	base.SetDevelopment(development)
	sampler := Sample(base, tick, first, thereafter)
	return sampler, sink
}

func buildExpectation(level zap.Level, nums ...int) []spy.Log {
	var expected []spy.Log
	for _, n := range nums {
		expected = append(expected, spy.Log{
			Level:  level,
			Msg:    "sample",
			Fields: []zap.Field{zap.Int("iter", n)},
		})
	}
	return expected
}

func TestSampler(t *testing.T) {
	tests := []struct {
		level       zap.Level
		logFunc     func(zap.Logger, int)
		development bool
	}{
		{
			level:   zap.Debug,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Debug("sample") },
		},
		{
			level:   zap.Info,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Info("sample") },
		},
		{
			level:   zap.Warn,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Warn("sample") },
		},
		{
			level:   zap.Error,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Error("sample") },
		},
		{
			level:   zap.Panic,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Panic("sample") },
		},
		{
			level:   zap.Fatal,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Fatal("sample") },
		},
		{
			level:   zap.Error,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).DFatal("sample") },
		},
		{
			level:       zap.Fatal,
			logFunc:     func(sampler zap.Logger, n int) { WithIter(sampler, n).DFatal("sample") },
			development: true,
		},
		{
			level:   zap.Error,
			logFunc: func(sampler zap.Logger, n int) { WithIter(sampler, n).Log(zap.Error, "sample") },
		},
	}

	for _, tt := range tests {
		sampler, sink := fakeSampler(time.Minute, 2, 3, tt.development)
		for i := 1; i < 10; i++ {
			tt.logFunc(sampler, i)
		}
		expected := buildExpectation(tt.level, 1, 2, 5, 8)
		assert.Equal(t, expected, sink.Logs(), "Unexpected output from sampled logger.")
	}
}

func TestSampledDisabledLevels(t *testing.T) {
	sampler, sink := fakeSampler(time.Minute, 1, 100, false)
	sampler.SetLevel(zap.Info)

	// Shouldn't be counted, because debug logging isn't enabled.
	WithIter(sampler, 1).Debug("sample")
	WithIter(sampler, 2).Info("sample")
	expected := buildExpectation(zap.Info, 2)
	assert.Equal(t, expected, sink.Logs(), "Expected to disregard disabled log levels.")
}

func TestSamplerWithSharesCounters(t *testing.T) {
	logger, sink := fakeSampler(time.Minute, 1, 100, false)

	expected := []spy.Log{
		{
			Level:  zap.Info,
			Msg:    "sample",
			Fields: []zap.Field{zap.String("child", "first"), zap.Int("iter", 1)},
		},
	}

	first := logger.With(zap.String("child", "first"))
	for i := 1; i < 10; i++ {
		WithIter(first, i).Info("sample")
	}
	second := logger.With(zap.String("child", "second"))
	// The new child logger should share the same counters, so we don't expect to
	// write these logs.
	for i := 10; i < 20; i++ {
		WithIter(second, i).Info("sample")
	}

	assert.Equal(t, expected, sink.Logs(), "Expected child loggers to share counters.")
}

func TestSamplerTicks(t *testing.T) {
	// Ensure that we're resetting the sampler's counter every tick.
	sampler, sink := fakeSampler(time.Millisecond, 1, 1000, false)

	// The first statement should be logged, the second should be skipped but
	// start the reset timer, and then we sleep. After sleeping for more than a
	// tick, the third statement should be logged.
	for i := 1; i < 4; i++ {
		if i == 3 {
			testutils.Sleep(5 * time.Millisecond)
		}
		WithIter(sampler, i).Info("sample")
	}

	expected := buildExpectation(zap.Info, 1, 3)
	assert.Equal(t, expected, sink.Logs(), "Expected sleeping for a tick to reset sampler.")
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
