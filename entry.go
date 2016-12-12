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
	"fmt"
	"sync"
	"time"
)

var _cePool = sync.Pool{
	New: func() interface{} {
		return &CheckedEntry{}
	},
}

// MakeEntryCaller makes an EntryCaller from the return signature of
// runtime.Caller().
func MakeEntryCaller(pc uintptr, file string, line int, ok bool) EntryCaller {
	if !ok {
		return EntryCaller{}
	}
	return EntryCaller{
		PC:      pc,
		File:    file,
		Line:    line,
		Defined: true,
	}
}

// EntryCaller represents a notable caller of a log entry.
type EntryCaller struct {
	Defined bool
	PC      uintptr
	File    string
	Line    int
}

// String returns a "file:line" string if the EntryCaller is Defined, and the
// empty string otherwise.
func (ec EntryCaller) String() string {
	if !ec.Defined {
		return ""
	}
	return fmt.Sprintf("%s:%d", ec.File, ec.Line)
}

// An Entry represents a complete log message. The entry's structured context
// is already serialized, but the log level, time, and message are available
// for inspection and modification.
//
// Entries are pooled, so any functions that accept them must be careful not to
// retain references to them.
type Entry struct {
	Level   Level
	Time    time.Time
	Message string
	Caller  EntryCaller
	Stack   string

	enc Encoder
}

// Fields returns a mutable reference to the entry's accumulated context.
func (e Entry) Fields() KeyValue {
	return e.enc
}

// CheckWriteAction indicates what action to take after (*CheckedEntry).Write
// is done.
type CheckWriteAction int

const (
	// WriteThenNoop is the default behavior to do nothing speccial after write.
	WriteThenNoop = CheckWriteAction(iota)
	// WriteThenFatal causes a fatal os.Exit() after Write.
	WriteThenFatal
	// WriteThenPanic causes a panic() after Write.
	WriteThenPanic
)

// CheckedEntry is an Entry together with an opaque Facility that has already
// agreed to log it (Facility.Enabled(Entry) == true). It is returned by
// Logger.Check to enable performance sensitive log sites to not allocate
// fields when disabled.
//
// CheckedEntry references should be created by calling AddFacility() or
// Should() on a nil *CheckedEntry. References are returned to a pool after
// Write, and MUST NOT be retained after calling their Write() method.
type CheckedEntry struct {
	Entry
	should CheckWriteAction
	facs   []Facility
}

// Write writes the entry to any Facility references stored, returning any
// errors, and returns the CheckedEntry reference to a pool for immediate
// re-use.
func (ce *CheckedEntry) Write(fields ...Field) error {
	if ce == nil {
		return nil
	}
	var errs multiError
	for i := range ce.facs {
		if err := ce.facs[i].Log(ce.Entry, fields...); err != nil {
			errs = append(errs, err)
		}
	}

	should, msg := ce.should, ce.Message
	ce.should = WriteThenNoop
	ce.facs = ce.facs[:0]
	_cePool.Put(ce)

	switch should {
	case WriteThenFatal:
		_exit(1)
	case WriteThenPanic:
		panic(msg)
	}

	return errs.asError()
}

// AddFacility adds a facility that has agreed to log this entry. It's intended
// to be used by Facility.Check implementations. If ce is nil then a new
// CheckedEntry is created. Returns a non-nil CheckedEntry, maybe just created.
func (ce *CheckedEntry) AddFacility(ent Entry, fac Facility) *CheckedEntry {
	if ce == nil {
		ce = _cePool.Get().(*CheckedEntry)
		ce.Entry = ent
	}
	ce.facs = append(ce.facs, fac)
	// TODO: we could provide static spac for the first N facilities to avoid
	// allocations in common cases
	return ce
}

// Should sets state so that a panic or fatal exit will happen after Write is
// called. Similarly to AddFacility, if ce is nil then a now CheckedEntry is
// built to record the intent to panic or fatal (this is why the caller must
// provide an Entry value, since ce may be nil).
func (ce *CheckedEntry) Should(ent Entry, should CheckWriteAction) *CheckedEntry {
	if ce == nil {
		ce = _cePool.Get().(*CheckedEntry)
		ce.Entry = ent
	}
	ce.should = should
	return ce
}
