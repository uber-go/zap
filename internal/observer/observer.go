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

package observer

import (
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

// An LoggedEntry is an encoding-agnostic representation of a log message.
// Field availability is context dependant.
type LoggedEntry struct {
	zapcore.Entry
	Context []zapcore.Field
}

// ObservedLogs is a concurrency-safe, ordered collection of observed logs.
type ObservedLogs struct {
	mu   sync.RWMutex
	logs []LoggedEntry
}

// Add appends a new observed log to the collection. It always returns nil
// error, but does so that it can be used as an observer sink function.
func (o *ObservedLogs) Add(log LoggedEntry) error {
	o.mu.Lock()
	o.logs = append(o.logs, log)
	o.mu.Unlock()
	return nil
}

// Len returns the number of items in the collection.
func (o *ObservedLogs) Len() int {
	o.mu.RLock()
	n := len(o.logs)
	o.mu.RUnlock()
	return n
}

// All returns a copy of all the observed logs.
func (o *ObservedLogs) All() []LoggedEntry {
	o.mu.RLock()
	ret := make([]LoggedEntry, len(o.logs))
	for i := range o.logs {
		ret[i] = o.logs[i]
	}
	o.mu.RUnlock()
	return ret
}

// TakeAll returns a copy of all the observed logs, and truncates the observed
// slice.
func (o *ObservedLogs) TakeAll() []LoggedEntry {
	o.mu.Lock()
	ret := o.logs
	o.logs = nil
	o.mu.Unlock()
	return ret
}

// AllUntimed returns a copy of all the observed logs, but overwrites the
// observed timestamps with time.Time's zero value. This is useful when making
// assertions in tests.
func (o *ObservedLogs) AllUntimed() []LoggedEntry {
	ret := o.All()
	for i := range ret {
		ret[i].Time = time.Time{}
	}
	return ret
}

type observer struct {
	zapcore.LevelEnabler
	sink func(LoggedEntry) error
}

// New creates a new Core that buffers logs in memory (without any encoding).
// It's particularly useful in tests.
func New(enab zapcore.LevelEnabler, sink func(LoggedEntry) error, withContext bool) zapcore.Core {
	if withContext {
		return &contextObserver{
			LevelEnabler: enab,
			sink:         sink,
		}
	}
	return &observer{
		LevelEnabler: enab,
		sink:         sink,
	}
}

func (o *observer) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if o.Enabled(ent.Level) {
		return ce.AddCore(ent, o)
	}
	return ce
}

func (o *observer) With(fields []zapcore.Field) zapcore.Core {
	return &observer{LevelEnabler: o.LevelEnabler, sink: o.sink}
}

func (o *observer) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return o.sink(LoggedEntry{ent, fields})
}

type contextObserver struct {
	zapcore.LevelEnabler
	sink    func(LoggedEntry) error
	context []zapcore.Field
}

func (co *contextObserver) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if co.Enabled(ent.Level) {
		return ce.AddCore(ent, co)
	}
	return ce
}

func (co *contextObserver) With(fields []zapcore.Field) zapcore.Core {
	return &contextObserver{
		LevelEnabler: co.LevelEnabler,
		sink:         co.sink,
		context:      append(co.context[:len(co.context):len(co.context)], fields...),
	}
}

func (co *contextObserver) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	all := make([]zapcore.Field, 0, len(fields)+len(co.context))
	all = append(all, co.context...)
	all = append(all, fields...)
	return co.sink(LoggedEntry{ent, all})
}
