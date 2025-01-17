// Copyright (c) 2016-2023 Uber Technologies, Inc.
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

package zapfield

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Uint constructs a field with the given string-like key and uint-like value.
func Uint[K ~string, V ~uint](k K, v V) zap.Field {
	return zap.Uint(string(k), uint(v))
}

// Uint64 constructs a field with the given string-like key and uint64-like value.
func Uint64[K ~string, V ~uint64](k K, v V) zap.Field {
	return zap.Uint64(string(k), uint64(v))
}

// Uint32 constructs a field with the given string-like key and uint32-like value.
func Uint32[K ~string, V ~uint32](k K, v V) zap.Field {
	return zap.Uint32(string(k), uint32(v))
}

// Uint16 constructs a field with the given string-like key and uint16-like value.
func Uint16[K ~string, V ~uint16](k K, v V) zap.Field {
	return zap.Uint16(string(k), uint16(v))
}

// Uint8 constructs a field with the given string-like key and uint8-like value.
func Uint8[K ~string, V ~uint8](k K, v V) zap.Field {
	return zap.Uint8(string(k), uint8(v))
}

// -----

// Int constructs a field with the given string-like key and int-like value.
func Int[K ~string, V ~int](k K, v V) zap.Field {
	return zap.Int(string(k), int(v))
}

// Int64 constructs a field with the given string-like key and int64-like value.
func Int64[K ~string, V ~int64](k K, v V) zap.Field {
	return zap.Int64(string(k), int64(v))
}

// Int32 constructs a field with the given string-like key and int32-like value.
func Int32[K ~string, V ~int32](k K, v V) zap.Field {
	return zap.Int32(string(k), int32(v))
}

// Int16 constructs a field with the given string-like key and int16-like value.
func Int16[K ~string, V ~int16](k K, v V) zap.Field {
	return zap.Int16(string(k), int16(v))
}

// Int8 constructs a field with the given string-like key and int8-like value.
func Int8[K ~string, V ~int8](k K, v V) zap.Field {
	return zap.Int8(string(k), int8(v))
}

// -----

// Float64 constructs a field with the given string-like key and float64-like value.
func Float64[K ~string, V ~float64](k K, v V) zap.Field {
	return zap.Float64(string(k), float64(v))
}

// Float32 constructs a field with the given string-like key and float32-like value.
func Float32[K ~string, V ~float32](k K, v V) zap.Field {
	return zap.Float32(string(k), float32(v))
}

// -----

// String constructs a field with the given string-like key and value.
func String[K ~string, V ~string](k K, v V) zap.Field {
	return zap.String(string(k), string(v))
}

// Str is an alias to String.
func Str[K ~string, V ~string](k K, v V) zap.Field {
	return zap.String(string(k), string(v))
}

type stringArray[T ~string] []T

func (a stringArray[T]) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for i := range a {
		enc.AppendString(string(a[i]))
	}
	return nil
}

// Strings constructs a field that carries a slice of string-like values.
func Strings[K ~string, V ~[]S, S ~string](k K, v V) zap.Field {
	return zap.Array(string(k), stringArray[S](v))
}

// Strs is an alias to Strings.
func Strs[K ~string, V ~[]S, S ~string](k K, v V) zap.Field {
	return zap.Array(string(k), stringArray[S](v))
}
