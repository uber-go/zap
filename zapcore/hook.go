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

import "go.uber.org/zap/internal/multierror"

type hooked struct {
	Facility
	funcs []func(Entry) error
}

// Hooked wraps a facility and runs a collection of user-defined callback hooks
// each time a message is logged.
func Hooked(fac Facility, hooks ...func(Entry) error) Facility {
	funcs := append([]func(Entry) error{}, hooks...)
	return &hooked{
		Facility: fac,
		funcs:    funcs,
	}
}

func (h *hooked) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if downstream := h.Facility.Check(ent, ce); downstream != nil {
		return downstream.AddFacility(ent, h)
	}
	return ce
}

func (h *hooked) With(fields []Field) Facility {
	return &hooked{
		Facility: h.Facility.With(fields),
		funcs:    h.funcs,
	}
}

func (h *hooked) Write(ent Entry, _ []Field) error {
	var errs multierror.Error
	for i := range h.funcs {
		errs = errs.Append(h.funcs[i](ent))
	}
	return errs.AsError()
}
