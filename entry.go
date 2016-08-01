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
	"sync"
	"time"
)

var (
	_timeNow   = time.Now // for tests
	_entryPool = sync.Pool{New: func() interface{} { return &Entry{} }}
)

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
	enc     Encoder
}

func newEntry(lvl Level, msg string, enc Encoder) *Entry {
	e := _entryPool.Get().(*Entry)
	e.Level = lvl
	e.Message = msg
	e.Time = _timeNow().UTC()
	e.enc = enc
	return e
}

// Fields returns a mutable reference to the entry's accumulated context.
func (e *Entry) Fields() KeyValue {
	return e.enc
}

func (e *Entry) free() {
	_entryPool.Put(e)
}
