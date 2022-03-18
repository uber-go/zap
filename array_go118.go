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

import "go.uber.org/zap/zapcore"

// Objects constructs a field with the given key, holding a list of the
// provided objects that can be marshaled by Zap.
//
// For example, given a struct User that can be marshaled with zap.Object,
//
//  type User struct{ ... }
//
//  func (u *User) MarshalLogObject(enc zapcore.ObjectEncoder) error
//
// Use Objects like so:
//
//  logger.Info("found users",
//    zapmarshal.Objects("users", []*User{u1, u2, u3}))
func Objects[T zapcore.ObjectMarshaler](key string, values []T) Field {
	return Array(key, objects[T](values))
}

type objects[T zapcore.ObjectMarshaler] []T

func (os objects[T]) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for _, o := range os {
		if err := arr.AppendObject(o); err != nil {
			return err
		}
	}
	return nil
}
