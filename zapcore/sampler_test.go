// Copyright (c) 2016-2022 Uber Technologies, Inc.
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

package zapcore_test

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap/internal/ztest"
	//revive:disable:dot-imports
	. "go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeSampler(lvl LevelEnabler, tick time.Duration, first, thereafter int) (Core, *observer.ObservedLogs) {
	core, logs := observer.New(lvl)
	// Keep using deprecated constructor for cc.
	core = NewSampler(core, tick, first, thereafter)
	return core, logs
}

func assertSequence(t testing.TB, logs []observer.LoggedEntry, lvl Level, seq ...int64) {
	seen := make([]int64, len(logs))
	for i, entry := range logs {
		require.Equal(t, "", entry.Message, "Message wasn't created by writeSequence.")
		require.Equal(t, 1, len(entry.Context), "Unexpected number of fields.")
		require.Equal(t, lvl, entry.Level, "Unexpected level.")
		f := entry.Context[0]
		require.Equal(t, "iter", f.Key, "Unexpected field key.")
		require.Equal(t, Int64Type, f.Type, "Unexpected field type")
		seen[i] = f.Integer
	}
	assert.Equal(t, seq, seen, "Unexpected sequence logged at level %v.", lvl)
}

func writeSequence(core Core, n int, lvl Level) {
	// All tests using writeSequence verify that counters are shared between
	// parent and child cores.
	core = core.With([]Field{makeInt64Field("iter", n)})
	if ce := core.Check(Entry{Level: lvl, Time: time.Now()}, nil); ce != nil {
		ce.Write()
	}
}

func TestSampler(t *testing.T) {
	for _, lvl := range []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel} {
		sampler, logs := fakeSampler(DebugLevel, time.Minute, 2, 3)

		// Ensure that counts aren't shared between levels.
		probeLevel := DebugLevel
		if lvl == DebugLevel {
			probeLevel = InfoLevel
		}
		for i := 0; i < 10; i++ {
			writeSequence(sampler, 1, probeLevel)
		}
		// Clear any output.
		logs.TakeAll()

		for i := 1; i < 10; i++ {
			writeSequence(sampler, i, lvl)
		}
		assertSequence(t, logs.TakeAll(), lvl, 1, 2, 5, 8)
	}
}

func TestLevelOfSampler(t *testing.T) {
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel}
	for _, lvl := range levels {
		lvl := lvl
		t.Run(lvl.String(), func(t *testing.T) {
			t.Parallel()

			sampler, _ := fakeSampler(lvl, time.Minute, 2, 3)
			assert.Equal(t, lvl, LevelOf(sampler), "Sampler level did not match.")
		})
	}
}

func TestSamplerDisabledLevels(t *testing.T) {
	sampler, logs := fakeSampler(InfoLevel, time.Minute, 1, 100)

	// Shouldn't be counted, because debug logging isn't enabled.
	writeSequence(sampler, 1, DebugLevel)
	writeSequence(sampler, 2, InfoLevel)
	assertSequence(t, logs.TakeAll(), InfoLevel, 2)
}

func TestSamplerTicking(t *testing.T) {
	// Ensure that we're resetting the sampler's counter every tick.
	sampler, logs := fakeSampler(DebugLevel, 10*time.Millisecond, 5, 10)

	// If we log five or fewer messages every tick, none of them should be
	// dropped.
	for tick := 0; tick < 2; tick++ {
		for i := 1; i <= 5; i++ {
			writeSequence(sampler, i, InfoLevel)
		}
		ztest.Sleep(15 * time.Millisecond)
	}
	assertSequence(
		t,
		logs.TakeAll(),
		InfoLevel,
		1, 2, 3, 4, 5, // first tick
		1, 2, 3, 4, 5, // second tick
	)

	// If we log quickly, we should drop some logs. The first five statements
	// each tick should be logged, then every tenth.
	for tick := 0; tick < 3; tick++ {
		for i := 1; i < 18; i++ {
			writeSequence(sampler, i, InfoLevel)
		}
		ztest.Sleep(10 * time.Millisecond)
	}

	assertSequence(
		t,
		logs.TakeAll(),
		InfoLevel,
		1, 2, 3, 4, 5, 15, // first tick
		1, 2, 3, 4, 5, 15, // second tick
		1, 2, 3, 4, 5, 15, // third tick
	)
}

type countingCore struct {
	logs atomic.Uint32
}

func (c *countingCore) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	return ce.AddCore(ent, c)
}

func (c *countingCore) Write(Entry, []Field) error {
	c.logs.Add(1)
	return nil
}

func (c *countingCore) With([]Field) Core { return c }
func (c *countingCore) Fields() []Field   { return nil }
func (*countingCore) Enabled(Level) bool  { return true }
func (*countingCore) Sync() error         { return nil }

func TestSamplerConcurrent(t *testing.T) {
	const (
		logsPerTick   = 10
		numMessages   = 5
		numTicks      = 25
		numGoroutines = 10
		tick          = 10 * time.Millisecond

		// We'll make a total of,
		// (numGoroutines * numTicks * logsPerTick * 2) log attempts
		// with numMessages unique messages.
		numLogAttempts = numGoroutines * logsPerTick * numTicks * 2
		// Of those, we'll accept (logsPerTick * numTicks) entries
		// for each unique message.
		expectedCount = numMessages * logsPerTick * numTicks
		// The rest will be dropped.
		expectedDropped = numLogAttempts - expectedCount
	)

	clock := ztest.NewMockClock()

	cc := &countingCore{}

	hook, dropped, sampled := makeSamplerCountingHook()
	sampler := NewSamplerWithOptions(cc, tick, logsPerTick, 100000, SamplerHook(hook))

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int, ticker *time.Ticker) {
			defer wg.Done()
			defer ticker.Stop()

			for {
				select {
				case <-stop:
					return

				case <-ticker.C:
					for j := 0; j < logsPerTick*2; j++ {
						msg := fmt.Sprintf("msg%v", i%numMessages)
						ent := Entry{
							Level:   DebugLevel,
							Message: msg,
							Time:    clock.Now(),
						}
						if ce := sampler.Check(ent, nil); ce != nil {
							ce.Write()
						}

						// Give a chance for other goroutines to run.
						runtime.Gosched()
					}
				}
			}
		}(i, clock.NewTicker(tick))
	}

	clock.Add(tick * numTicks)
	close(stop)
	wg.Wait()

	assert.Equal(
		t,
		expectedCount,
		int(cc.logs.Load()),
		"Unexpected number of logs",
	)
	assert.Equal(t,
		expectedCount,
		int(sampled.Load()),
		"Unexpected number of logs sampled",
	)
	assert.Equal(t,
		expectedDropped,
		int(dropped.Load()),
		"Unexpected number of logs dropped",
	)
}

func TestSamplerRaces(t *testing.T) {
	sampler, _ := fakeSampler(DebugLevel, time.Minute, 1, 1000)

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-start
			for j := 0; j < 100; j++ {
				writeSequence(sampler, j, InfoLevel)
			}
			wg.Done()
		}()
	}

	close(start)
	wg.Wait()
}

func TestSamplerUnknownLevels(t *testing.T) {
	// Prove that out-of-bounds levels don't panic.
	unknownLevels := []Level{
		DebugLevel - 1,
		FatalLevel + 1,
	}

	for _, lvl := range unknownLevels {
		t.Run(lvl.String(), func(t *testing.T) {
			sampler, logs := fakeSampler(lvl, time.Minute, 2, 3)
			for i := 1; i < 10; i++ {
				writeSequence(sampler, i, lvl)
			}

			// Expect no sampling for unknown levels.
			assertSequence(t, logs.TakeAll(), lvl, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		})
	}
}

func TestSamplerWithZeroThereafter(t *testing.T) {
	var counter countingCore

	// Logs two messages per second.
	sampler := NewSamplerWithOptions(&counter, time.Second, 2, 0)

	now := time.Now()

	for i := 0; i < 1000; i++ {
		ent := Entry{
			Level:   InfoLevel,
			Message: "msg",
			Time:    now,
		}
		if ce := sampler.Check(ent, nil); ce != nil {
			ce.Write()
		}
	}

	assert.Equal(t, 2, int(counter.logs.Load()),
		"Unexpected number of logs")

	now = now.Add(time.Second)

	for i := 0; i < 1000; i++ {
		ent := Entry{
			Level:   InfoLevel,
			Message: "msg",
			Time:    now,
		}
		if ce := sampler.Check(ent, nil); ce != nil {
			ce.Write()
		}
	}

	assert.Equal(t, 4, int(counter.logs.Load()),
		"Unexpected number of logs")
}
