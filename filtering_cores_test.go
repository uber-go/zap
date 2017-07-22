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
	"testing"

	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFieldFilteringCore(t *testing.T) {
	next, logs := observer.New(zapcore.ErrorLevel)

	const skipKey = "skip"
	skipField := Bool(skipKey, true)

	core := NewFieldFilteringCore(next, func(field zapcore.Field) bool {
		return field.Key != skipKey
	})

	New(core).Error("nuked", String("a", "b"), String("b", "c"), skipField)

	fields := logs.All()[0].Context
	if n := len(fields); n != 2 {
		t.Errorf("unexpected number of entry fields: expected 2, got %v", n)
	}
	for _, field := range fields {
		if field.Key == skipKey {
			t.Error("the field supposedly filtered out was actually not filtered out")
		}
	}
}
