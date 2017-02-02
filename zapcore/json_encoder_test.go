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
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.uber.org/zap/zapcore"
)

var epoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

func withJSONEncoder(f func(Encoder)) {
	f(NewJSONEncoder(testEncoderConfig()))
}

func TestJSONEncoderConfiguration(t *testing.T) {
	epoch := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	entry := Entry{
		LoggerName: "main",
		Level:      InfoLevel,
		Message:    `hello`,
		Time:       epoch,
		Stack:      "fake-stack",
		Caller:     EntryCaller{Defined: true, File: "foo.go", Line: 42},
	}
	base := testEncoderConfig()
	// expected: `{"level":"info","ts":100,"name":"main","caller":"foo.go:42","msg":"hello","stacktrace":"fake-stack"}`,

	tests := []struct {
		desc     string
		cfg      EncoderConfig
		extra    func(Encoder)
		expected string
	}{
		{
			desc: "use custom entry keys",
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
			expected: `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
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
			expected: `{"T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
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
			expected: `{"L":"info","N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
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
			expected: `{"L":"info","T":0,"N":"main","C":"foo.go:42","S":"fake-stack"}`,
		},
		{
			desc: "skip name is NameKey is omitted",
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
			expected: `{"L":"info","T":0,"C":"foo.go:42","M":"hello","S":"fake-stack"}`,
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
			expected: `{"L":"info","T":0,"N":"main","M":"hello","S":"fake-stack"}`,
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
			expected: `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello"}`,
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
				enc.AddTime("extra", epoch)
				enc.AddArray("extras", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
					enc.AppendTime(epoch)
					return nil
				}))
			},
			expected: `{"L":"info","T":"1970-01-01 00:00:00 +0000 UTC","N":"main","C":"foo.go:42","M":"hello","extra":"1970-01-01 00:00:00 +0000 UTC","extras":["1970-01-01 00:00:00 +0000 UTC"],"S":"fake-stack"}`,
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
			expected: `{"L":"info","T":0,"N":"main","C":"foo.go:42","M":"hello","extra":"1s","extras":["1m0s"],"S":"fake-stack"}`,
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
			expected: `{"L":"INFO","T":0,"N":"main","C":"foo.go:42","M":"hello","S":"fake-stack"}`,
		},
	}

	for i, tt := range tests {
		enc := NewJSONEncoder(tt.cfg)
		if tt.extra != nil {
			tt.extra(enc)
		}
		buf, err := enc.EncodeEntry(entry, nil)
		if assert.NoError(t, err, "Unexpected error encoding entry in case #%d.", i) {
			assert.Equal(t, tt.expected+"\n", string(buf), "Unexpected output: expected to %v.", tt.desc)
		}
	}
}

func TestJSONEncodeEntry(t *testing.T) {
	epoch := time.Unix(0, 0)
	withJSONEncoder(func(enc Encoder) {
		// Messages should be escaped.
		enc.AddString("foo", "bar")
		buf, err := enc.EncodeEntry(Entry{
			LoggerName: "",
			Level:      InfoLevel,
			Message:    `hello\`,
			Time:       epoch,
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":0,"msg":"hello\\","foo":"bar"}`+"\n",
			string(buf),
		)

		// We should be able to re-use the encoder, preserving the accumulated
		// fields.
		buf, err = enc.EncodeEntry(Entry{
			LoggerName: "main",
			Level:      InfoLevel,
			Message:    `hello`,
			Time:       time.Unix(0, 100*int64(time.Millisecond)),
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"name":"main","msg":"hello","foo":"bar"}`+"\n",
			string(buf),
		)

		// Stacktraces are included.
		buf, err = enc.EncodeEntry(Entry{
			LoggerName: `main\`,
			Level:      InfoLevel,
			Message:    `hello`,
			Time:       time.Unix(0, 100*int64(time.Millisecond)),
			Stack:      "trace",
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"name":"main\\","msg":"hello","foo":"bar","stacktrace":"trace"}`+"\n",
			string(buf),
		)

		// Caller is included.
		buf, err = enc.EncodeEntry(Entry{
			LoggerName: "main.lib.foo",
			Level:      InfoLevel,
			Message:    `hello`,
			Time:       time.Unix(0, 100*int64(time.Millisecond)),
			Caller:     MakeEntryCaller(runtime.Caller(0)),
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Regexp(
			t,
			`{"level":"info","ts":100,"name":"main.lib.foo","caller":"/.*zap/zapcore/json_encoder_test.go:\d+","msg":"hello","foo":"bar"}`+"\n",
			string(buf),
		)
	})
}

func TestJSONEncodeEntryClosesNamespaces(t *testing.T) {
	withJSONEncoder(func(enc Encoder) {
		enc.OpenNamespace("outer")
		enc.OpenNamespace("inner")
		buf, err := enc.EncodeEntry(Entry{Message: `hello`, Time: time.Unix(0, 0)}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":0,"msg":"hello","outer":{"inner":{}}}`+"\n",
			string(buf),
		)
	})
}
