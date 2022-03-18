// Copyright (c) 2022 Uber Technologies, Inc.
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

//go:build go1.18
// +build go1.18

package zap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give Field
		want []any
	}{
		{
			desc: "nil slice",
			give: Objects[*emptyObject]("", nil),
			want: []any{},
		},
		{
			desc: "empty slice",
			give: Objects("", []*emptyObject{}),
			want: []any{},
		},
		{
			desc: "single item",
			give: Objects("", []*emptyObject{
				{},
			}),
			want: []any{
				map[string]any{},
			},
		},
		{
			desc: "multiple different objects",
			give: Objects("", []zapcore.ObjectMarshalerFunc{
				func(enc zapcore.ObjectEncoder) error {
					enc.AddString("foo", "bar")
					return nil
				},
				func(enc zapcore.ObjectEncoder) error {
					enc.AddInt("baz", 42)
					return nil
				},
				func(enc zapcore.ObjectEncoder) error {
					enc.AddBool("qux", true)
					return nil
				},
			}),
			want: []any{
				map[string]any{"foo": "bar"},
				map[string]any{"baz": 42},
				map[string]any{"qux": true},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			tt.give.Key = "k"

			enc := zapcore.NewMapObjectEncoder()
			tt.give.AddTo(enc)
			assert.Equal(t, tt.want, enc.Fields["k"])
		})
	}
}

type emptyObject struct{}

func (*emptyObject) MarshalLogObject(zapcore.ObjectEncoder) error {
	return nil
}

func TestObjects_marshalError(t *testing.T) {
	t.Parallel()

	enc := zapcore.NewMapObjectEncoder()
	Objects("k", []zapcore.ObjectMarshalerFunc{
		func(enc zapcore.ObjectEncoder) error {
			enc.AddString("foo", "bar")
			return nil
		},
		func(enc zapcore.ObjectEncoder) error {
			enc.AddString("baz", "qux")
			return errors.New("great sadness")
		},
		func(enc zapcore.ObjectEncoder) error {
			t.Fatal("this item should not be encoded")
			return nil
		},
	}).AddTo(enc)

	require.Contains(t, enc.Fields, "k")
	assert.Equal(t,
		[]any{
			map[string]any{"foo": "bar"},
			map[string]any{"baz": "qux"},
		},
		enc.Fields["k"])

	// AddTo puts the error in a "%vError" field based on the name of the
	// original field.
	require.Contains(t, enc.Fields, "kError")
	assert.Equal(t, "great sadness", enc.Fields["kError"],
		"error should get encoded")
}
