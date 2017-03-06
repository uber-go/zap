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
	"fmt"
	"sync"

	"go.uber.org/zap/zapcore"
)

var (
	errNoEncoderNameSpecified = errors.New("no encoder name specified ")
	encoderNameToConstructor  = make(map[string]func(zapcore.EncoderConfig) zapcore.Encoder)
	encoderMutex              sync.RWMutex
)

func init() {
	registerDefaultEncoders()
}

// RegisterEncoder registers an encoder constructor for the given name.
//
// If an encoder with the same name already exists, this will panic.
// By default, the encoders "json" and "console" are registered.
func MustRegisterEncoder(name string, constructor func(zapcore.EncoderConfig) zapcore.Encoder) {
	if err := RegisterEncoder(name, constructor); err != nil {
		panic(err.Error())
	}
}

// RegisterEncoder registers an encoder constructor for the given name.
//
// If an encoder with the same name already exists, this will return an error.
// By default, the encoders "json" and "console" are registered.
func RegisterEncoder(name string, constructor func(zapcore.EncoderConfig) zapcore.Encoder) error {
	encoderMutex.Lock()
	defer encoderMutex.Unlock()
	if name == "" {
		return errNoEncoderNameSpecified
	}
	if _, ok := encoderNameToConstructor[name]; ok {
		return fmt.Errorf("encoder already registered for name %s", name)
	}
	encoderNameToConstructor[name] = constructor
	return nil
}

func registerDefaultEncoders() {
	MustRegisterEncoder("console", zapcore.NewConsoleEncoder)
	MustRegisterEncoder("json", zapcore.NewJSONEncoder)
}

func newEncoder(name string, encoderConfig zapcore.EncoderConfig) (zapcore.Encoder, error) {
	encoderMutex.RLock()
	defer encoderMutex.RUnlock()
	if name == "" {
		return nil, errNoEncoderNameSpecified
	}
	constructor, ok := encoderNameToConstructor[name]
	if !ok {
		return nil, fmt.Errorf("no encoder registered for name %s", name)
	}
	return constructor(encoderConfig), nil
}
