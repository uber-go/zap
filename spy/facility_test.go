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

package spy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
	"go.uber.org/zap/spy"
)

func TestSpy_With(t *testing.T) {
	sf1, sink := spy.New(zap.InfoLevel)

	// need to pad out enough initial fields so that the underlying slice cap()
	// gets ahead of its len() so that the sf3/4 With append's could choose
	// not to copy (if the implementation doesn't force them)
	sf1 = sf1.With(zap.Int("a", 1)).With(zap.Int("b", 2))

	sf2 := sf1.With(zap.Int("c", 3))
	sf3 := sf2.With(zap.Int("d", 4))
	sf4 := sf2.With(zap.Int("e", 5))

	for i, f := range []zap.Facility{sf2, sf3, sf4} {
		f.Log(zap.Entry{
			Level:   zap.InfoLevel,
			Message: "hello",
		}, []zap.Field{zap.Int("i", i)})
	}

	assert.Equal(t, []spy.Log{
		{
			Level: zap.InfoLevel,
			Msg:   "hello",
			Fields: []zap.Field{
				zap.Int("a", 1),
				zap.Int("b", 2),
				zap.Int("c", 3),
				zap.Int("i", 0),
			},
		},
		{
			Level: zap.InfoLevel,
			Msg:   "hello",
			Fields: []zap.Field{
				zap.Int("a", 1),
				zap.Int("b", 2),
				zap.Int("c", 3),
				zap.Int("d", 4),
				zap.Int("i", 1),
			},
		},
		{
			Level: zap.InfoLevel,
			Msg:   "hello",
			Fields: []zap.Field{
				zap.Int("a", 1),
				zap.Int("b", 2),
				zap.Int("c", 3),
				zap.Int("e", 5),
				zap.Int("i", 2),
			},
		},
	}, sink.Logs(), "expected no field sharing between With siblings")
}
