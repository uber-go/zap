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

	"go.uber.org/zap/internal/exit"

	"github.com/stretchr/testify/assert"
)

func assertGoexit(t *testing.T, f func()) {
	var finished bool
	recovered := make(chan interface{})
	go func() {
		defer func() {
			recovered <- recover()
		}()

		f()
		finished = true
	}()

	assert.Nil(t, <-recovered, "Goexit should cause recover to return nil")
	assert.False(t, finished, "Goroutine should not finish after Goexit")
}

func TestPutNilEntry(t *testing.T) {
	// Pooling nil entries defeats the purpose.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			putCheckedEntry(nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ce := getCheckedEntry()
			assert.NotNil(t, ce, "Expected only non-nil CheckedEntries in pool.")
			assert.False(t, ce.dirty, "Unexpected dirty bit set.")
			assert.Nil(t, ce.ErrorOutput, "Non-nil ErrorOutput.")
			assert.Nil(t, ce.after, "Unexpected terminal behavior.")
			assert.Equal(t, 0, len(ce.cores), "Expected empty slice of cores.")
			assert.True(t, cap(ce.cores) > 0, "Expected pooled CheckedEntries to pre-allocate slice of Cores.")
		}
	}()

	wg.Wait()
}

func TestEntryCaller(t *testing.T) {
	tests := []struct {
		caller EntryCaller
		full   string
		short  string
	}{
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, false),
			full:   "undefined",
			short:  "undefined",
		},
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, true),
			full:   "/path/to/foo.go:42",
			short:  "to/foo.go:42",
		},
		{
			caller: NewEntryCaller(100, "to/foo.go", 42, true),
			full:   "to/foo.go:42",
			short:  "to/foo.go:42",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.full, tt.caller.String(), "Unexpected string from EntryCaller.")
		assert.Equal(t, tt.full, tt.caller.FullPath(), "Unexpected FullPath from EntryCaller.")
		assert.Equal(t, tt.short, tt.caller.TrimmedPath(), "Unexpected TrimmedPath from EntryCaller.")
	}
}

func TestCheckedEntryWrite(t *testing.T) {
	t.Run("nil is safe", func(t *testing.T) {
		var ce *CheckedEntry
		assert.NotPanics(t, func() { ce.Write() }, "Unexpected panic writing nil CheckedEntry.")
	})

	t.Run("WriteThenPanic", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenPanic)
		assert.Panics(t, func() { ce.Write() }, "Expected to panic when WriteThenPanic is set.")
	})

	t.Run("WriteThenGoexit", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenGoexit)
		assertGoexit(t, func() { ce.Write() })
	})

	t.Run("WriteThenFatal", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.After(Entry{}, WriteThenFatal)
		stub := exit.WithStub(func() {
			ce.Write()
		})
		assert.True(t, stub.Exited, "Expected to exit when WriteThenFatal is set.")
		assert.Equal(t, 1, stub.Code, "Expected to exit when WriteThenFatal is set.")
	})

	t.Run("After", func(t *testing.T) {
		var ce *CheckedEntry
		hook := &customHook{}
		ce = ce.After(Entry{}, hook)
		ce.Write()
		assert.True(t, hook.called, "Expected to call custom action after Write.")
	})
}

type customHook struct {
	called bool
}

func (c *customHook) OnWrite(_ *CheckedEntry, _ []Field) {
	c.called = true
}

func TestCheckedEntryBefore(t *testing.T) {
	t.Run("nil is safe", func(t *testing.T) {
		var ce *CheckedEntry
		ce = ce.Before(Entry{Message: "hello"}, func(ent Entry, fields []Field) (Entry, []Field) {
			ent.Message = "modified"
			return ent, fields
		})
		assert.NotNil(t, ce)
		assert.Equal(t, "hello", ce.Entry.Message)
	})

	t.Run("modifies entry and fields", func(t *testing.T) {
		core := &recordingCore{}
		var ce *CheckedEntry
		ce = ce.AddCore(Entry{Message: "original"}, core)
		ce = ce.Before(Entry{}, func(ent Entry, fields []Field) (Entry, []Field) {
			ent.Message = "modified"
			fields = append(fields, Field{Key: "added", Type: StringType, String: "value"})
			return ent, fields
		})
		ce.Write(Field{Key: "initial", Type: StringType, String: "v"})

		assert.Equal(t, "modified", core.entry.Message)
		assert.Len(t, core.fields, 2)
		assert.Equal(t, "initial", core.fields[0].Key)
		assert.Equal(t, "added", core.fields[1].Key)
	})

	t.Run("multiple hooks chain in order", func(t *testing.T) {
		core := &recordingCore{}
		var ce *CheckedEntry
		ce = ce.AddCore(Entry{Message: "start"}, core)
		ce = ce.Before(Entry{}, func(ent Entry, fields []Field) (Entry, []Field) {
			ent.Message = ent.Message + "-first"
			return ent, fields
		})
		ce = ce.Before(Entry{}, func(ent Entry, fields []Field) (Entry, []Field) {
			ent.Message = ent.Message + "-second"
			return ent, fields
		})
		ce.Write()

		assert.Equal(t, "start-first-second", core.entry.Message)
	})

	t.Run("hooks reset on pool reuse", func(t *testing.T) {
		core := &recordingCore{}
		var ce *CheckedEntry
		ce = ce.AddCore(Entry{Message: "first"}, core)
		ce = ce.Before(Entry{}, func(ent Entry, fields []Field) (Entry, []Field) {
			ent.Message = "hooked"
			return ent, fields
		})
		ce.Write()
		assert.Equal(t, "hooked", core.entry.Message)

		// Get a new entry from the pool — hooks should be cleared.
		ce2 := getCheckedEntry()
		assert.Empty(t, ce2.before)
		putCheckedEntry(ce2)
	})
}

// recordingCore captures the last entry and fields written to it.
type recordingCore struct {
	entry  Entry
	fields []Field
}

func (c *recordingCore) Enabled(Level) bool                              { return true }
func (c *recordingCore) With([]Field) Core                               { return c }
func (c *recordingCore) Check(ent Entry, ce *CheckedEntry) *CheckedEntry { return ce.AddCore(ent, c) }
func (c *recordingCore) Sync() error                                     { return nil }
func (c *recordingCore) Write(ent Entry, fields []Field) error {
	c.entry = ent
	c.fields = fields
	return nil
}
