// Copyright (c) 2021 Uber Technologies, Inc.
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

package zapio

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc   string
		level  zapcore.Level // defaults to info
		writes []string
		want   []zapcore.Entry
	}{
		{
			desc: "simple",
			writes: []string{
				"foo\n",
				"bar\n",
				"baz\n",
			},
			want: []zapcore.Entry{
				{Level: zap.InfoLevel, Message: "foo"},
				{Level: zap.InfoLevel, Message: "bar"},
				{Level: zap.InfoLevel, Message: "baz"},
			},
		},
		{
			desc:  "level too low",
			level: zap.DebugLevel,
			writes: []string{
				"foo\n",
				"bar\n",
			},
			want: []zapcore.Entry{},
		},
		{
			desc:  "multiple newlines in a message",
			level: zap.WarnLevel,
			writes: []string{
				"foo\nbar\n",
				"baz\n",
				"qux\nquux\n",
			},
			want: []zapcore.Entry{
				{Level: zap.WarnLevel, Message: "foo"},
				{Level: zap.WarnLevel, Message: "bar"},
				{Level: zap.WarnLevel, Message: "baz"},
				{Level: zap.WarnLevel, Message: "qux"},
				{Level: zap.WarnLevel, Message: "quux"},
			},
		},
		{
			desc:  "message split across multiple writes",
			level: zap.ErrorLevel,
			writes: []string{
				"foo",
				"bar\nbaz",
				"qux",
			},
			want: []zapcore.Entry{
				{Level: zap.ErrorLevel, Message: "foobar"},
				{Level: zap.ErrorLevel, Message: "bazqux"},
			},
		},
		{
			desc: "blank lines in the middle",
			writes: []string{
				"foo\n\nbar\nbaz",
			},
			want: []zapcore.Entry{
				{Level: zap.InfoLevel, Message: "foo"},
				{Level: zap.InfoLevel, Message: ""},
				{Level: zap.InfoLevel, Message: "bar"},
				{Level: zap.InfoLevel, Message: "baz"},
			},
		},
		{
			desc: "blank line at the end",
			writes: []string{
				"foo\nbar\nbaz\n",
			},
			want: []zapcore.Entry{
				{Level: zap.InfoLevel, Message: "foo"},
				{Level: zap.InfoLevel, Message: "bar"},
				{Level: zap.InfoLevel, Message: "baz"},
			},
		},
		{
			desc: "multiple blank line at the end",
			writes: []string{
				"foo\nbar\nbaz\n\n",
			},
			want: []zapcore.Entry{
				{Level: zap.InfoLevel, Message: "foo"},
				{Level: zap.InfoLevel, Message: "bar"},
				{Level: zap.InfoLevel, Message: "baz"},
				{Level: zap.InfoLevel, Message: ""},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // for t.Parallel
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			core, observed := observer.New(zap.InfoLevel)

			w := Writer{
				Log:   zap.New(core),
				Level: tt.level,
			}

			for _, s := range tt.writes {
				_, err := io.WriteString(&w, s)
				require.NoError(t, err, "Writer.Write failed.")
			}

			assert.NoError(t, w.Close(), "Writer.Close failed.")

			// Turn []observer.LoggedEntry => []zapcore.Entry
			got := make([]zapcore.Entry, observed.Len())
			for i, ent := range observed.AllUntimed() {
				got[i] = ent.Entry
			}
			assert.Equal(t, tt.want, got, "Logged entries do not match.")
		})
	}
}

func TestWrite_Sync(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)

	w := Writer{
		Log:   zap.New(core),
		Level: zap.InfoLevel,
	}

	io.WriteString(&w, "foo")
	io.WriteString(&w, "bar")

	t.Run("no sync", func(t *testing.T) {
		assert.Zero(t, observed.Len(), "Expected no logs yet")
	})

	t.Run("sync", func(t *testing.T) {
		defer observed.TakeAll()

		require.NoError(t, w.Sync(), "Sync must not fail")

		assert.Equal(t, []observer.LoggedEntry{
			{Entry: zapcore.Entry{Message: "foobar"}, Context: []zapcore.Field{}},
		}, observed.AllUntimed(), "Log messages did not match")
	})

	t.Run("sync on empty", func(t *testing.T) {
		require.NoError(t, w.Sync(), "Sync must not fail")
		assert.Zero(t, observed.Len(), "Expected no logs yet")
	})
}

func BenchmarkWriter(b *testing.B) {
	tests := []struct {
		name   string
		writes [][]byte
	}{
		{
			name: "single",
			writes: [][]byte{
				[]byte("foobar\n"),
				[]byte("bazqux\n"),
			},
		},
		{
			name: "splits",
			writes: [][]byte{
				[]byte("foo"),
				[]byte("bar\nbaz"),
				[]byte("qux\n"),
			},
		},
	}

	writer := Writer{
		Log:   zap.New(new(partiallyNopCore)),
		Level: zapcore.DebugLevel,
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for _, bs := range tt.writes {
					writer.Write(bs)
				}
			}
		})
	}
}

// partiallyNopCore behaves exactly like NopCore except it always returns true
// for whether the provided level is enabled, and accepts all Check requests.
//
// This lets us measure the overhead of the writer without measuring the cost
// of logging.
type partiallyNopCore struct{}

func (*partiallyNopCore) Enabled(zapcore.Level) bool { return true }

func (c *partiallyNopCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}

func (c *partiallyNopCore) With([]zapcore.Field) zapcore.Core        { return c }
func (*partiallyNopCore) Write(zapcore.Entry, []zapcore.Field) error { return nil }
func (*partiallyNopCore) Sync() error                                { return nil }
