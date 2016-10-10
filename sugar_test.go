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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSugarGetSugarFields(t *testing.T) {
	var (
		fields []Field
		err    error
	)

	_, err = getSugarFields("test")
	assert.Error(t, err, "Should return error on invalid number of arguments")

	_, err = getSugarFields("test1", 1, "test2")
	assert.Error(t, err, "Should return error on invalid number of arguments")

	_, err = getSugarFields("test1", 1, "error", errors.New(""))
	assert.Error(t, err, "Should return error when error passed as value (special case of unknown type)")

	_, err = getSugarFields(1, 1)
	assert.Error(t, err, "Should return error on non-string field name")

	fields, _ = getSugarFields("test", 1)
	assert.Len(t, fields, 1, "Should return 1 field")

	fields, _ = getSugarFields("test1", 1, "test2", 2)
	assert.Len(t, fields, 2, "Should return 2 fields")

	fields, _ = getSugarFields(errors.New("error"), "test1", 1)
	assert.Len(t, fields, 2, "Should return 2 fields")
}

func TestSugarLevel(t *testing.T) {
	assert.Equal(t, DebugLevel, NewSugar(New(NewJSONEncoder(), DebugLevel)).Level())
	assert.Equal(t, FatalLevel, NewSugar(New(NewJSONEncoder(), FatalLevel)).Level())
}

func TestSugarSetLevel(t *testing.T) {
	sugar := NewSugar(New(NewJSONEncoder()))
	sugar.SetLevel(FatalLevel)
	assert.Equal(t, FatalLevel, sugar.Level())
}

func withSugarLogger(t testing.TB, opts []Option, f func(Sugar, *testBuffer, *testBuffer)) {
	sink := &testBuffer{}
	errSink := &testBuffer{}

	allOpts := make([]Option, 0, 3+len(opts))
	allOpts = append(allOpts, DebugLevel, Output(sink), ErrorOutput(errSink))
	allOpts = append(allOpts, opts...)
	logger := New(newJSONEncoder(NoTime()), allOpts...)
	sugar := NewSugar(logger)

	f(sugar, sink, errSink)
}

func TestSugarLog(t *testing.T) {
	opts := opts(Fields(Int("foo", 42)))
	withSugarLogger(t, opts, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.Debug("debug message", "a", "b")
		logger.Info("info message", "c", "d")
		logger.Warn("warn message", "e", "f")
		logger.Error("error message", "g", "h")
		assert.Equal(t, []string{
			`{"level":"debug","msg":"debug message","foo":42,"a":"b"}`,
			`{"level":"info","msg":"info message","foo":42,"c":"d"}`,
			`{"level":"warn","msg":"warn message","foo":42,"e":"f"}`,
			`{"level":"error","msg":"error message","foo":42,"g":"h"}`,
		}, out.Lines(), "Incorrect output from logger")
	})
}

func TestSugarLogTypes(t *testing.T) {
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.Debug("")
		logger.Debug("", "bool", true)
		logger.Debug("", "float64", float64(1.23456789))
		logger.Debug("", "int", int(-1))
		logger.Debug("", "int64", int64(-1))
		logger.Debug("", "uint", uint(1))
		logger.Debug("", "uint64", uint64(1))
		logger.Debug("", "string", "")
		logger.Debug("", "time", time.Unix(0, 0))
		logger.Debug("", "duration", time.Second)
		logger.Debug("", "stringer", DebugLevel)
		logger.Debug("", "object", []string{"foo", "bar"})
		assert.Equal(t, []string{
			`{"level":"debug","msg":""}`,
			`{"level":"debug","msg":"","bool":true}`,
			`{"level":"debug","msg":"","float64":1.23456789}`,
			`{"level":"debug","msg":"","int":-1}`,
			`{"level":"debug","msg":"","int64":-1}`,
			`{"level":"debug","msg":"","uint":1}`,
			`{"level":"debug","msg":"","uint64":1}`,
			`{"level":"debug","msg":"","string":""}`,
			`{"level":"debug","msg":"","time":0}`,
			`{"level":"debug","msg":"","duration":1000000000}`,
			`{"level":"debug","msg":"","stringer":"debug"}`,
			`{"level":"debug","msg":"","object":["foo","bar"]}`,
		}, out.Lines(), "Incorrect output from logger")
	})
}

func TestSugarLogError(t *testing.T) {
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.Debug("with error", errors.New("this is a error"))
		assert.Equal(t, []string{
			`{"level":"debug","msg":"with error","error":"this is a error"}`,
		}, out.Lines(), "Incorrect output from logger")
	})
}

func TestSugarPanic(t *testing.T) {
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		assert.Panics(t, func() { logger.Panic("foo") }, "Expected logging at Panic level to panic.")
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, out.Stripped(), "Unexpected output from panic-level Log.")
	})
}

func TestSugarFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.Fatal("foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, out.Stripped(), "Unexpected output from fatal-level Log.")
		stub.AssertStatus(t, 1)
	})
}

func TestSugarDFatal(t *testing.T) {
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.DFatal("foo")
		assert.Equal(t, `{"level":"error","msg":"foo"}`, out.Stripped(), "Unexpected output from dfatal")
	})

	stub := stubExit()
	defer stub.Unstub()
	opts := opts(Development())
	withSugarLogger(t, opts, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		logger.DFatal("foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, out.Stripped(), "Unexpected output from DFatal in dev mode")
		stub.AssertStatus(t, 1)
	})
}

func TestSugarLogErrors(t *testing.T) {
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, err *testBuffer) {
		logger.Log(InfoLevel, "foo", "a")
		assert.Equal(t, `{"level":"info","msg":"foo"}`, out.Stripped(), "Should log invalid number of arguments")
		assert.Equal(t, `invalid number of arguments`, err.Stripped(), "Should log invalid number of arguments")
	})
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, err *testBuffer) {
		logger.Log(InfoLevel, "foo", 1, "foo")
		assert.Equal(t, `{"level":"info","msg":"foo"}`, out.Stripped(), "Should log invalid name type")
		assert.Equal(t, `field name must be string`, err.Stripped(), "Should log invalid name type")
	})
	withSugarLogger(t, nil, func(logger Sugar, out *testBuffer, err *testBuffer) {
		logger.Log(InfoLevel, "foo", "foo", errors.New("b"))
		assert.Equal(t, `{"level":"info","msg":"foo"}`, out.Stripped(), "Should log error argument position is invalid")
		assert.Equal(t, `error can only be the first argument`, err.Stripped(), "Should log error argument position is invalid")
	})
}

func TestSugarLogDiscards(t *testing.T) {
	withSugarLogger(t, opts(InfoLevel), func(logger Sugar, buf *testBuffer, _ *testBuffer) {
		logger.Debug("should be discarded")
		logger.Debug("should be discarded even with invalid arg count", "bla")
		logger.Info("should be logged")
		assert.Equal(t, []string{
			`{"level":"info","msg":"should be logged"}`,
		}, buf.Lines(), "")
	})
}

func TestSugarWith(t *testing.T) {
	opts := opts()
	withSugarLogger(t, opts, func(logger Sugar, out *testBuffer, _ *testBuffer) {
		child := logger.With("a", "b")
		child.Debug("debug message", "c", "d")
		assert.Equal(t, []string{
			`{"level":"debug","msg":"debug message","a":"b","c":"d"}`,
		}, out.Lines(), "Incorrect output from logger")
	})
}

func TestSugarWithErrors(t *testing.T) {
	opts := opts()
	withSugarLogger(t, opts, func(logger Sugar, _ *testBuffer, err *testBuffer) {
		logger.With("a")
		assert.Equal(t, "invalid number of arguments", err.Stripped(), "Should log invalid number of arguments")
	})
}
