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
	"io"
	"os"
	"sync"
	"time"
)

var _entryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{}
	},
}

// Meta is implementation-agnostic state management for Loggers. Most Logger
// implementations can reduce the required boilerplate by embedding a Meta.
//
// Note that while the level-related fields and methods are safe for concurrent
// use, the remaining fields are not.
type Meta struct {
	LevelEnabler

	Development bool
	Encoder     Encoder
	Hooks       []Hook
	Output      WriteSyncer
	ErrorOutput WriteSyncer
}

// MakeMeta returns a new meta struct with sensible defaults: logging at
// InfoLevel, development mode off, and writing to standard error and standard
// out.
func MakeMeta(enc Encoder, options ...Option) Meta {
	m := Meta{
		Encoder:      enc,
		Output:       newLockedWriteSyncer(os.Stdout),
		ErrorOutput:  newLockedWriteSyncer(os.Stderr),
		LevelEnabler: InfoLevel,
	}
	for _, opt := range options {
		opt.apply(&m)
	}
	return m
}

// Clone creates a copy of the meta struct. It deep-copies the encoder, but not
// the hooks (since they rarely change).
func (m Meta) Clone() Meta {
	m.Encoder = m.Encoder.Clone()
	return m
}

// Check returns a CheckedMessage logging the given message is Enabled, nil
// otherwise.
func (m Meta) Check(log Logger, lvl Level, msg string) *CheckedMessage {
	switch lvl {
	case PanicLevel, FatalLevel:
		// Panic and Fatal should always cause a panic/exit, even if the level
		// is disabled.
		break
	default:
		if !m.Enabled(lvl) {
			return nil
		}
	}
	return NewCheckedMessage(log, lvl, msg)
}

// InternalError prints an internal error message to the configured
// ErrorOutput. This method should only be used to report internal logger
// problems and should not be used to report user-caused problems.
func (m Meta) InternalError(cause string, err error) {
	fmt.Fprintf(m.ErrorOutput, "%v %s error: %v\n", time.Now().UTC(), cause, err)
	m.ErrorOutput.Sync()
}

// Encode runs any Hook functions and then writes an encoded log entry to the
// given io.Writer, returning any error.
func (m Meta) Encode(w io.Writer, t time.Time, lvl Level, msg string, fields []Field) error {
	enc := m.Encoder.Clone()
	addFields(enc, fields)
	if len(m.Hooks) >= 0 {
		entry := _entryPool.Get().(*Entry)
		entry.Level = lvl
		entry.Message = msg
		entry.Time = t
		entry.enc = enc
		for _, hook := range m.Hooks {
			if err := hook(entry); err != nil {
				m.InternalError("hook", err)
			}
		}
		msg, enc = entry.Message, entry.enc
		_entryPool.Put(entry)
	}
	err := enc.WriteEntry(w, msg, lvl, t)
	enc.Free()
	return err
}
