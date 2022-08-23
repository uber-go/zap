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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestObjectsAndObjectValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give Field
		want []any
	}{
		{
			desc: "Objects/nil slice",
			give: Objects[*emptyObject]("", nil),
			want: []any{},
		},
		{
			desc: "ObjectValues/nil slice",
			give: ObjectValues[emptyObject]("", nil),
			want: []any{},
		},
		{
			desc: "ObjectValues/empty slice",
			give: ObjectValues("", []emptyObject{}),
			want: []any{},
		},
		{
			desc: "ObjectValues/single item",
			give: ObjectValues("", []emptyObject{
				{},
			}),
			want: []any{
				map[string]any{},
			},
		},
		{
			desc: "Objects/multiple different objects",
			give: Objects("", []*fakeObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
				map[string]any{"value": "baz"},
			},
		},
		{
			desc: "ObjectValues/multiple different objects",
			give: ObjectValues("", []fakeObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
				map[string]any{"value": "baz"},
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

type fakeObject struct {
	value string
	err   error // marshaling error, if any
}

func (o *fakeObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("value", o.value)
	return o.err
}

func TestObjectsAndObjectValues_marshalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		give    Field
		want    []any
		wantErr string
	}{
		{
			desc: "Objects",
			give: Objects("", []*fakeObject{
				{value: "foo"},
				{value: "bar", err: errors.New("great sadness")},
				{value: "baz"}, // does not get marshaled
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
			},
			wantErr: "great sadness",
		},
		{
			desc: "ObjectValues",
			give: ObjectValues("", []fakeObject{
				{value: "foo"},
				{value: "bar", err: errors.New("stuff failed")},
				{value: "baz"}, // does not get marshaled
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
			},
			wantErr: "stuff failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			tt.give.Key = "k"

			enc := zapcore.NewMapObjectEncoder()
			tt.give.AddTo(enc)

			require.Contains(t, enc.Fields, "k")
			assert.Equal(t, tt.want, enc.Fields["k"])

			// AddTo puts the error in a "%vError" field based on the name of the
			// original field.
			require.Contains(t, enc.Fields, "kError")
			assert.Equal(t, tt.wantErr, enc.Fields["kError"])
		})
	}
}

type stringerObject struct {
	value string
}

func (s stringerObject) String() string {
	return s.value
}

func TestStringers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give Field
		want []any
	}{
		{
			desc: "Stringers",
			give: Stringers("", []stringerObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				"foo",
				"bar",
				"baz",
			},
		},
		{
			desc: "Stringers with []fmt.Stringer",
			give: Stringers("", []fmt.Stringer{
				stringerObject{value: "foo"},
				stringerObject{value: "bar"},
				stringerObject{value: "baz"},
			}),
			want: []any{
				"foo",
				"bar",
				"baz",
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
