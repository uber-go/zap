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

func TestCheckedMessageUnsafeWrite(t *testing.T) {
	withJSONLogger(t, opts(InfoLevel), func(logger Logger, buf *testBuffer) {
		cm := logger.Check(InfoLevel, "bob lob law blog")
		cm.Write()
		cm.Write()
		assert.Equal(t, []string{
			`{"level":"info","msg":"bob lob law blog"}`,
			`{"level":"dpanic","msg":"Must not call zap.(*CheckedMessage).Write() more than once","prior":{"level":"info","msg":"bob lob law blog"}}`,
		}, buf.Lines(), "Expected one lob log, and a fatal")
	})
}

func TestCheckedMessage_Chain(t *testing.T) {
	withJSONLogger(t, opts(InfoLevel), func(logger Logger, buf *testBuffer) {
		loga := logger.With(String("name", "A"))
		logb := logger.With(String("name", "B"))
		logc := logger.With(String("name", "C"))

		for i, tt := range []struct {
			cms        []*CheckedMessage
			expectedOK bool
			desc       string
			expected   []string
		}{
			// not-ok base cases
			{
				cms:        []*CheckedMessage{nil},
				expectedOK: false,
				desc:       "nil",
				expected:   []string{},
			},
			{
				cms: []*CheckedMessage{
					nil,
					loga.Check(DebugLevel, "nope"),
				},
				expectedOK: false,
				desc:       "nil, A decline",
				expected:   []string{},
			},
			{
				cms: []*CheckedMessage{
					nil,
					loga.Check(DebugLevel, "nope"),
					logb.Check(DebugLevel, "nope"),
				},
				expectedOK: false,
				desc:       "nil, A and B decline",
				expected:   []string{},
			},

			// singleton ok cases
			{
				cms: []*CheckedMessage{
					nil,
					loga.Check(InfoLevel, "apple"),
				},
				expectedOK: true,
				desc:       "nil, A OK",
				expected: []string{
					`{"level":"info","msg":"apple","name":"A","i":3}`,
				},
			},
			{
				cms: []*CheckedMessage{
					loga.Check(InfoLevel, "banana"),
					logb.Check(DebugLevel, "banana"),
				},
				expectedOK: true,
				desc:       "A OK, B decline",
				expected: []string{
					`{"level":"info","msg":"banana","name":"A","i":4}`,
				},
			},

			// compound ok cases
			{
				cms: []*CheckedMessage{
					loga.Check(InfoLevel, "cherry"),
					logb.Check(InfoLevel, "cherry"),
				},
				expectedOK: true,
				desc:       "A and B OK",
				expected: []string{
					`{"level":"info","msg":"cherry","name":"A","i":5}`,
					`{"level":"info","msg":"cherry","name":"B","i":5}`,
				},
			},
			{
				cms: []*CheckedMessage{
					loga.Check(InfoLevel, "durian"),
					nil,
					logb.Check(InfoLevel, "durian"),
				},
				expectedOK: true,
				desc:       "A OK, nil, B OK",
				expected: []string{
					`{"level":"info","msg":"durian","name":"A","i":6}`,
					`{"level":"info","msg":"durian","name":"B","i":6}`,
				},
			},
			{
				cms: []*CheckedMessage{
					loga.Check(InfoLevel, "elderberry"),
					logb.Check(InfoLevel, "elderberry"),
					nil,
				},
				expectedOK: true,
				desc:       "A OK, B OK, nil",
				expected: []string{
					`{"level":"info","msg":"elderberry","name":"A","i":7}`,
					`{"level":"info","msg":"elderberry","name":"B","i":7}`,
				},
			},

			// trees
			{
				cms: []*CheckedMessage{
					loga.Check(InfoLevel, "fig"),
					logb.Check(InfoLevel, "fig").Chain(logc.Check(InfoLevel, "fig")),
				},
				expectedOK: true,
				desc:       "A OK, ( B OK, C OK )",
				expected: []string{
					`{"level":"info","msg":"fig","name":"A","i":8}`,
					`{"level":"info","msg":"fig","name":"B","i":8}`,
					`{"level":"info","msg":"fig","name":"C","i":8}`,
				},
			},
			{
				cms: []*CheckedMessage{
					logb.Check(InfoLevel, "guava").Chain(logc.Check(InfoLevel, "guava")),
					loga.Check(InfoLevel, "guava"),
				},
				expectedOK: true,
				desc:       "( B OK, C OK ), A OK",
				expected: []string{
					`{"level":"info","msg":"guava","name":"B","i":9}`,
					`{"level":"info","msg":"guava","name":"C","i":9}`,
					`{"level":"info","msg":"guava","name":"A","i":9}`,
				},
			},
		} {
			cm := tt.cms[0].Chain(tt.cms[1:]...)
			assert.Equal(t, cm.OK(), tt.expectedOK, "%s: expected cm.OK()", tt.desc)
			cm.Write(Int("i", i))
			assert.Equal(t, tt.expected, buf.Lines(), "expected output from MultiCheckedMessage")
			buf.Reset()
		}
	})
}
