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

// Tee creates a Logger that duplicates its log calls to two or more
// loggers. It is similar to io.MultiWriter.
//
// For each logging level method (.Debug, .Info, etc), the Tee calls
// each sub-logger's level method.
//
// Exceptions are made for the DPanic, Panic, and Fatal methods: the returned
// logger calls .Log(DPanicLevel, ...), .Log(PanicLevel, ...), and
// .Log(FatalLevel, ...) respectively. Only after all sub-loggers have received
// the message, then the Tee terminates the process (using os.Exit or panic()
// per usual semantics).
//
// NOTE: DPanic will currently never panic, since the Tee Logger does not
// accept options (nor even have a development flag).
//
// Check returns a CheckedMessage chain of any OK CheckedMessages returned by
// all sub-loggers. The returned message is OK if any of the sub-messages are.
// An exception is made for FatalLevel and PanicLevel, where a CheckedMessage
// is returned against the Tee itself. This is so that tlog.Check(PanicLevel,
// ...).Write(...) is equivalent to tlog.Panic(...) (likewise for FatalLevel).
func Tee(logs ...Logger) Logger {
	switch len(logs) {
	case 0:
		return nil
	case 1:
		return logs[0]
	default:
		return multiLogger(logs)
	}
}

type multiLogger []Logger

func (ml multiLogger) Log(lvl Level, msg string, fields ...Field) {
	ml.log(lvl, msg, fields)
}

func (ml multiLogger) Debug(msg string, fields ...Field) {
	ml.log(DebugLevel, msg, fields)
}

func (ml multiLogger) Info(msg string, fields ...Field) {
	ml.log(InfoLevel, msg, fields)
}

func (ml multiLogger) Warn(msg string, fields ...Field) {
	ml.log(WarnLevel, msg, fields)
}

func (ml multiLogger) Error(msg string, fields ...Field) {
	ml.log(ErrorLevel, msg, fields)
}

func (ml multiLogger) Panic(msg string, fields ...Field) {
	ml.log(PanicLevel, msg, fields)
	panic(msg)
}

func (ml multiLogger) Fatal(msg string, fields ...Field) {
	ml.log(FatalLevel, msg, fields)
	_exit(1)
}

func (ml multiLogger) log(lvl Level, msg string, fields []Field) {
	for _, log := range ml {
		log.Log(lvl, msg, fields...)
	}
}

func (ml multiLogger) DPanic(msg string, fields ...Field) {
	ml.log(DPanicLevel, msg, fields)
	// TODO: Implement development/DPanic?
}

func (ml multiLogger) With(fields ...Field) Logger {
	clone := make(multiLogger, len(ml))
	for i := range ml {
		clone[i] = ml[i].With(fields...)
	}
	return clone
}

func (ml multiLogger) Check(lvl Level, msg string) *CheckedMessage {
	switch lvl {
	case FatalLevel, PanicLevel:
		// need to end up calling multiLogger Fatal and Panic methods, to avoid
		// sub-logger termination (by merely logging at FatalLevel and
		// PanicLevel).
		return NewCheckedMessage(ml, lvl, msg)
	}
	var cm *CheckedMessage
	for _, log := range ml {
		cm = cm.Chain(log.Check(lvl, msg))
	}
	return cm
}
