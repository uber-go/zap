// Package zapfield provides experimental zap.Field helpers whose APIs may be unstable.
package zapfield

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Str constructs a field with the given string-like key and value.
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

// Strs constructs a field that carries a slice of string-like values.
func Strs[K ~string, V ~[]S, S ~string](k K, v V) zap.Field {
	return zap.Array(string(k), stringArray[S](v))
}