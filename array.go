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
	"go.uber.org/zap/internal/multierror"
	"go.uber.org/zap/zapcore"
)

// TODO: actually add an example showing off how to use these wrappers with the
// Array field constructor.

// Arrays wraps a slice of ArrayMarshalers so that it satisfies the
// ArrayMarshaler interface. See the Array function for a usage example.
func Arrays(as []zapcore.ArrayMarshaler) zapcore.ArrayMarshaler {
	return arrayMarshalers(as)
}

// Objects wraps a slice of ObjectMarshalers so that it satisfies the
// ArrayMarshaler interface. See the Array function for a usage example.
func Objects(os []zapcore.ObjectMarshaler) zapcore.ArrayMarshaler {
	return objectMarshalers(os)
}

// Bools wraps a slice of bools so that it satisfies the ArrayMarshaler
// interface. See the Array function for a usage example.
func Bools(bs []bool) zapcore.ArrayMarshaler {
	return bools(bs)
}

type arrayMarshalers []zapcore.ArrayMarshaler

func (as arrayMarshalers) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	var errs *multierror.Error
	for i := range as {
		if as[i] == nil {
			continue
		}
		errs = errs.Append(arr.AppendArray(as[i]))
	}
	return errs.AsError()
}

type objectMarshalers []zapcore.ObjectMarshaler

func (os objectMarshalers) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	var errs *multierror.Error
	for i := range os {
		if os[i] == nil {
			continue
		}
		errs = errs.Append(arr.AppendObject(os[i]))
	}
	return errs.AsError()
}

type bools []bool

func (bs bools) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range bs {
		arr.AppendBool(bs[i])
	}
	return nil
}
