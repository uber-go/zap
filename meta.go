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
	"os"
	"sync"
)

// Meta is implementation-agnostic state management for Loggers. Most Logger
// implementations can reduce the required boilerplate by embedding a Meta.
//
// Note that while the level-related fields and methods are safe for concurrent
// use, the remaining fields are not.
type Meta struct {
	Development bool
	Encoder     Encoder
	Hooks       []Hook
	Output      WriteSyncer
	ErrorOutput WriteSyncer

	enabLock *sync.RWMutex
	enab     *LevelEnabler
}

// MakeMeta returns a new meta struct with sensible defaults: logging at
// InfoLevel, development mode off, and writing to standard error and standard
// out.
func MakeMeta(enc Encoder, options ...Option) Meta {
	enb := LevelEnabler(InfoLevel)
	m := Meta{
		Encoder:     enc,
		Output:      newLockedWriteSyncer(os.Stdout),
		ErrorOutput: newLockedWriteSyncer(os.Stderr),
		enabLock:    &sync.RWMutex{},
		enab:        &enb,
	}
	for _, opt := range options {
		opt.apply(&m)
	}
	return m
}

// Level returns the minimum enabled log level, if simple monotonic level
// enabling is being used; otherwise it returns FatalLevel+1.
func (m Meta) Level() Level {
	m.enabLock.RLock()
	if lvl, ok := (*m.enab).(Level); ok {
		m.enabLock.RUnlock()
		return lvl
	}
	m.enabLock.RUnlock()
	return FatalLevel + 1
}

// SetLevel atomically alters the level enabler so that all logs of a given
// level and up are enabled.
func (m Meta) SetLevel(lvl Level) {
	m.enabLock.Lock()
	*m.enab = lvl
	m.enabLock.Unlock()
}

// Clone creates a copy of the meta struct. It deep-copies the encoder, but not
// the hooks (since they rarely change).
func (m Meta) Clone() Meta {
	m.Encoder = m.Encoder.Clone()
	return m
}

// Enabled returns true if logging a message at a particular level is enabled.
func (m Meta) Enabled(lvl Level) bool {
	m.enabLock.RLock()
	e := (*m.enab).Enabled(lvl)
	m.enabLock.RUnlock()
	return e
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
