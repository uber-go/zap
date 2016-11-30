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

import "sync"

var _cmPool = sync.Pool{
	New: func() interface{} {
		return &CheckedMessage{
			lvl: invalidLevel,
		}
	},
}

// A CheckedMessage is the result of a call to Logger.Check, which allows
// especially performance-sensitive applications to avoid allocations for disabled
// or heavily sampled log levels.
type CheckedMessage struct {
	logger      Logger
	safeToWrite bool
	lvl         Level
	msg         string

	// singly linked list built by Chain
	next *CheckedMessage // carried by each part of Chain-ed list
	tail *CheckedMessage // non-nil only in the head of a Chain-ed list
	// NOTE: suffixes are NOT valid lists, since they lack a tail pointer; tail
	// pointer exists solely as an optimization for calling head.Chain.
}

// NewCheckedMessage constructs a CheckedMessage. It's only intended for use by
// wrapper libraries, and shouldn't be necessary in application code.
func NewCheckedMessage(logger Logger, lvl Level, msg string) *CheckedMessage {
	m := _cmPool.Get().(*CheckedMessage)
	m.safeToWrite, m.logger, m.lvl, m.msg = true, logger, lvl, msg
	return m
}

// Write logs the pre-checked message with the supplied fields. It will call
// the underlying level method (Debug, Info, Warn, Error, DPanic, Panic, and
// Fatal) for the defined levels; the Log method is only called for unknown
// logging levels.
//
// It MUST be called at most once, since Write will return the *CheckedMessage
// to an internal pool for potentially immediate re-use; re-using a
// *CheckedMessage after calling Write() will result in data races or other
// undefined behavior. An attempt is made to detect and DPanic log any re-use,
// but such detection is not guaranteed due to race conditions.
func (m *CheckedMessage) Write(fields ...Field) {
	if m == nil {
		return
	}

	if !m.safeToWrite {
		// we're living in racy times, so copy what we can out of the pointer
		// that we have, and at least tell the user something
		if logger := m.logger; logger != nil {
			lvl, msg := m.lvl, m.msg
			logger.DPanic(
				"Must not call zap.(*CheckedMessage).Write() more than once",
				Nest("prior", Stringer("level", lvl), String("msg", msg)),
			)
		}
		return
	}
	m.safeToWrite = false

	switch m.lvl {
	case DebugLevel:
		m.logger.Debug(m.msg, fields...)
	case InfoLevel:
		m.logger.Info(m.msg, fields...)
	case WarnLevel:
		m.logger.Warn(m.msg, fields...)
	case ErrorLevel:
		m.logger.Error(m.msg, fields...)
	case PanicLevel:
		m.logger.Panic(m.msg, fields...)
	case FatalLevel:
		m.logger.Fatal(m.msg, fields...)
	default:
		m.logger.Log(m.lvl, m.msg, fields...)
	}

	m.next.Write(fields...)
	m.next, m.tail = nil, nil
	_cmPool.Put(m)
}

// Chain combines two or more CheckedMessages. If the receiver message is not
// OK(), the passed message is returned. Otherwise if the passed message is
// OK(), then it is retained such that its Write() will be called after the
// receiver's Write(), and any previously Chain()'ed messages already so
// retained.
func (m *CheckedMessage) Chain(ms ...*CheckedMessage) *CheckedMessage {
	for _, m2 := range ms {
		if !m2.OK() {
			continue
		}
		if m.OK() {
			m.push(m2)
		} else {
			m = m2
		}
	}
	return m
}

func (m *CheckedMessage) push(next *CheckedMessage) {
	if m.tail != nil {
		m.tail.next = next
	} else if m.next != nil {
		m.logger.DPanic("invalid CheckedMessage linked list; did we lose our head?", String("original", m.msg))
	} else {
		m.next = next
	}
	if next.tail != nil {
		m.tail, next.tail = next.tail, nil
	} else {
		m.tail = next
	}
}

// OK checks whether it's safe to call Write.
func (m *CheckedMessage) OK() bool {
	return m != nil
}
