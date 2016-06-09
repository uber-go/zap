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
	"time"

	"github.com/uber-go/zap"

	"github.com/uber-common/bark"
)

// Barkify wraps zap's JSON logger to make it compatible with the bark.Logger
// interface.
func Barkify(l zap.Logger) bark.Logger {
	if wrapper, ok := l.(*zapper); ok {
		return wrapper.bl
	}
	return &barker{
		zl:     l,
		fields: make(bark.Fields),
	}
}

type barker struct {
	zl     zap.Logger
	fields bark.Fields
}

func (l *barker) Debug(args ...interface{}) {
	l.zl.Debug(fmt.Sprint(args...))
}

func (l *barker) Debugf(format string, args ...interface{}) {
	l.zl.Debug(fmt.Sprintf(format, args...))
}

func (l *barker) Info(args ...interface{}) {
	l.zl.Info(fmt.Sprint(args...))
}

func (l *barker) Infof(format string, args ...interface{}) {
	l.zl.Info(fmt.Sprintf(format, args...))
}

func (l *barker) Warn(args ...interface{}) {
	l.zl.Warn(fmt.Sprint(args...))
}

func (l *barker) Warnf(format string, args ...interface{}) {
	l.zl.Warn(fmt.Sprintf(format, args...))
}

func (l *barker) Error(args ...interface{}) {
	l.zl.Error(fmt.Sprint(args...))
}

func (l *barker) Errorf(format string, args ...interface{}) {
	l.zl.Error(fmt.Sprintf(format, args...))
}

func (l *barker) Panic(args ...interface{}) {
	l.zl.Panic(fmt.Sprint(args...))
}

func (l *barker) Panicf(format string, args ...interface{}) {
	l.zl.Panic(fmt.Sprintf(format, args...))
}

func (l *barker) Fatal(args ...interface{}) {
	l.zl.Fatal(fmt.Sprint(args...))
}

func (l *barker) Fatalf(format string, args ...interface{}) {
	l.zl.Fatal(fmt.Sprintf(format, args...))
}

func (l *barker) WithField(key string, val interface{}) bark.Logger {
	newFields := bark.Fields{key: val}
	return &barker{
		zl:     l.addZapFields(newFields),
		fields: l.addBarkFields(newFields),
	}
}

func (l *barker) WithFields(fs bark.LogFields) bark.Logger {
	newFields := fs.Fields()
	return &barker{
		zl:     l.addZapFields(newFields),
		fields: l.addBarkFields(newFields),
	}
}

func (l *barker) Fields() bark.Fields {
	return l.fields
}

func (l *barker) addBarkFields(fs bark.Fields) bark.Fields {
	newFields := make(bark.Fields, len(l.fields)+len(fs))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fs {
		newFields[k] = v
	}
	return newFields
}

func (l *barker) addZapFields(fs bark.Fields) zap.Logger {
	zfs := make([]zap.Field, 0, len(fs))
	for key, val := range fs {
		switch v := val.(type) {
		case bool:
			zfs = append(zfs, zap.Bool(key, v))
		case float64:
			zfs = append(zfs, zap.Float64(key, v))
		case int:
			zfs = append(zfs, zap.Int(key, v))
		case int64:
			zfs = append(zfs, zap.Int64(key, v))
		case string:
			zfs = append(zfs, zap.String(key, v))
		case time.Time:
			zfs = append(zfs, zap.Time(key, v))
		case time.Duration:
			zfs = append(zfs, zap.Duration(key, v))
		// zap.Marshaler takes precedence over other interfaces.
		case zap.LogMarshaler:
			zfs = append(zfs, zap.Marshaler(key, v))
		case error:
			// zap.Err ignores the user-supplied key.
			zfs = append(zfs, zap.String(key, v.Error()))
		case fmt.Stringer:
			zfs = append(zfs, zap.Stringer(key, v))
		default:
			zfs = append(zfs, zap.Object(key, v))
		}
	}
	return l.zl.With(zfs...)
}
