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

	"go.uber.org/zap"
	"go.uber.org/zap/zwrap"

	"github.com/uber-common/bark"
)

type zapperBarkFields zwrap.KeyValueMap

// Debarkify wraps bark.Logger to make it compatible with zap.logger.
func Debarkify(bl bark.Logger, lvl zap.Level) zap.Logger {
	if wrapper, ok := bl.(*barker); ok {
		return wrapper.zl
	}
	return zap.Neo(&barkFacility{
		Level: lvl,
		bl:    bl,
	})
}

type barkFacility struct {
	zap.Level
	bl bark.Logger
}

// Create a child logger, and optionally add some context to that logger.
func (bf *barkFacility) With(fields ...zap.Field) zap.Facility {
	return &barkFacility{
		bl: bf.bl.WithFields(zapToBark(fields)),
	}
}

func (bf *barkFacility) Log(ent zap.Entry, fields ...zap.Field) error {
	if bf.Enabled(ent.Level) {
		return bf.Write(ent, fields)
	}
	return nil
}

func (bf *barkFacility) Check(ent zap.Entry, ce *zap.CheckedEntry) *zap.CheckedEntry {
	if bf.Enabled(ent.Level) {
		ce = ce.AddFacility(ent, bf)
	}
	return ce
}

func (bf *barkFacility) Write(ent zap.Entry, fields []zap.Field) error {
	// NOTE: logging at panic and fatal level actually panic and exit the
	// process, meaning that bark loggers cannot compose well.
	bl := bf.bl.WithFields(zapToBark(fields))
	switch ent.Level {
	case zap.DebugLevel:
		bl.Debug(ent.Message)
	case zap.InfoLevel:
		bl.Info(ent.Message)
	case zap.WarnLevel:
		bl.Warn(ent.Message)
	case zap.ErrorLevel:
		bl.Error(ent.Message)
	case zap.DPanicLevel:
		bl.Error(ent.Message)
	case zap.PanicLevel:
		bl.Panic(ent.Message)
	case zap.FatalLevel:
		bl.Fatal(ent.Message)
	default:
		return fmt.Errorf("unable to map zap.Level %v to bark", ent.Level)
	}
	return nil
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
