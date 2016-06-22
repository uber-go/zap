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

package spy

import (
	"sync"

	"github.com/uber-go/zap"
)

// A Log is an encoding-agnostic representation of a log message.
type Log struct {
	Level  zap.Level
	Msg    string
	Fields []zap.Field
}

// A Sink stores Log structs.
type Sink struct {
	sync.Mutex

	logs []Log
}

// WriteLog writes a log message to the LogSink.
func (s *Sink) WriteLog(lvl zap.Level, msg string, fields []zap.Field) {
	s.Lock()
	log := Log{
		Msg:    msg,
		Level:  lvl,
		Fields: fields,
	}
	s.logs = append(s.logs, log)
	s.Unlock()
}

// Logs returns a copy of the sink's accumulated logs.
func (s *Sink) Logs() []Log {
	var logs []Log
	s.Lock()
	logs = append(logs, s.logs...)
	s.Unlock()
	return logs
}

// Logger satisfies zap.Logger, but makes testing convenient.
type Logger struct {
	sync.Mutex

	level       zap.Level
	sink        *Sink
	context     []zap.Field
	development bool
}

// New returns a new spy logger at the default level and its sink.
func New() (*Logger, *Sink) {
	s := &Sink{}
	return &Logger{
		// Use the same defaults as the core logger.
		level: zap.NewJSON().Level(),
		sink:  s,
	}, s
}

// StubTime is a no-op, since the spy logger omits time entirely.
func (l *Logger) StubTime() {}

// SetDevelopment sets the development flag.
func (l *Logger) SetDevelopment(dev bool) {
	l.Lock()
	l.development = dev
	l.Unlock()
}

// Level returns the currently-enabled logging level.
func (l *Logger) Level() zap.Level {
	l.Lock()
	defer l.Unlock()

	return l.level
}

// SetLevel alters the enabled log level.
func (l *Logger) SetLevel(new zap.Level) {
	l.Lock()
	defer l.Unlock()

	l.level = new
}

// With creates a new Logger with additional fields added to the logging context.
func (l *Logger) With(fields ...zap.Field) zap.Logger {
	return &Logger{
		level:       l.level,
		sink:        l.sink,
		context:     append(l.context, fields...),
		development: l.development,
	}
}

// Check returns a CheckedMessage if logging a particular message would succeed.
func (l *Logger) Check(lvl zap.Level, msg string) *zap.CheckedMessage {
	if !(lvl >= l.Level()) {
		return nil
	}
	return zap.NewCheckedMessage(l, lvl, msg)
}

// Log writes a message at the specified level.
func (l *Logger) Log(lvl zap.Level, msg string, fields ...zap.Field) {
	l.sink.WriteLog(lvl, msg, l.allFields(fields))
}

// Debug logs at the Debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.DebugLevel, msg, l.allFields(fields))
}

// Info logs at the Info level.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.InfoLevel, msg, l.allFields(fields))

}

// Warn logs at the Warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.WarnLevel, msg, l.allFields(fields))

}

// Error logs at the Error level.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.ErrorLevel, msg, l.allFields(fields))

}

// Panic logs at the Panic level. Note that the spy Logger doesn't actually
// panic.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.PanicLevel, msg, l.allFields(fields))

}

// Fatal logs at the Fatal level. Note that the spy logger doesn't actuall call
// os.Exit.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.sink.WriteLog(zap.FatalLevel, msg, l.allFields(fields))

}

// DFatal logs at the Fatal level if the development flag is set, and the Fatal
// level otherwise.
func (l *Logger) DFatal(msg string, fields ...zap.Field) {
	if l.development {
		l.sink.WriteLog(zap.FatalLevel, msg, l.allFields(fields))
	} else {
		l.sink.WriteLog(zap.ErrorLevel, msg, l.allFields(fields))
	}
}

func (l *Logger) allFields(added []zap.Field) []zap.Field {
	all := make([]zap.Field, 0, len(added)+len(l.context))
	all = append(all, l.context...)
	all = append(all, added...)
	return all
}
