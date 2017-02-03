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

package zapcore_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.uber.org/zap/zapcore"
)

var (
	_epoch     = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	_testEntry = Entry{
		LoggerName: "main",
		Level:      InfoLevel,
		Message:    `hello`,
		Time:       _epoch,
		Stack:      "fake-stack",
		Caller:     EntryCaller{Defined: true, File: "foo.go", Line: 42},
	}
)

func testEncoderConfig() EncoderConfig {
	return EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "name",
		TimeKey:        "ts",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeTime:     func(t time.Time, enc ArrayEncoder) { enc.AppendInt64(t.UnixNano() / int64(time.Millisecond)) },
		EncodeLevel:    func(l Level, enc ArrayEncoder) { enc.AppendString(l.String()) },
		EncodeDuration: func(d time.Duration, enc ArrayEncoder) { enc.AppendInt64(int64(d)) },
	}
}

func humanEncoderConfig() EncoderConfig {
	cfg := testEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc ArrayEncoder) { enc.AppendString(t.Format(time.RFC3339)) }
	cfg.EncodeLevel = func(l Level, enc ArrayEncoder) { enc.AppendString(strings.ToUpper(l.String())) }
	cfg.EncodeDuration = func(d time.Duration, enc ArrayEncoder) { enc.AppendString(d.String()) }
	return cfg
}

func withJSONEncoder(f func(Encoder)) {
	f(NewJSONEncoder(testEncoderConfig()))
}

func withConsoleEncoder(f func(Encoder)) {
	f(NewConsoleEncoder(humanEncoderConfig()))
}

func TestEncoderConfiguration(t *testing.T) {
	base := testEncoderConfig()

	tests := []struct {
		desc            string
		cfg             EncoderConfig
		extra           func(Encoder)
		expectedJSON    string
		expectedConsole string
	}{
		{
			desc: "use custom entry keys in JSON output and ignore them in console output",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tmain@foo.go:42\thello\nfake-stack",
		},
		{
			desc: "skip level if LevelKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
			expectedConsole: "0\tmain@foo.go:42\thello\nfake-stack",
		},
		{
			desc: "skip timestamp if TimeKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
			expectedConsole: "info\tmain@foo.go:42\thello\nfake-stack",
		},
		{
			desc: "skip message if MessageKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","T":0,"N":"main","C":"foo.go:42","S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tmain@foo.go:42\nfake-stack",
		},
		{
			desc: "skip name if NameKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","T":0,"C":"foo.go:42","M":"hello","S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tfoo.go:42\thello\nfake-stack",
		},
		{
			desc: "skip caller if CallerKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","T":0,"N":"main","M":"hello","S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tmain\thello\nfake-stack",
		},
		{
			desc: "skip stacktrace if StacktraceKey is omitted",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			expectedJSON:    `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello"}`,
			expectedConsole: "0\tinfo\tmain@foo.go:42\thello",
		},
		{
			desc: "use the supplied EncodeTime, for both the entry and any times added",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     func(t time.Time, enc ArrayEncoder) { enc.AppendString(t.String()) },
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			extra: func(enc Encoder) {
				enc.AddTime("extra", _epoch)
				enc.AddArray("extras", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
					enc.AppendTime(_epoch)
					return nil
				}))
			},
			expectedJSON: `{"L":"info","T":"1970-01-01 00:00:00 +0000 UTC","N":"main","C":"foo.go:42","M":"hello","extra":"1970-01-01 00:00:00 +0000 UTC","extras":["1970-01-01 00:00:00 +0000 UTC"],"S":"fake-stack"}`,
			expectedConsole: "1970-01-01 00:00:00 +0000 UTC\tinfo\tmain@foo.go:42\thello\t" + // plain-text preamble
				`{"extra": "1970-01-01 00:00:00 +0000 UTC", "extras": ["1970-01-01 00:00:00 +0000 UTC"]}` + // JSON context
				"\nfake-stack", // stacktrace after newline
		},
		{
			desc: "use the supplied EncodeDuration for any durations added",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: func(d time.Duration, enc ArrayEncoder) { enc.AppendString(d.String()) },
				EncodeLevel:    base.EncodeLevel,
			},
			extra: func(enc Encoder) {
				enc.AddDuration("extra", time.Second)
				enc.AddArray("extras", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
					enc.AppendDuration(time.Minute)
					return nil
				}))
			},
			expectedJSON: `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello","extra":"1s","extras":["1m0s"],"S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tmain@foo.go:42\thello\t" + // preamble
				`{"extra": "1s", "extras": ["1m0s"]}` + // context
				"\nfake-stack", // stacktrace
		},
		{
			desc: "use the supplied EncodeLevel",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    func(l Level, enc ArrayEncoder) { enc.AppendString(strings.ToUpper(l.String())) },
			},
			expectedJSON:    `{"L":"INFO","T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
			expectedConsole: "0\tINFO\tmain@foo.go:42\thello\nfake-stack",
		},
		{
			desc: "close all open namespaces",
			cfg: EncoderConfig{
				LevelKey:       "L",
				TimeKey:        "T",
				MessageKey:     "M",
				NameKey:        "N",
				CallerKey:      "C",
				StacktraceKey:  "S",
				EncodeTime:     base.EncodeTime,
				EncodeDuration: base.EncodeDuration,
				EncodeLevel:    base.EncodeLevel,
			},
			extra: func(enc Encoder) {
				enc.OpenNamespace("outer")
				enc.OpenNamespace("inner")
				enc.AddString("foo", "bar")
				enc.OpenNamespace("innermost")
			},
			expectedJSON: `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello","outer":{"inner":{"foo":"bar","innermost":{}}},"S":"fake-stack"}`,
			expectedConsole: "0\tinfo\tmain@foo.go:42\thello\t" +
				`{"outer": {"inner": {"foo": "bar", "innermost": {}}}}` +
				"\nfake-stack",
		},
	}

	for i, tt := range tests {
		json := NewJSONEncoder(tt.cfg)
		console := NewConsoleEncoder(tt.cfg)
		if tt.extra != nil {
			tt.extra(json)
			tt.extra(console)
		}
		jsonOut, jsonErr := json.EncodeEntry(_testEntry, nil)
		if assert.NoError(t, jsonErr, "Unexpected error JSON-encoding entry in case #%d.", i) {
			assert.Equal(
				t,
				tt.expectedJSON+"\n",
				string(jsonOut),
				"Unexpected JSON output: expected to %v.", tt.desc,
			)
		}
		consoleOut, consoleErr := console.EncodeEntry(_testEntry, nil)
		if assert.NoError(t, consoleErr, "Unexpected error console-encoding entry in case #%d.", i) {
			assert.Equal(
				t,
				tt.expectedConsole+"\n",
				string(consoleOut),
				"Unexpected console output: expected to %v.", tt.desc,
			)
		}
	}
}
