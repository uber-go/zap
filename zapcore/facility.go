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
	"fmt"
	"io"
	"sync"
	"time"

	"go.uber.org/zap/internal/buffers"

	"go.uber.org/atomic"
)

// Facility is a destination for log entries. It can have pervasive fields
// added with With().
type Facility interface {
	LevelEnabler

	With([]Field) Facility
	Check(Entry, *CheckedEntry) *CheckedEntry
	Write(Entry, []Field) error
}

type nopFacility struct{}

// NopFacility returns a no-op Facility.
func NopFacility() Facility { return nopFacility{} }

func (nopFacility) Enabled(Level) bool { return false }

func (n nopFacility) With([]Field) Facility { return n }

func (n nopFacility) Check(_ Entry, ce *CheckedEntry) *CheckedEntry { return ce }

func (nopFacility) Write(Entry, []Field) error { return nil }

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
	buf, err := iof.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	err = checkPartialWrite(iof.out, buf)
	buffers.Put(buf)
	if err != nil {
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

// checkPartialWrite writes to an io.Writer, and upgrades partial writes to an
// error if no other write error occured.
func checkPartialWrite(w io.Writer, buf []byte) error {
	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, len(buf))
	}
	return nil
}
