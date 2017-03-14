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

// Package multierror provides a simple way to treat a collection of errors as
// a single error.
package multierror

import "go.uber.org/zap/internal/bufferpool"

// implement the standard lib's error interface on a private type so that we
// can't forget to call Error.AsError().
type errSlice []error

func (es errSlice) Error() string {
	b := bufferpool.Get()
	for i, err := range es {
		if i > 0 {
			b.AppendByte(';')
			b.AppendByte(' ')
		}
		b.AppendString(err.Error())
	}
	ret := b.String()
	b.Free()
	return ret
}

// Error wraps a []error to implement the error interface.
type Error struct {
	errs errSlice
}

// AsError converts the collection to a single error value.
//
// Note that failing to use AsError will almost certainly lead to bugs with
// non-nil interfaces containing nil concrete values.
func (e Error) AsError() error {
	switch len(e.errs) {
	case 0:
		return nil
	case 1:
		return e.errs[0]
	default:
		return e.errs
	}
}

// Append adds an error to the collection. Adding a nil error is a no-op.
func (e Error) Append(err error) Error {
	if err == nil {
		return e
	}
	e.errs = append(e.errs, err)
	return e
}
