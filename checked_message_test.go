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
	"github.com/stretchr/testify/require"
)

func TestJSONLoggerCheck(t *testing.T) {
	withJSONLogger(t, opts(InfoLevel), func(logger Logger, buf *testBuffer) {
		assert.False(
			t,
			logger.Check(DebugLevel, "Debug.").OK(),
			"Expected CheckedMessage to be not OK at disabled levels.",
		)

		cm := logger.Check(InfoLevel, "Info.")
		require.True(t, cm.OK(), "Expected CheckedMessage to be OK at enabled levels.")
		cm.Write(Int("magic", 42))
		assert.Equal(
			t,
			`{"level":"info","msg":"Info.","magic":42}`,
			buf.Stripped(),
			"Unexpected output after writing a CheckedMessage.",
		)
	})
}

func TestCheckedMessageIsSingleUse(t *testing.T) {
	expected := []string{
		`{"level":"info","msg":"single-use"}`,
		`{"level":"error","msg":"Shouldn't re-use a CheckedMessage.","original":"single-use"}`,
	}
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		cm := logger.Check(InfoLevel, "single-use")
		cm.Write() // ok
		cm.Write() // first re-use logs error
		cm.Write() // second re-use is silently ignored
		assert.Equal(t, expected, buf.Lines(), "Expected re-using a CheckedMessage to log an error.")
	})
}

func TestCheckedMessage_Chain(t *testing.T) {
	expected := []string{
		// not-ok base cases
		// NOTE no output

		// singleton ok cases
		`{"level":"info","msg":"apple","name":"A","i":3}`,
		`{"level":"info","msg":"banana","name":"A","i":4}`,

		// compound ok cases
		`{"level":"info","msg":"cherry","name":"A","i":5}`,
		`{"level":"info","msg":"cherry","name":"B","i":5}`,
	}
	withJSONLogger(t, opts(InfoLevel), func(logger Logger, buf *testBuffer) {
		loga := logger.With(String("name", "A"))
		logb := logger.With(String("name", "B"))

		tests := []struct {
			init       func() *CheckedMessage
			logs       []Logger
			lvl        Level
			msg        string
			expectedOK bool
			desc       string
		}{
			// not-ok base cases
			{
				init:       nil,
				expectedOK: false,
				desc:       "nil init",
			},
			{
				init:       nil,
				logs:       []Logger{loga},
				lvl:        DebugLevel,
				msg:        "nope",
				expectedOK: false,
				desc:       "nil init, one decline",
			},
			{
				init:       nil,
				logs:       []Logger{loga, logb},
				lvl:        DebugLevel,
				msg:        "nope",
				expectedOK: false,
				desc:       "nil init, two decline",
			},

			// singleton ok cases
			{
				init:       nil,
				logs:       []Logger{loga},
				lvl:        InfoLevel,
				msg:        "apple",
				expectedOK: true,
				desc:       "nil init, A OK",
			},
			{
				init:       func() *CheckedMessage { return loga.Check(InfoLevel, "banana") },
				logs:       []Logger{logb},
				lvl:        DebugLevel, // XXX hack, but we don't have split-level log{a,b,c} setup
				msg:        "banana",
				expectedOK: true,
				desc:       "A init, B decline",
			},

			// compound ok cases
			{
				init:       func() *CheckedMessage { return loga.Check(InfoLevel, "cherry") },
				logs:       []Logger{logb},
				lvl:        InfoLevel,
				msg:        "cherry",
				expectedOK: true,
				desc:       "A init, B OK",
			},
			// XXX: if we had split-level log{a,b,c} setup, we could easily test:
			// - A init, B decline, C OK
			// - A init, B OK, C decline
		}

		for i, tt := range tests {
			var cm *CheckedMessage
			if tt.init != nil {
				cm = tt.init()
			}
			for _, log := range tt.logs {
				cm = cm.Chain(log, tt.lvl, tt.msg)
			}
			assert.Equal(t, cm.OK(), tt.expectedOK, "%s: expected cm.OK()", tt.desc)
			cm.Write(Int("i", i))
		}
		assert.Equal(t, expected, buf.Lines(), "expected output from MultiCheckedMessage")
	})
}
