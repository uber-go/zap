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
	zap.Meta

	sink    *Sink
	context []zap.Field
}

// New constructs a spy logger that collects spy.Log records to a Sink. It
// returns the logger and its sink.
//
// Options can change things like log level and initial fields, but any output
// related options will not be honored.
func New(options ...zap.Option) (*Logger, *Sink) {
	s := &Sink{}
	return &Logger{
		Meta: zap.MakeMeta(zap.NewJSONEncoder(zap.NoTime()), options...),
		sink: s,
	}, s
}

// With creates a new Logger with additional fields added to the logging context.
func (l *Logger) With(fields ...zap.Field) zap.Logger {
	return &Logger{
		Meta:    l.Meta.Clone(),
		sink:    l.sink,
		context: append(l.context, fields...),
	}
}

// Check returns a CheckedMessage if logging a particular message would succeed.
func (l *Logger) Check(lvl zap.Level, msg string) *zap.CheckedMessage {
	return l.Meta.Check(l, lvl, msg)
}

// Log writes a message at the specified level.
func (l *Logger) Log(lvl zap.Level, msg string, fields ...zap.Field) {
	l.log(lvl, msg, fields)
}

// Debug logs at the Debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log(zap.DebugLevel, msg, fields)
}

// Info logs at the Info level.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log(zap.InfoLevel, msg, fields)
}

// Warn logs at the Warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log(zap.WarnLevel, msg, fields)
}

// Error logs at the Error level.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log(zap.ErrorLevel, msg, fields)
}

// Panic logs at the Panic level. Note that the spy Logger doesn't actually
// panic.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.log(zap.PanicLevel, msg, fields)
}

// Fatal logs at the Fatal level. Note that the spy logger doesn't actuall call
// os.Exit.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log(zap.FatalLevel, msg, fields)
}

// DFatal logs at the Fatal level if the development flag is set, and the Error
// level otherwise.
func (l *Logger) DFatal(msg string, fields ...zap.Field) {
	if l.Development {
		l.log(zap.FatalLevel, msg, fields)
	} else {
		l.log(zap.ErrorLevel, msg, fields)
	}
}

func (l *Logger) log(lvl zap.Level, msg string, fields []zap.Field) {
	if l.Meta.Enabled(lvl) {
		l.sink.WriteLog(lvl, msg, l.allFields(fields))
	}
}

func (l *Logger) allFields(added []zap.Field) []zap.Field {
	all := make([]zap.Field, 0, len(added)+len(l.context))
	all = append(all, l.context...)
	all = append(all, added...)
	return all
}
