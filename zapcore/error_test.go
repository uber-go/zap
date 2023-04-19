// Copyright (c) 2017 Uber Technologies, Inc.
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
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/multierr"
	. "go.uber.org/zap/zapcore"
)

type errTooManyUsers int

func (e errTooManyUsers) Error() string {
	return fmt.Sprintf("%d too many users", int(e))
}

func (e errTooManyUsers) Format(s fmt.State, verb rune) {
	// Implement fmt.Formatter, but don't add any information beyond the basic
	// Error method.
	if verb == 'v' && s.Flag('+') {
		io.WriteString(s, e.Error())
	}
}

func (e errTooManyUsers) MarshalLogObject(oe ObjectEncoder) error {
	oe.AddInt("numUsers", int(e))
	return nil
}

type customMultierr struct{}

func (e customMultierr) Error() string {
	return "great sadness"
}

func (e customMultierr) Errors() []error {
	return []error{
		errors.New("foo"),
		nil,
		multierr.Append(
			errors.New("bar"),
			errors.New("baz"),
		),
	}
}

type customErrObject struct {
	name       string
	underlying error
}

func (e customErrObject) Unwrap() error { return e.underlying }
func (e customErrObject) Error() string { return fmt.Sprintf("error %s", e.name) }
func (e customErrObject) MarshalLogObject(enc ObjectEncoder) error {
	enc.AddString(e.name, "err")
	return fmt.Errorf("marshal %v failed", e.name)
}

func TestErrorEncoding(t *testing.T) {
	tests := []struct {
		msg   string
		k     string
		t     FieldType // defaults to ErrorType
		iface interface{}
		want  map[string]interface{}
	}{
		{
			msg:   "custom key and fields",
			k:     "k",
			iface: errTooManyUsers(2),
			want: map[string]interface{}{
				"k": "2 too many users",
				"kFields": map[string]interface{}{
					"numUsers": 2,
				},
			},
		},
		{
			msg: "multierr",
			k:   "err",
			iface: multierr.Combine(
				errors.New("foo"),
				errors.New("bar"),
				errors.New("baz"),
			),
			want: map[string]interface{}{
				"err": "foo; bar; baz",
				"errCauses": []interface{}{
					map[string]interface{}{"error": "foo"},
					map[string]interface{}{"error": "bar"},
					map[string]interface{}{"error": "baz"},
				},
			},
		},
		{
			msg:   "nested error causes",
			k:     "e",
			iface: customMultierr{},
			want: map[string]interface{}{
				"e": "great sadness",
				"eCauses": []interface{}{
					map[string]interface{}{"error": "foo"},
					map[string]interface{}{
						"error": "bar; baz",
						"errorCauses": []interface{}{
							map[string]interface{}{"error": "bar"},
							map[string]interface{}{"error": "baz"},
						},
					},
				},
			},
		},
		{
			msg:   "simple error",
			k:     "k",
			iface: fmt.Errorf("failed: %w", errors.New("egad")),
			want: map[string]interface{}{
				"k": "failed: egad",
			},
		},
		{
			msg: "multierr with causes",
			k:   "error",
			iface: multierr.Combine(
				fmt.Errorf("hello: %w",
					multierr.Combine(errors.New("foo"), errors.New("bar")),
				),
				errors.New("baz"),
				fmt.Errorf("world: %w", errors.New("qux")),
			),
			want: map[string]interface{}{
				"error": "hello: foo; bar; baz; world: qux",
				"errorCauses": []interface{}{
					map[string]interface{}{
						"error": "hello: foo; bar",
					},
					map[string]interface{}{"error": "baz"},
					map[string]interface{}{"error": "world: qux"},
				},
			},
		},
		{
			msg: "single error with marshal fields error",
			k:   "error",
			iface: customErrObject{
				name: "leaf",
			},
			want: map[string]interface{}{
				"error":      "error leaf",
				"errorError": "marshal leaf failed",
				"errorFields": map[string]interface{}{
					"leaf": "err",
				},
			},
		},
		{
			msg: "nested error with marshal fields error",
			k:   "error",
			iface: customErrObject{
				name: "top",
				underlying: customErrObject{
					name: "leaf",
				},
			},
			want: map[string]interface{}{
				"error":      "error top",
				"errorError": "marshal top failed; marshal leaf failed",
				"errorFields": map[string]interface{}{
					"top":  "err",
					"leaf": "err",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			if tt.t == UnknownType {
				tt.t = ErrorType
			}

			enc := NewMapObjectEncoder()
			f := Field{Key: tt.k, Type: tt.t, Interface: tt.iface}
			f.AddTo(enc)
			assert.Equal(t, tt.want, enc.Fields, "Unexpected output from field %+v.", f)
		})
	}
}

func TestRichErrorSupport(t *testing.T) {
	f := Field{
		Type:      ErrorType,
		Interface: fmt.Errorf("failed: %w", errors.New("egad")),
		Key:       "k",
	}
	enc := NewMapObjectEncoder()
	f.AddTo(enc)
	assert.Equal(t, "failed: egad", enc.Fields["k"], "Unexpected basic error message.")
}
