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

func TestSugarGetSugarFields(t *testing.T) {
	// logger := NewSugar(New(NewJSONEncoder()))
	// assert.Error(t, logger.getFields("test"), )

	var (
		fields []Field
		err    error
	)

	_, err = getSugarFields("test")
	assert.Error(t, err, "Should return error on invalid number of arguments")

	_, err = getSugarFields("test1", 1, "test2")
	assert.Error(t, err, "Should return error on invalid number of arguments")

	_, err = getSugarFields("test1", nil)
	assert.Error(t, err, "Should return error on argument of unknown type")

	_, err = getSugarFields(1, 1)
	assert.Error(t, err, "Should return error on non-string field name")

	fields, _ = getSugarFields("test", 1)
	assert.Equal(t, 1, len(fields), "Should return 1 field")

	fields, _ = getSugarFields("test1", 1, "test2", 2)
	assert.Equal(t, 2, len(fields), "Should return 2 fields")
}

func withSugarLogger(t testing.TB, opts []Option, f func(Sugar, *testBuffer)) {
	sink := &testBuffer{}
	errSink := &testBuffer{}

	allOpts := make([]Option, 0, 3+len(opts))
	allOpts = append(allOpts, DebugLevel, Output(sink), ErrorOutput(errSink))
	allOpts = append(allOpts, opts...)
	logger := New(newJSONEncoder(NoTime()), allOpts...)
	sugar := NewSugar(logger)

	f(sugar, sink)
}

func TestSugarLog(t *testing.T) {
	opts := opts(Fields(Int("foo", 42)))
	withSugarLogger(t, opts, func(logger Sugar, buf *testBuffer) {
		logger.Debug("debug message", "a", "b")
		logger.Info("info message", "c", "d")
		logger.Warn("warn message", "e", "f")
		logger.Error("error message", "g", "h")
		assert.Equal(t, []string{
			`{"level":"debug","msg":"debug message","foo":42,"a":"b"}`,
			`{"level":"info","msg":"info message","foo":42,"c":"d"}`,
			`{"level":"warn","msg":"warn message","foo":42,"e":"f"}`,
			`{"level":"error","msg":"error message","foo":42,"g":"h"}`,
		}, buf.Lines(), "Incorrect output from logger")
	})
}

func TestSugarWith(t *testing.T) {
	opts := opts()
	withSugarLogger(t, opts, func(logger Sugar, buf *testBuffer) {
		child, _ := logger.With("a", "b")
		child.Debug("debug message", "c", "d")
		assert.Equal(t, []string{
			`{"level":"debug","msg":"debug message","a":"b","c":"d"}`,
		}, buf.Lines(), "Incorrect output from logger")
	})
}
