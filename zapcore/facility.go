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
	"time"

	"go.uber.org/atomic"
)

// EntryWriter is a destination for log entttries.  It can have pervasive
// fields added with With().
type EntryWriter interface {
	Write(Entry, []Field) error
	With([]Field) EntryWriter
}

// Facility is a destination for log entries. It can have pervasive fields
// added with With().
type Facility interface {
	LevelEnabler

	// Spiritually this embeds EntryWriter with a promoted With return type.
	With([]Field) Facility
	Write(Entry, []Field) error

	Check(Entry, *CheckedEntry) *CheckedEntry
}

// WriterFacility creates a facility that writes logs to a WriteSyncer. By
// default, if w is nil, os.Stdout is used.
func WriterFacility(enc Encoder, ws WriteSyncer, enab LevelEnabler) Facility {
	return &ioFacility{
		LevelEnabler: enab,
		enc:          enc,
		out:          Lock(ws),
	}
}

type ioFacility struct {
	LevelEnabler
	enc Encoder
	out WriteSyncer
}

func (iof *ioFacility) With(fields []Field) Facility {
	clone := iof.clone()
	addFields(clone.enc, fields)
	return clone
}

func (iof *ioFacility) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if iof.Enabled(ent.Level) {
		return ce.AddFacility(ent, iof)
	}
	return ce
}

func (iof *ioFacility) Write(ent Entry, fields []Field) error {
	if err := iof.enc.WriteEntry(iof.out, ent, fields); err != nil {
		return err
	}
	if ent.Level > ErrorLevel {
		// Since we may be crashing the program, sync the output.
		return iof.out.Sync()
	}
	return nil
}

func (iof *ioFacility) clone() *ioFacility {
	return &ioFacility{
		LevelEnabler: iof.LevelEnabler,
		enc:          iof.enc.Clone(),
		out:          iof.out,
	}
}

// An ObservedLog is an encoding-agnostic representation of a log message.
type ObservedLog struct {
	Entry
	Context []Field
}

func (o ObservedLog) untimed() ObservedLog {
	o.Entry.Time = time.Time{}
	return o
}

// ObservedLogs is a concurrency-safe, ordered collection of observed logs. It
// has a fixed capacity, after which new logs begin overwriting the oldest
// observations.
type ObservedLogs struct {
	cap  int
	mu   sync.RWMutex
	logs []ObservedLog
}

// Add appends a new observed log to the collection.
func (o *ObservedLogs) Add(ent Entry, context []Field) {
	log := ObservedLog{
		Entry:   ent,
		Context: context,
	}
	o.mu.Lock()
	o.logs = append(o.logs, log)
	if len(o.logs) > o.cap {
		o.logs = o.logs[1:]
	}
	o.mu.Unlock()
}

// Len returns the number of items in the collection.
func (o *ObservedLogs) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.logs)
}

// All returns a copy of all the observed logs.
func (o *ObservedLogs) All() []ObservedLog {
	o.mu.RLock()
	ret := make([]ObservedLog, len(o.logs))
	for i := range o.logs {
		ret[i] = o.logs[i]
	}
	o.mu.RUnlock()
	return ret
}

// AllUntimed returns a copy of all the observed logs, but overwrites the
// observed timestamps with time.Time's zero value. This is useful when making
// assertions in tests.
func (o *ObservedLogs) AllUntimed() []ObservedLog {
	o.mu.RLock()
	ret := make([]ObservedLog, len(o.logs))
	for i := range o.logs {
		ret[i] = o.logs[i].untimed()
	}
	o.mu.RUnlock()
	return ret
}

type observerFacility struct {
	LevelEnabler
	sink    *ObservedLogs
	context []Field
}

// NewObserver creates a new facility that buffers logs in memory (without any
// encoding). It's particularly useful in tests, though it can serve a variety
// of other purposes as well. This constructor returns the facility itself and
// a function to retrieve the observed logs.
func NewObserver(enab LevelEnabler, cap int) (Facility, *ObservedLogs) {
	fac := &observerFacility{
		LevelEnabler: enab,
		sink:         &ObservedLogs{cap: cap, logs: make([]ObservedLog, 0, cap)},
	}
	return fac, fac.sink
}

func (o *observerFacility) With(fields []Field) Facility {
	return &observerFacility{
		LevelEnabler: o.LevelEnabler,
		sink:         o.sink,
		context:      append(o.context[:len(o.context):len(o.context)], fields...),
	}
}

func (o *observerFacility) Write(ent Entry, fields []Field) error {
	all := make([]Field, 0, len(fields)+len(o.context))
	all = append(all, o.context...)
	all = append(all, fields...)
	o.sink.Add(ent, all)
	return nil
}

func (o *observerFacility) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if o.Enabled(ent.Level) {
		ce = ce.AddFacility(ent, o)
	}
	return ce
}

type counters struct {
	sync.RWMutex
	counts map[string]*atomic.Uint64
}

func (c *counters) Inc(key string) uint64 {
	c.RLock()
	count, ok := c.counts[key]
	c.RUnlock()
	if ok {
		return count.Inc()
	}

	c.Lock()
	count, ok = c.counts[key]
	if ok {
		c.Unlock()
		return count.Inc()
	}

	c.counts[key] = atomic.NewUint64(1)
	c.Unlock()
	return 1
}

func (c *counters) Reset(key string) {
	c.Lock()
	count := c.counts[key]
	c.Unlock()
	count.Store(0)
}

// Sample creates a facility that samples incoming entries.
func Sample(fac Facility, tick time.Duration, first, thereafter int) Facility {
	return &sampler{
		Facility:   fac,
		tick:       tick,
		counts:     &counters{counts: make(map[string]*atomic.Uint64)},
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

// samplre must be a Facility, and not an EntryWriter, since it must make a
// stateful `Check` decision, while leaving `Enabled` pure.
type sampler struct {
	Facility

	tick       time.Duration
	counts     *counters
	first      uint64
	thereafter uint64
}

func (s *sampler) With(fields []Field) Facility {
	return &sampler{
		Facility:   s.Facility.With(fields),
		tick:       s.tick,
		counts:     s.counts,
		first:      s.first,
		thereafter: s.thereafter,
	}
}

func (s *sampler) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if !s.Enabled(ent.Level) {
		return ce
	}
	if n := s.counts.Inc(ent.Message); n > s.first {
		if n == s.first+1 {
			time.AfterFunc(s.tick, func() { s.counts.Reset(ent.Message) })
		}
		if (n-s.first)%s.thereafter != 0 {
			return ce
		}
	}
	return s.Facility.Check(ent, ce)
}
