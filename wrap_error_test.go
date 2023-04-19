// Copyright (c) 2023 Uber Technologies, Inc.
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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestWrapError(t *testing.T) {
	var (
		rootErr = errors.New("root err")
		wrap1   = fmt.Errorf("wrap1: %w", rootErr)
		wrap2   = WrapError(wrap1,
			String("user", "foo"),
			Int("count", 12),
		)
	)
	assert.True(t, errors.Is(wrap2, rootErr), "errors.Is")
	assert.True(t, errors.Is(wrap2, wrap1), "errors.Is")

	enc := zapcore.NewMapObjectEncoder()
	Error(wrap2).AddTo(enc)
	assert.Equal(t, map[string]any{
		"error": "wrap1: root err",
		"errorFields": map[string]any{
			"user":  "foo",
			"count": int64(12),
		},
	}, enc.Fields)

	var (
		wrap3 = fmt.Errorf("wrap3: %w", wrap2)
		wrap4 = WrapError(wrap3, Bool("wrap4", true))
	)
	Error(wrap4).AddTo(enc)
	assert.Equal(t, map[string]any{
		"error": "wrap3: wrap1: root err",
		"errorFields": map[string]any{
			"user":  "foo",
			"count": int64(12),
			"wrap4": true,
		},
	}, enc.Fields)

	var (
		wrap5 = fmt.Errorf("wrap5 no wrap: %v", wrap3)
		wrap6 = WrapError(wrap5, Bool("wrap5", true))
	)
	Error(wrap6).AddTo(enc)
	assert.Equal(t, map[string]any{
		"error": "wrap5 no wrap: wrap3: wrap1: root err",
		"errorFields": map[string]any{
			"wrap5": true,
		},
	}, enc.Fields)
}

func TestWrapErrorDuplicateField(t *testing.T) {
	var (
		rootErr = errors.New("root err")
		wrap1   = WrapError(rootErr, String("f1", "a"), String("f2", "b"))
		wrap2   = WrapError(wrap1, String("f1", "c"))
	)
	enc := zapcore.NewMapObjectEncoder()
	Error(wrap2).AddTo(enc)
	assert.Equal(t, map[string]any{
		"error": "root err",
		"errorFields": map[string]any{
			// fields are added in Unwrap order, and last added field wins in the map encoder
			// which is the first field added.
			"f1": "a",
			"f2": "b",
		},
	}, enc.Fields)
}
