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

func TestMultiCheckedMessage(t *testing.T) {
	expected := []string{
		// not-ok base cases
		// NOTE no output

		// singleton ok cases
		`{"level":"info","msg":"apple","name":"A","i":4}`,
		`{"level":"info","msg":"banana","name":"A","i":5}`,
		`{"level":"info","msg":"cherry","name":"B","i":6}`,

		// compound ok cases
		`{"level":"info","msg":"durian","name":"A","i":7}`,
		`{"level":"info","msg":"durian","name":"B","i":7}`,
		`{"level":"info","msg":"elderberry","name":"A","i":8}`,
		`{"level":"info","msg":"elderberry","name":"B","i":8}`,
		`{"level":"info","msg":"fig","name":"A","i":9}`,
		`{"level":"info","msg":"fig","name":"B","i":9}`,
		`{"level":"info","msg":"fig","name":"C","i":9}`,

		// proliferation
		`{"level":"warn","msg":"guava","name":"B","i":10}`,
		`{"level":"warn","msg":"guava","name":"A","i":10}`,
		`{"level":"warn","msg":"guava","name":"C","i":10}`,
		// NOTE huckleberry debugging suppressed
		`{"level":"error","msg":"jack","name":"B","i":10}`,
		`{"level":"error","msg":"jack","name":"C","i":10}`,
		`{"level":"error","msg":"jack","name":"A","i":10}`,
	}
	withJSONLogger(t, opts(InfoLevel), func(logger Logger, buf *testBuffer) {
		loga := logger.With(String("name", "A"))
		logb := logger.With(String("name", "B"))
		logc := logger.With(String("name", "C"))

		tests := []struct {
			cms        []*CheckedMessage
			expectedOK bool
			desc       string
		}{
			// not-ok base cases
			{nil, false, "nil slice"},
			{[]*CheckedMessage{}, false, "empty slice"},
			{[]*CheckedMessage{nil}, false, "singleton-nil"},
			{[]*CheckedMessage{nil, nil}, false, "doubleton-nil"},

			// singleton ok cases
			{[]*CheckedMessage{loga.Check(InfoLevel, "apple")}, true, "just A"},
			{[]*CheckedMessage{loga.Check(InfoLevel, "banana"), nil}, true, "A and nil"},
			{[]*CheckedMessage{nil, logb.Check(InfoLevel, "cherry")}, true, "nil and B"},

			// compound ok cases
			{[]*CheckedMessage{
				loga.Check(InfoLevel, "durian"),
				logb.Check(InfoLevel, "durian"),
			}, true, "A and B"},
			{[]*CheckedMessage{
				loga.Check(InfoLevel, "elderberry"),
				logb.Check(InfoLevel, "elderberry"),
			}, true, "A, nil, and B"},
			{[]*CheckedMessage{
				loga.Check(InfoLevel, "fig"),
				logb.Check(InfoLevel, "fig"),
				logc.Check(InfoLevel, "fig"),
			}, true, "A, B, and C"},

			// proliferation
			{[]*CheckedMessage{
				logb.Check(WarnLevel, "guava"),
				loga.Check(WarnLevel, "guava"),
				logc.Check(WarnLevel, "guava"),
				logc.Check(DebugLevel, "huckleberry"),
				loga.Check(DebugLevel, "huckleberry"),
				logb.Check(DebugLevel, "huckleberry"),
				logb.Check(ErrorLevel, "jack"),
				logc.Check(ErrorLevel, "jack"),
				loga.Check(ErrorLevel, "jack"),
			}, true, "A, B, and C"},
		}

		for i, tt := range tests {
			cm := MultiCheckedMessage(tt.cms...)

			var first *CheckedMessage
			n := 0
			for _, cmi := range tt.cms {
				if cmi.OK() {
					first = cmi
					n++
				}
			}
			if n == 1 {
				assert.Equal(t, first, cm, "%s: expected singleton degeneration", tt.desc)
			} else if n > 1 && cm != nil {
				for i, cmi := range tt.cms {
					assert.NotEqual(t, cmi, cm, "%s: unexpected message re-use @[%v]", tt.desc, i)
				}
			}

			assert.Equal(t, cm.OK(), tt.expectedOK, "%s: expected MultiCheckedMessage(...).OK()", tt.desc)
			cm.Write(Int("i", i))
		}
		assert.Equal(t, expected, buf.Lines(), "expected output from MultiCheckedMessage")
	})
}
