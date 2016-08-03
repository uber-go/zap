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

import "time"

// JSONOption is used to set options for a JSON encoder. MessageFormatters,
// TimeFormatters, and LevelFormatters all implement the JSONOption interface.
type JSONOption interface {
	apply(*jsonEncoder)
}

// A MessageFormatter defines how to convert a log message into a Field.
// MessageFormatters implement the JSONOption interface.
type MessageFormatter func(string) Field

func (mf MessageFormatter) apply(enc *jsonEncoder) {
	enc.messageF = mf
}

// MessageKey encodes log messages under the provided key.
func MessageKey(key string) MessageFormatter {
	return MessageFormatter(func(msg string) Field {
		return String(key, msg)
	})
}

// A TimeFormatter defines how to convert the time of a log entry into a Field.
// TimeFormatters implement the JSONOption interface.
type TimeFormatter func(time.Time) Field

func (tf TimeFormatter) apply(enc *jsonEncoder) {
	enc.timeF = tf
}

// EpochFormatter uses the Time field (floating-point seconds since epoch) to
// encode the entry time under the provided key.
func EpochFormatter(key string) TimeFormatter {
	return TimeFormatter(func(t time.Time) Field {
		return Time(key, t)
	})
}

// RFC3339Formatter encodes the entry time as an RFC3339-formatted string under
// the provided key.
func RFC3339Formatter(key string) TimeFormatter {
	return TimeFormatter(func(t time.Time) Field {
		return String(key, t.Format(time.RFC3339))
	})
}

// NoTime drops the entry time altogether. It's often useful in testing, since
// it removes the need to stub time.Now.
func NoTime() TimeFormatter {
	return TimeFormatter(func(_ time.Time) Field {
		return Skip()
	})
}

// A LevelFormatter defines how to convert an entry's logging level into a
// Field. LevelFormatters implement the JSONOption interface.
type LevelFormatter func(Level) Field

func (lf LevelFormatter) apply(enc *jsonEncoder) {
	enc.levelF = lf
}

// LevelString encodes the entry's level under the provided key. It uses the
// level's String method to serialize it.
func LevelString(key string) LevelFormatter {
	return LevelFormatter(func(l Level) Field {
		return String(key, l.String())
	})
}
