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

package zbark

import (
	"fmt"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"

	"github.com/uber-common/bark"
)

type zapperBarkFields zwrap.KeyValueMap

// Debarkify wraps bark.Logger to make it compatible with zap's JSON logger
func Debarkify(bl bark.Logger, lvl zap.Level) zap.Logger {
	if wrapper, ok := bl.(*barker); ok {
		return wrapper.zl
	}
	return &zapper{
		Meta: zap.MakeMeta(nil, lvl),
		bl:   bl,
	}
}

type zapper struct {
	zap.Meta
	bl bark.Logger
}

func (z *zapper) DFatal(msg string, fields ...zap.Field) {
	// TODO: Implement development/DFatal?
	z.Error(msg, fields...)
}

func (z *zapper) Log(l zap.Level, msg string, fields ...zap.Field) {
	// NOTE: logging at panic and fatal level actually panic and exit the
	// process, meaning that bark loggers cannot compose well.
	switch l {
	case zap.PanicLevel, zap.FatalLevel:
	default:
		if !z.Meta.Enabled(l) {
			return
		}
	}
	bl := z.bl.WithFields(zapToBark(fields))
	switch l {
	case zap.DebugLevel:
		bl.Debug(msg)
	case zap.InfoLevel:
		bl.Info(msg)
	case zap.WarnLevel:
		bl.Warn(msg)
	case zap.ErrorLevel:
		bl.Error(msg)
	case zap.PanicLevel:
		bl.Panic(msg)
	case zap.FatalLevel:
		bl.Fatal(msg)
	default:
		panic(fmt.Errorf("passed an unknown zap.Level: %v", l))
	}
}

// Create a child logger, and optionally add some context to that logger.
func (z *zapper) With(fields ...zap.Field) zap.Logger {
	return &zapper{
		Meta: z.Meta,
		bl:   z.bl.WithFields(zapToBark(fields)),
	}
}

func (z *zapper) Check(l zap.Level, msg string) *zap.CheckedMessage {
	return z.Meta.Check(z, l, msg)
}

func (z *zapper) Debug(msg string, fields ...zap.Field) {
	z.Log(zap.DebugLevel, msg, fields...)
}

func (z *zapper) Info(msg string, fields ...zap.Field) {
	z.Log(zap.InfoLevel, msg, fields...)
}

func (z *zapper) Warn(msg string, fields ...zap.Field) {
	z.Log(zap.WarnLevel, msg, fields...)
}

func (z *zapper) Error(msg string, fields ...zap.Field) {
	z.Log(zap.ErrorLevel, msg, fields...)
}

func (z *zapper) Panic(msg string, fields ...zap.Field) {
	z.Log(zap.PanicLevel, msg, fields...)
}

func (z *zapper) Fatal(msg string, fields ...zap.Field) {
	z.Log(zap.FatalLevel, msg, fields...)
}

func (zbf zapperBarkFields) Fields() map[string]interface{} {
	return zbf
}

func zapToBark(zfs []zap.Field) bark.LogFields {
	zbf := make(zwrap.KeyValueMap, len(zfs))
	for _, zf := range zfs {
		zf.AddTo(zbf)
	}
	return zapperBarkFields(zbf)
}
