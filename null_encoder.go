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
	"io"
	"time"
)

// nullEncoder is an Encoder implementation that throws everything away.
type nullEncoder struct{}

var _nullEncoder nullEncoder

// NullEncoder returns the fast, no-allocation encoder.
func NullEncoder() Encoder {
	return _nullEncoder
}

func (nullEncoder) Free() {}

func (nullEncoder) AddString(_, _ string)          {}
func (nullEncoder) AddBool(_ string, _ bool)       {}
func (nullEncoder) AddInt(_ string, _ int)         {}
func (nullEncoder) AddInt64(_ string, _ int64)     {}
func (nullEncoder) AddUint(_ string, _ uint)       {}
func (nullEncoder) AddUint64(_ string, _ uint64)   {}
func (nullEncoder) AddUintptr(_ string, _ uintptr) {}
func (nullEncoder) AddFloat64(_ string, _ float64) {}

func (nullEncoder) AddMarshaler(_ string, _ LogMarshaler) error { return nil }
func (nullEncoder) AddObject(_ string, _ interface{}) error     { return nil }

// Clone copies the current encoder, including any data already encoded.
func (nullEncoder) Clone() Encoder {
	return _nullEncoder
}

// WriteEntry writes nothing to the supplied writer, but demands a valid writer.
// It's safe to call from multiple goroutines.
func (nullEncoder) WriteEntry(sink io.Writer, _ string, _ Level, _ time.Time) error {
	if sink == nil {
		return errNilSink
	}
	return nil
}
