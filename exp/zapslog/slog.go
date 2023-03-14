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

package zapslog

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
)

type handler struct {
	logger *zap.Logger
}

var _ slog.Handler = (*handler)(nil)

// group holds all the Attrs saved in a slog.GroupValue.
type group struct {
	attrs []slog.Attr
}

func (g *group) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range g.attrs {
		convertAttrToField(attr).AddTo(enc)
	}
	return nil
}

func convertAttrToField(attr slog.Attr) zapcore.Field {
	switch attr.Value.Kind() {
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		return zap.Object(attr.Key, &group{attrs: attr.Value.Group()})
	case slog.KindLogValuer:
		return convertAttrToField(slog.Attr{
			Key:   attr.Key,
			Value: attr.Value.Resolve(),
		})
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}

// convertSlogLevel maps slog Levels to zap Levels.
// Note that there is some room between slog levels while zap levels are continuous, so we can't 1:1 map them.
// See also https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md?pli=1#levels
func convertSlogLevel(l slog.Level) zapcore.Level {
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

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.logger.Core().Enabled(convertSlogLevel(level))
}

func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	fields := make([]zapcore.Field, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) {
		fields = append(fields, convertAttrToField(attr))
	})
	h.logger.Log(convertSlogLevel(record.Level), record.Message, fields...)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zapcore.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = convertAttrToField(attr)
	}
	return &handler{
		logger: h.logger.With(fields...),
	}

}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		logger: h.logger.With(zap.Namespace(name)),
	}
}

// New returns a *slog.Logger which writes to the supplied zap Logger.
func New(logger *zap.Logger) *slog.Logger {
	const slogCallerDepth = 1
	const loggerWriterDepth = 2
	return slog.New(&handler{
		logger: logger.WithOptions(zap.AddCallerSkip(slogCallerDepth + loggerWriterDepth)),
	})
}
