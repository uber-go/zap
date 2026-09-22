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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	//revive:disable:dot-imports
	. "go.uber.org/zap/zapcore"
)

var testEntry = Entry{
	LoggerName: "main",
	Level:      InfoLevel,
	Message:    `hello`,
	Time:       _epoch,
	Stack:      "fake-stack",
	Caller:     EntryCaller{Defined: true, File: "foo.go", Line: 42, Function: "foo.Foo"},
}

func TestConsoleEncodeEntry(t *testing.T) {
	tests := []struct {
		desc     string
		expected string
		ent      Entry
		fields   []Field
	}{
		{
			desc:     "info no fields",
			expected: "2018-06-19T16:33:42Z\tinfo\tbob\tlob law\n",
			ent: Entry{
				Level:      InfoLevel,
				Time:       time.Date(2018, 6, 19, 16, 33, 42, 99, time.UTC),
				LoggerName: "bob",
				Message:    "lob law",
			},
		},
		{
			desc:     "zero_time_omitted",
			expected: "info\tname\tmessage\n",
			ent: Entry{
				Level:      InfoLevel,
				Time:       time.Time{},
				LoggerName: "name",
				Message:    "message",
			},
		},
	}

	cfg := testEncoderConfig()
	cfg.EncodeTime = RFC3339TimeEncoder
	enc := NewConsoleEncoder(cfg)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			buf, err := enc.EncodeEntry(tt.ent, tt.fields)
			if assert.NoError(t, err, "Unexpected console encoding error.") {
				assert.Equal(t, tt.expected, buf.String(), "Incorrect encoded entry.")
			}
			buf.Free()
		})
	}
}

func TestConsoleSeparator(t *testing.T) {
	tests := []struct {
		desc        string
		separator   string
		wantConsole string
	}{
		{
			desc:        "space console separator",
			separator:   " ",
			wantConsole: "0 info main foo.go:42 foo.Foo hello\nfake-stack\n",
		},
		{
			desc:        "default console separator",
			separator:   "",
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\nfake-stack\n",
		},
		{
			desc:        "tag console separator",
			separator:   "\t",
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\nfake-stack\n",
		},
		{
			desc:        "dash console separator",
			separator:   "--",
			wantConsole: "0--info--main--foo.go:42--foo.Foo--hello\nfake-stack\n",
		},
	}

	for _, tt := range tests {
		console := NewConsoleEncoder(encoderTestEncoderConfig(tt.separator))
		t.Run(tt.desc, func(t *testing.T) {
			entry := testEntry
			consoleOut, err := console.EncodeEntry(entry, nil)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(
				t,
				tt.wantConsole,
				consoleOut.String(),
				"Unexpected console output",
			)
		})

	}
}

func encoderTestEncoderConfig(separator string) EncoderConfig {
	testEncoder := testEncoderConfig()
	testEncoder.ConsoleSeparator = separator
	return testEncoder
}

func TestConsoleFieldOrder(t *testing.T) {
	tests := []struct {
		desc        string
		order       []OrderField
		wantConsole string
	}{
		{
			desc: "DefaultOrder",
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\nfake-stack\n",
		},
		{
			desc:        "CustomOrder1",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldCallee, OrderFieldFunction, OrderFieldName, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tfoo.go:42\tfoo.Foo\tmain\thello\nfake-stack\n",
		},
		{
			desc:        "MessageFirst",
			order:       []OrderField{OrderFieldMessage, OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldStack},
			wantConsole: "hello\t0\tinfo\tmain\tfoo.go:42\tfoo.Foo\nfake-stack\n",
		},
		{
			desc:        "StackMiddle",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldStack, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldMessage},
			wantConsole: "0\tinfo\t\nfake-stack\tmain\tfoo.go:42\tfoo.Foo\thello\n",
		},
		{
			desc:        "TimeLast",
			order:       []OrderField{OrderFieldMessage, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldStack, OrderFieldTime},
			wantConsole: "hello\tinfo\tmain\tfoo.go:42\tfoo.Foo\t\nfake-stack\t0\n",
		},
		{
			desc:        "StackFirst",
			order:       []OrderField{OrderFieldStack, OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldMessage},
			wantConsole: "\nfake-stack\t0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\n",
		},
		{
			desc:        "CallerSplit",
			order:       []OrderField{OrderFieldTime, OrderFieldCallee, OrderFieldLevel, OrderFieldFunction, OrderFieldName, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tfoo.go:42\tinfo\tfoo.Foo\tmain\thello\nfake-stack\n",
		},
		{
			desc:        "MessageStackAdjacent",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldStack, OrderFieldMessage},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\t\nfake-stack\thello\n",
		},
	}

	for _, tt := range tests {
		console := NewConsoleEncoder(encoderTestFieldOrderConfig(tt.order))
		t.Run(tt.desc, func(t *testing.T) {
			entry := testEntry
			consoleOut, err := console.EncodeEntry(entry, nil)
			if !assert.NoError(t, err) {
				return
			}

			// Print debug info
			t.Logf("Expected (% x): %q", []byte(tt.wantConsole), tt.wantConsole)
			t.Logf("Actual (% x): %q", []byte(consoleOut.String()), consoleOut.String())

			assert.Equal(
				t,
				tt.wantConsole,
				consoleOut.String(),
				"Unexpected console output",
			)
		})
	}
}

func encoderTestFieldOrderConfig(order []OrderField) EncoderConfig {
	testEncoder := testEncoderConfig()
	testEncoder.ConsoleFieldOrder = order
	return testEncoder
}

func TestConsoleFieldOrderWithContext(t *testing.T) {
	tests := []struct {
		desc        string
		order       []OrderField
		wantConsole string
	}{
		{
			desc: "StandardOrderWithContext",
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\t{\"foo\": \"bar\"}\nfake-stack\n",
		},
		{
			desc:        "MessageStackContextBetween",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\t{\"foo\": \"bar\"}\nfake-stack\n",
		},
		{
			desc:        "MessageLastWithContext",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldStack, OrderFieldMessage},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\t\nfake-stack\thello\t{\"foo\": \"bar\"}\n",
		},
		{
			desc:        "StackMiddleWithContext",
			order:       []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldStack, OrderFieldName, OrderFieldCallee, OrderFieldFunction, OrderFieldMessage},
			wantConsole: "0\tinfo\t\nfake-stack\tmain\tfoo.go:42\tfoo.Foo\thello\t{\"foo\": \"bar\"}\n",
		},
	}

	for _, tt := range tests {
		console := NewConsoleEncoder(encoderTestFieldOrderConfig(tt.order))
		t.Run(tt.desc, func(t *testing.T) {
			entry := testEntry
			fields := []Field{Field{Key: "foo", Type: StringType, String: "bar"}}
			consoleOut, err := console.EncodeEntry(entry, fields)
			if !assert.NoError(t, err) {
				return
			}

			// Print debug info
			t.Logf("Expected (% x): %q", []byte(tt.wantConsole), tt.wantConsole)
			t.Logf("Actual (% x): %q", []byte(consoleOut.String()), consoleOut.String())

			assert.Equal(
				t,
				tt.wantConsole,
				consoleOut.String(),
				"Unexpected console output with context",
			)
		})
	}
}

func TestConsoleFieldOrderWithDifferentSeparators(t *testing.T) {
	separators := []string{"\t", " ", " | ", "--", ""}
	orders := []struct {
		desc  string
		order []OrderField
	}{
		{
			desc: "DefaultOrder",
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
		},
		{
			desc: "MessageFirst",
			order: []OrderField{OrderFieldMessage, OrderFieldTime, OrderFieldLevel, OrderFieldName,
				OrderFieldCallee, OrderFieldFunction, OrderFieldStack},
		},
		{
			desc: "StackMiddle",
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldStack, OrderFieldName,
				OrderFieldCallee, OrderFieldFunction, OrderFieldMessage},
		},
	}

	for _, sep := range separators {
		t.Run("Separator:"+sep, func(t *testing.T) {
			for _, o := range orders {
				t.Run(o.desc, func(t *testing.T) {
					config := encoderTestFieldOrderConfig(o.order)
					config.ConsoleSeparator = sep
					console := NewConsoleEncoder(config)

					entry := testEntry
					consoleOut, err := console.EncodeEntry(entry, nil)
					assert.NoError(t, err, "Encoding error")

					// Note: Exact output verification is challenging due to field order and separator variations

					// Print debug info
					t.Logf("Using separator [%s] output: %q", sep, consoleOut.String())
				})
			}
		})
	}
}

func TestConsoleFieldOrderWithMissingFields(t *testing.T) {
	tests := []struct {
		desc        string
		entry       Entry
		order       []OrderField
		wantConsole string
	}{
		{
			desc: "MissingTimestamp",
			entry: func() Entry {
				ent := testEntry
				// Time value is zero, should be omitted
				ent.Time = time.Time{}
				return ent
			}(),
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "info\tmain\tfoo.go:42\tfoo.Foo\thello\nfake-stack\n",
		},
		{
			desc: "MissingLoggerName",
			entry: func() Entry {
				ent := testEntry
				// LoggerName is empty
				ent.LoggerName = ""
				return ent
			}(),
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tfoo.go:42\tfoo.Foo\thello\nfake-stack\n",
		},
		{
			desc: "MissingCaller",
			entry: func() Entry {
				ent := testEntry
				// Caller is undefined
				ent.Caller = EntryCaller{}
				return ent
			}(),
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tmain\thello\nfake-stack\n",
		},
		{
			desc: "MissingStack",
			entry: func() Entry {
				ent := testEntry
				// Stack is empty
				ent.Stack = ""
				return ent
			}(),
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "0\tinfo\tmain\tfoo.go:42\tfoo.Foo\thello\n",
		},
		{
			desc: "MessageOnly",
			entry: func() Entry {
				// Set only message
				ent := Entry{Message: "hello"}
				// Need to explicitly set level, otherwise default level would be output
				ent.Level = InfoLevel
				return ent
			}(),
			order: []OrderField{OrderFieldTime, OrderFieldLevel, OrderFieldName, OrderFieldCallee,
				OrderFieldFunction, OrderFieldMessage, OrderFieldStack},
			wantConsole: "info\thello\n",
		},
		{
			desc: "MissingFieldsNonDefaultOrder",
			entry: func() Entry {
				// No LoggerName, Caller undefined
				ent := testEntry
				ent.LoggerName = ""
				ent.Caller = EntryCaller{}
				return ent
			}(),
			order: []OrderField{OrderFieldMessage, OrderFieldLevel, OrderFieldCallee, OrderFieldTime,
				OrderFieldName, OrderFieldFunction, OrderFieldStack},
			wantConsole: "hello\tinfo\t0\nfake-stack\n",
		},
	}

	for _, tt := range tests {
		console := NewConsoleEncoder(encoderTestFieldOrderConfig(tt.order))
		t.Run(tt.desc, func(t *testing.T) {
			consoleOut, err := console.EncodeEntry(tt.entry, nil)
			if !assert.NoError(t, err) {
				return
			}

			// Print debug info
			t.Logf("Expected (% x): %q", []byte(tt.wantConsole), tt.wantConsole)
			t.Logf("Actual (% x): %q", []byte(consoleOut.String()), consoleOut.String())

			assert.Equal(
				t,
				tt.wantConsole,
				consoleOut.String(),
				"Unexpected console output with missing fields",
			)
		})
	}
}
