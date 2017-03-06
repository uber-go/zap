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

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

var (
	newNilEncoder = func(_ zapcore.EncoderConfig) zapcore.Encoder {
		return nil
	}
)

func TestRegisterDefaultEncoders(t *testing.T) {
	testEncodersRegistered(t, "console", "json")
}

func TestRegisterEncoder(t *testing.T) {
	testEncoders(func() {
		assert.NoError(t, RegisterEncoder("foo", newNilEncoder))
		testEncodersRegistered(t, "foo")
	})
}

func TestDuplicateRegisterEncoder(t *testing.T) {
	testEncoders(func() {
		RegisterEncoder("foo", newNilEncoder)
		assert.Error(t, RegisterEncoder("foo", newNilEncoder))
	})
}

func TestDuplicateMustRegisterEncoder(t *testing.T) {
	testEncoders(func() {
		RegisterEncoder("foo", newNilEncoder)
		assert.Panics(t, func() {
			MustRegisterEncoder("foo", newNilEncoder)
		})
	})
}

func TestRegisterEncoderNoName(t *testing.T) {
	assert.Equal(t, errNoEncoderNameSpecified, RegisterEncoder("", newNilEncoder))
}

func TestNewEncoder(t *testing.T) {
	testEncoders(func() {
		RegisterEncoder("foo", newNilEncoder)
		encoder, err := newEncoder("foo", zapcore.EncoderConfig{})
		assert.NoError(t, err)
		assert.Nil(t, encoder)
	})
}

func TestNewEncoderNotRegistered(t *testing.T) {
	_, err := newEncoder("foo", zapcore.EncoderConfig{})
	assert.Error(t, err)
}

func TestNewEncoderNoName(t *testing.T) {
	_, err := newEncoder("", zapcore.EncoderConfig{})
	assert.Equal(t, errNoEncoderNameSpecified, err)
}

func testEncoders(f func()) {
	encoderNameToConstructor = make(map[string]func(zapcore.EncoderConfig) zapcore.Encoder)
	defer func() {
		encoderNameToConstructor = make(map[string]func(zapcore.EncoderConfig) zapcore.Encoder)
		registerDefaultEncoders()
	}()
	f()
}

func testEncodersRegistered(t *testing.T, names ...string) {
	assert.Len(t, encoderNameToConstructor, len(names))
	for _, name := range names {
		assert.NotNil(t, encoderNameToConstructor[name])
	}
}
