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

import "github.com/uber-go/atomic"

// TeeLogger creates a Logger that duplicates its log calls to two or
// more loggers. It is similar to io.MultiWriter.
//
// The returned logger will initially have the minimum level of all passed
// loggers. Changing the returned logger's level will change the level on all
// wrapped loggers.
//
// NOTE: Panic and Fatal may not work as expected, since the first core logger
// will itself Panic or Fatal (interrupting the TeeLogger loop).
func TeeLogger(logs ...Logger) Logger {
	switch len(logs) {
	case 0:
		return nil
	case 1:
		return logs[0]
	default:
		lvl := logs[0].Level()
		for _, log := range logs[1:] {
			if ll := log.Level(); ll < lvl {
				lvl = ll
			}
		}
		ml := &multiLogger{
			logs: logs,
			lvl:  atomic.NewInt32(int32(lvl)),
		}
		return ml
	}
}

type multiLogger struct {
	logs []Logger
	lvl  *atomic.Int32
}

func (ml multiLogger) Level() Level {
	return Level(ml.lvl.Load())
}

func (ml multiLogger) SetLevel(lvl Level) {
	for _, log := range ml.logs {
		log.SetLevel(lvl)
	}
	ml.lvl.Store(int32(lvl))
}

func (ml multiLogger) Log(lvl Level, msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Log(lvl, msg, fields...)
	}
}

func (ml multiLogger) Debug(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Debug(msg, fields...)
	}
}

func (ml multiLogger) Info(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Info(msg, fields...)
	}
}

func (ml multiLogger) Warn(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Warn(msg, fields...)
	}
}

func (ml multiLogger) Error(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Error(msg, fields...)
	}
}

func (ml multiLogger) Panic(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Panic(msg, fields...)
	}
}

func (ml multiLogger) Fatal(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.Fatal(msg, fields...)
	}
}

func (ml multiLogger) DFatal(msg string, fields ...Field) {
	for _, log := range ml.logs {
		log.DFatal(msg, fields...)
	}
}

func (ml multiLogger) With(fields ...Field) Logger {
	ml.logs = append([]Logger(nil), ml.logs...)
	for i, log := range ml.logs {
		ml.logs[i] = log.With(fields...)
	}
	return ml
}

func (ml multiLogger) Check(lvl Level, msg string) *CheckedMessage {
lvlSwitch:
	switch lvl {
	case PanicLevel, FatalLevel:
		// Panic and Fatal should always cause a panic/exit, even if the level
		// is disabled.
		break
	default:
		for _, log := range ml.logs {
			if cm := log.Check(lvl, msg); cm.OK() {
				break lvlSwitch
			}
		}
		return nil
	}
	return NewCheckedMessage(ml, lvl, msg)
}
