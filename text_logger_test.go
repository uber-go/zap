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

	"github.com/stretchr/testify/assert"
)

func withTextLogger(t testing.TB, opts []Option, f func(Logger, *testBuffer)) {
	sink := &testBuffer{}
	errSink := &testBuffer{}

	allOpts := make([]Option, 0, 3+len(opts))
	allOpts = append(allOpts, DebugLevel, Output(sink), ErrorOutput(errSink))
	allOpts = append(allOpts, opts...)
	logger := New(newTextEncoder(TextNoTime()), allOpts...)

	f(logger, sink)
	assert.Empty(t, errSink.String(), "Expected error sink to be empty.")
}

func TestTextLoggerDebugLevel(t *testing.T) {
	withTextLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Log(DebugLevel, "foo")
		assert.Equal(t, "[D] foo", buf.Stripped(), "Unexpected output from logger")
	})
}

func TestTextLoggerNestedMarshal(t *testing.T) {
	m := LogMarshalerFunc(func(kv KeyValue) error {
		kv.AddString("loggable", "yes")
		kv.AddInt("number", 1)
		return nil
	})

	withTextLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Info("Fields", String("f1", "{"), Marshaler("m", m))
		assert.Equal(t, "[I] Fields f1={ m={loggable=yes number=1}", buf.Stripped(), "Unexpected output from logger")
	})
}

func TestTextLoggerAddMarshalEmpty(t *testing.T) {
	empty := LogMarshalerFunc(func(_ KeyValue) error { return nil })
	withTextLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Info("Empty", Marshaler("m", empty), String("something", "val"))
		assert.Equal(t, "[I] Empty m={} something=val", buf.Stripped(), "Unexpected output from logger")
	})
}
