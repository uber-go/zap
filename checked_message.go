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

import "github.com/uber-go/atomic"

// A CheckedMessage is the result of a call to Logger.Check, which allows
// especially performance-sensitive applications to avoid allocations for disabled
// or heavily sampled log levels.
type CheckedMessage struct {
	logger Logger
	used   atomic.Uint32
	lvl    Level
	msg    string

	// singly linked list built by Chain
	next *CheckedMessage // carried by each part of Chain-ed list
	tail *CheckedMessage // non-nil only in the head of a Chain-ed list
	// NOTE: suffixes are NOT valid lists, since they lack a tail pointer; tail
	// pointer exist solely as an optimization for calling head.Chain.
}

// NewCheckedMessage constructs a CheckedMessage. It's only intended for use by
// wrapper libraries, and shouldn't be necessary in application code.
func NewCheckedMessage(logger Logger, lvl Level, msg string) *CheckedMessage {
	return &CheckedMessage{
		logger: logger,
		lvl:    lvl,
		msg:    msg,
	}
}

// Write logs the pre-checked message with the supplied fields. It should only
// be used once; if a CheckedMessage is re-used, it also logs an error message
// with the underlying logger's DFatal method.
//
// Write will call the underlying level method (Debug, Info, Warn, Error,
// Panic, and Fatal) for the defined levels; the Log method is only called for
// unknown logging levels.
func (m *CheckedMessage) Write(fields ...Field) {
	if m == nil {
		return
	}

	if n := m.used.Inc(); n > 1 {
		if n == 2 {
			// Log an error on the first re-use. After that, skip the I/O and
			// allocations and just return.
			m.logger.DFatal("Shouldn't re-use a CheckedMessage.", String("original", m.msg))
		}
		return
	}

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

	if m.next != nil {
		m.next.Write(fields...)
	}
}

// Chain builds a chain of checked messages over several loggers. It returns an
// OK message if either the prior message, or the new one returned by
// logger.Check is OK.
//
// The returned message will first Write to any prior loggers before logging to
// the newly passed logger.
func (m *CheckedMessage) Chain(logger Logger, lvl Level, msg string) *CheckedMessage {
	if m == nil {
		return logger.Check(lvl, msg)
	}
	if next := NewCheckedMessage(logger, lvl, msg); next.OK() {
		m.push(next)
	}
	return m
}

func (m *CheckedMessage) push(next *CheckedMessage) {
	if m.tail != nil {
		m.tail.next = next
	} else if m.next != nil {
		panic("invalid CheckedMessage linked list; did we loose our head?")
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
