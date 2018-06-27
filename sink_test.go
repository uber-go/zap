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
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterSink(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		factory   func() (Sink, error)
		wantError bool
	}{
		{"valid", "valid", func() (Sink, error) { return nopCloserSink{os.Stdout}, nil }, false},
		{"empty", "", func() (Sink, error) { return nopCloserSink{os.Stdout}, nil }, true},
		{"stdout", "stdout", func() (Sink, error) { return nopCloserSink{os.Stdout}, nil }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterSink(tt.key, tt.factory)
			if tt.wantError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, _sinkFactories[tt.key], "expected the factory to be present")
			}
		})
	}
}

func TestNewSink(t *testing.T) {
	defer resetSinkRegistry()
	errTestSink := errors.New("test erroring")
	err := RegisterSink("errors", func() (Sink, error) { return nil, errTestSink })
	assert.Nil(t, err)
	tests := []struct {
		key string
		err error
	}{
		{"stdout", nil},
		{"errors", errTestSink},
		{"nonexistent", &errSinkNotFound{"nonexistent"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			_, err := newSink(tt.key)
			assert.Equal(t, tt.err, err)
		})
	}
}
