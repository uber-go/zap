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
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap/internal/buffers"
	"go.uber.org/zap/internal/exit"
)

var (
	_cePool = sync.Pool{New: func() interface{} {
		// Pre-allocate some space for facilities.
		return &CheckedEntry{
			facs: make([]Facility, 4),
		}
	}}
)

func getCheckedEntry() *CheckedEntry {
	ce := _cePool.Get().(*CheckedEntry)
	ce.reset()
	return ce
}

func putCheckedEntry(ce *CheckedEntry) {
	if ce == nil {
		return
	}
	_cePool.Put(ce)
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
	buf := buffers.Get()
	buf = append(buf, ec.File...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(ec.Line), 10)
	caller := string(buf)
	buffers.Put(buf)
	return caller
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
}

// CheckWriteAction indicates what action to take after (*CheckedEntry).Write
// is done. Actions are ordered in increasing severity.
type CheckWriteAction uint8

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
	ErrorOutput WriteSyncer
	dirty       bool // best-effort detection of pool misuse
	should      CheckWriteAction
	facs        []Facility
}

func (ce *CheckedEntry) reset() {
	ce.Entry = Entry{}
	ce.ErrorOutput = nil
	ce.dirty = false
	ce.should = WriteThenNoop
	ce.facs = ce.facs[:0]
}

// Write writes the entry to any Facility references stored, returning any
// errors, and returns the CheckedEntry reference to a pool for immediate
// re-use. Write finally does any panic or fatal exiting that should happen.
func (ce *CheckedEntry) Write(fields ...Field) {
	if ce == nil {
		return
	}

	if ce.dirty {
		if ce.ErrorOutput != nil {
			// this races with (*CheckedEntry).AddFacility; recover a log entry for
			// debugging, which may be an amalgamation of both the prior one
			// returned to pool, and any new one being set
			ent := ce.Entry
			fmt.Fprintf(ce.ErrorOutput, "%v race-prone write dropped circa log entry %+v\n", time.Now().UTC(), ent)
			ce.ErrorOutput.Sync()
		}
		return
	}
	ce.dirty = true

	var errs multiError
	for i := range ce.facs {
		if err := ce.facs[i].Write(ce.Entry, fields); err != nil {
			errs = append(errs, err)
		}
	}
	if ce.ErrorOutput != nil {
		if err := errs.asError(); err != nil {
			fmt.Fprintf(ce.ErrorOutput, "%v write error: %v\n", time.Now().UTC(), err)
			ce.ErrorOutput.Sync()
		}
	}

	should, msg := ce.should, ce.Message
	putCheckedEntry(ce)

	switch should {
	case WriteThenFatal:
		exit.Exit()
	case WriteThenPanic:
		panic(msg)
	}
}

// AddFacility adds a facility that has agreed to log this entry. It's intended
// to be used by Facility.Check implementations. If ce is nil then a new
// CheckedEntry is created. Returns a non-nil CheckedEntry, maybe just created.
func (ce *CheckedEntry) AddFacility(ent Entry, fac Facility) *CheckedEntry {
	if ce != nil {
		ce.facs = append(ce.facs, fac)
		return ce
	}
	ce = getCheckedEntry()
	ce.Entry = ent
	ce.facs = append(ce.facs, fac)
	return ce
}

// Should sets state so that a panic or fatal exit will happen after Write is
// called. Similarly to AddFacility, if ce is nil then a new CheckedEntry is
// built to record the intent to panic or fatal (this is why the caller must
// provide an Entry value, since ce may be nil).
func (ce *CheckedEntry) Should(ent Entry, should CheckWriteAction) *CheckedEntry {
	if ce != nil {
		ce.should = should
		return ce
	}
	ce = getCheckedEntry()
	ce.Entry = ent
	ce.should = should
	return ce
}
