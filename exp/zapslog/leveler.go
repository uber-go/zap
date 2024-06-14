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

//go:build go1.21

package zapslog

import (
	"log/slog"

	"go.uber.org/zap/zapcore"
)

// ConvertLeveler maps from [log/slog.Level] to [go.uber.org/zap/zapcore.Level].
// Note that there is some room between slog levels while zap levels are continuous, so we can't 1:1 map them.
// See also [structured logging proposal]
//
// [structured logging proposal]: https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md?pli=1#levels
type ConvertLeveler interface {
	ConvertLevel(l slog.Level) zapcore.Level
}

// DefaultConvertLeveler static maps from [log/slog.Level] to [go.uber.org/zap/zapcore.Level].
// implements: [go.uber.org/zap/exp/zapslog.ConvertLeveler]
type DefaultConvertLeveler struct{}

// ConvertLevel static maps from [log/slog.Level] to [go.uber.org/zap/zapcore.Level].
//   - [log/slog.LevelError] to [go.uber.org/zap/zapcore.ErrorLevel]
//   - [log/slog.LevelWarn] to [go.uber.org/zap/zapcore.WarnLevel]
//   - [log/slog.LevelInfo] to [go.uber.org/zap/zapcore.InfoLevel]
//   - [log/slog.LevelDebug] or default to [go.uber.org/zap/zapcore.DebugLevel]
//
// Note that there is some room between slog levels while zap levels are continuous, so we can't 1:1 map them.
// See also [structured logging proposal]
//
// [structured logging proposal]: https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md?pli=1#levels
func (c *DefaultConvertLeveler) ConvertLevel(l slog.Level) zapcore.Level {
	switch {
	case l >= slog.LevelError:
		return zapcore.ErrorLevel
	case l >= slog.LevelWarn:
		return zapcore.WarnLevel
	case l >= slog.LevelInfo:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}
