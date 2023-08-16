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
	"context"
	"log/slog"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Handler implements the slog.Handler by writing to a zap Core.
type Handler struct {
	core      zapcore.Core
	name      string // logger name
	addSource bool
}

// HandlerOptions are options for a Zap-based [slog.Handler].
type HandlerOptions struct {
	// LoggerName is used for log entries received from slog.
	//
	// Defaults to empty.
	LoggerName string

	// AddSource configures the handler to annotate each message with the filename,
	// line number, and function name.
	// AddSource is false by default to skip the cost of computing
	// this information.
	AddSource bool
}

// NewHandler builds a [Handler] that writes to the supplied [zapcore.Core]
// with the default options.
func NewHandler(core zapcore.Core, opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	return &Handler{
		core:      core,
		name:      opts.LoggerName,
		addSource: opts.AddSource,
	}
}

var _ slog.Handler = (*Handler)(nil)

// groupObject holds all the Attrs saved in a slog.GroupValue.
type groupObject []slog.Attr

func (gs groupObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range gs {
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
		return zap.Object(attr.Key, groupObject(attr.Value.Group()))
	case slog.KindLogValuer:
		return convertAttrToField(slog.Attr{
			Key: attr.Key,
			// TODO: resolve the value in a lazy way.
			// This probably needs a new Zap field type
			// that can be resolved lazily.
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

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.core.Enabled(convertSlogLevel(level))
}

// Handle handles the Record.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	ent := zapcore.Entry{
		Level:      convertSlogLevel(record.Level),
		Time:       record.Time,
		Message:    record.Message,
		LoggerName: h.name,
		// TODO: do we need to set the following fields?
		// Stack:
	}
	ce := h.core.Check(ent, nil)
	if ce == nil {
		return nil
	}

	if h.addSource && record.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
		if frame.PC != 0 {
			ce.Caller = zapcore.EntryCaller{
				Defined:  true,
				PC:       frame.PC,
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
			}
		}
	}

	fields := make([]zapcore.Field, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, convertAttrToField(attr))
		return true
	})
	ce.Write(fields...)
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zapcore.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = convertAttrToField(attr)
	}
	return h.withFields(fields...)
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *Handler) WithGroup(group string) slog.Handler {
	return h.withFields(zap.Namespace(group))
}

// withFields returns a cloned Handler with the given fields.
func (h *Handler) withFields(fields ...zapcore.Field) *Handler {
	cloned := *h
	cloned.core = h.core.With(fields)
	return &cloned
}
