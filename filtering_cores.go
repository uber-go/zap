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
	"go.uber.org/zap/zapcore"
)

// FilterFunc is used to check whether to filter the given field out.
type FilterFunc func(zapcore.Field) bool

type fieldFilteringCore struct {
	zapcore.Core
	filter FilterFunc
}

// NewFieldFilteringCore returns a core that uses the given filter function
// to filter fields before passing them to the core being wrapped.
func NewFieldFilteringCore(next zapcore.Core, filter FilterFunc) zapcore.Core {
	return &fieldFilteringCore{next, filter}
}

func (core *fieldFilteringCore) With(fields []zapcore.Field) zapcore.Core {
	filteredFields := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		if core.filter(field) {
			filteredFields = append(filteredFields, field)
		}
	}

	return NewFieldFilteringCore(core.Core.With(filteredFields), core.filter)
}

func (core *fieldFilteringCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	filteredFields := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		if core.filter(field) {
			filteredFields = append(filteredFields, field)
		}
	}

	return core.Core.Write(entry, filteredFields)
}
