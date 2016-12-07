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

type filterLogger [FatalLevel - DebugLevel + 1]multiLogger

// LeveledLogger is a logger with specific level
type LeveledLogger struct {
	Level  Level
	Logger Logger
}

// Filter creates a logger that filter its log according to the level set in LeveledLogger.
// This would be useful when seperating log into distinct files by its level. Different
// loggers can have same level, the message would be written to all subloggers by multiLogger.
//
// It is inspired by the implementation of Tee
func Filter(logs ...LeveledLogger) Logger {
	if len(logs) == 0 {
		return nil
	}

	var fl filterLogger
	for _, log := range logs {
		idx := log.Level - DebugLevel
		fl[idx] = append(fl[idx], log.Logger)
	}
	return fl
}

func (fl filterLogger) Log(lvl Level, msg string, fields ...Field) {
	fl.log(lvl, msg, fields)
}

func (fl filterLogger) Debug(msg string, fields ...Field) {
	fl.log(DebugLevel, msg, fields)
}

func (fl filterLogger) Info(msg string, fields ...Field) {
	fl.log(InfoLevel, msg, fields)
}

func (fl filterLogger) Warn(msg string, fields ...Field) {
	fl.log(WarnLevel, msg, fields)
}

func (fl filterLogger) Error(msg string, fields ...Field) {
	fl.log(ErrorLevel, msg, fields)
}

func (fl filterLogger) Panic(msg string, fields ...Field) {
	fl.log(PanicLevel, msg, fields)
	panic(msg)
}

func (fl filterLogger) Fatal(msg string, fields ...Field) {
	fl.log(FatalLevel, msg, fields)
	_exit(1)
}

func (fl filterLogger) log(lvl Level, msg string, fields []Field) {
	for _, log := range fl[lvl-DebugLevel] {
		log.Log(lvl, msg, fields...)
	}
}

func (fl filterLogger) DPanic(msg string, fields ...Field) {
	fl.log(DPanicLevel, msg, fields)
	// TODO: Implement development/DPanic?
}

func (fl filterLogger) With(fields ...Field) Logger {
	var clone filterLogger
	for i := range fl {
		clone[i] = make(multiLogger, len(fl[i]))
		for j := range fl[i] {
			clone[i][j] = fl[i][j].With(fields...)
		}
	}
	return clone
}

func (fl filterLogger) Check(lvl Level, msg string) *CheckedMessage {
	switch lvl {
	case FatalLevel, PanicLevel:
		// need to end up calling multiLogger Fatal and Panic methods, to avoid
		// sub-logger termination (by merely logging at FatalLevel and
		// PanicLevel).
		return NewCheckedMessage(fl, lvl, msg)
	}
	var cm *CheckedMessage
	for _, log := range fl[lvl-DebugLevel] {
		cm = cm.Chain(log.Check(lvl, msg))
	}
	return cm
}
