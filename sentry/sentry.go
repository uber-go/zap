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

package sentry

import (
	"os"
	"time"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"

	raven "github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

const (
	_platform          = "go"
	_traceContextLines = 3
	_traceSkipFrames   = 3
)

var _zapToRavenMap = map[zap.Level]raven.Severity{
	zap.DebugLevel: raven.INFO,
	zap.InfoLevel:  raven.INFO,
	zap.WarnLevel:  raven.WARNING,
	zap.ErrorLevel: raven.ERROR,
	zap.PanicLevel: raven.FATAL,
	zap.FatalLevel: raven.FATAL,
}

// Logger automatically sends logs above a certain threshold to Sentry.
type Logger struct {
	Capturer

	// Minimum level threshold for sending a Sentry event
	minLevel zap.Level

	// Controls if stack trace should be automatically generated, and how many frames to skip
	traceEnabled      bool
	traceSkipFrames   int
	traceContextLines int
	traceAppPrefixes  []string

	// Additional information that should be attached to each packet, i.e. customerUUID
	extra zwrap.KeyValueMap
}

// Option pattern for logger creation.
type Option func(l *Logger)

// New Sentry logger.
func New(dsn string, options ...Option) (*Logger, error) {
	l := &Logger{
		minLevel:          zap.ErrorLevel,
		traceEnabled:      true,
		traceSkipFrames:   _traceSkipFrames,
		traceContextLines: _traceContextLines,
		extra:             zwrap.KeyValueMap{},
	}

	for _, option := range options {
		option(l)
	}

	client, err := raven.New(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sentry client")
	}

	l.Capturer = &nonBlockingCapturer{Client: client}

	return l, nil
}

// MinLevel provides a minimum level threshold, above which Sentry packtes will be triggered
func MinLevel(level zap.Level) Option {
	return func(l *Logger) {
		l.minLevel = level
	}
}

// TraceEnabled allows to change wether (the somewhat expensive, relatively)
// stack trace generation takes place.
func TraceEnabled(enabled bool) Option {
	return func(l *Logger) {
		l.traceEnabled = enabled
	}
}

// TraceCfg allows to change the number of skipped frames, number of context lines and
// list of go prefixes that are considered "in-app", i.e. "github.com/uber-go/zap".
func TraceCfg(skip, context int, prefixes []string) Option {
	return func(l *Logger) {
		l.traceSkipFrames = skip
		l.traceContextLines = context
		l.traceAppPrefixes = prefixes
	}
}

// Extra stores additional information for each Sentry event.
func Extra(extra map[string]interface{}) Option {
	return func(l *Logger) {
		l.extra = extra
	}
}

// Log sends Sentry information provided minimum threshold is met.
func (l *Logger) Log(lvl zap.Level, msg string, fields ...zap.Field) {
	l.log(lvl, msg, fields)
}

// Debug is a nop.
func (l *Logger) Debug(msg string, fields ...zap.Field) {}

// Info sends Sentry information provided minimum threshold is met.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log(zap.InfoLevel, msg, fields)
}

// Warn sends Sentry information provided minimum threshold is met.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log(zap.WarnLevel, msg, fields)
}

// Error sends Sentry information provided minimum threshold is met.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log(zap.ErrorLevel, msg, fields)
}

// Panic sends Sentry information provided minimum threshold is met, then panics.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.log(zap.PanicLevel, msg, fields)
	panic(msg)
}

// Fatal sends Sentry information provided minimum threshold is met, then exits.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log(zap.FatalLevel, msg, fields)
	os.Exit(1)
}

// DFatal is a nop.
func (l *Logger) DFatal(msg string, fields ...zap.Field) {}

// With returns Sentry logger with additional context.
func (l *Logger) With(fields ...zap.Field) zap.Logger {
	// TODO(glib): this is terribly inefficient as a lot of logging messages
	// are going to be produced under the Error limit. Need to make sure
	// that the amount of work here is minimal...

	m := make(zwrap.KeyValueMap, len(l.extra)+len(fields))

	// add original fields
	// TODO(glib): deep copy here or change approach of passing maps around
	for k, v := range l.extra {
		m[k] = v
	}

	// add new fields
	for _, f := range fields {
		f.AddTo(m)
	}

	return &Logger{
		extra:             m,
		Capturer:          l.Capturer,
		minLevel:          l.minLevel,
		traceAppPrefixes:  l.traceAppPrefixes,
		traceContextLines: l.traceContextLines,
		traceEnabled:      l.traceEnabled,
		traceSkipFrames:   l.traceSkipFrames,
	}
}

// Notify sentry if the log level meets minimum threshold.
func (l *Logger) log(lvl zap.Level, msg string, fields []zap.Field) {
	if lvl >= l.minLevel {
		// append all the fields from the current log message to the stored ones
		extra := l.extra
		for _, f := range fields {
			f.AddTo(extra)
		}

		packet := &raven.Packet{
			Message:   msg,
			Timestamp: raven.Timestamp(time.Now()),
			Level:     _zapToRavenMap[lvl],
			Platform:  _platform,
			Extra:     extra,
		}

		if l.traceEnabled {
			trace := raven.NewStacktrace(l.traceSkipFrames, l.traceContextLines, l.traceAppPrefixes)
			if trace != nil {
				packet.Interfaces = append(packet.Interfaces, trace)
			}
		}

		l.Capture(packet)
	}
}

// Check returns a CheckedMessage logging the given message is Enabled, nil otherwise.
func (l *Logger) Check(lvl zap.Level, msg string) *zap.CheckedMessage {
	if lvl >= l.minLevel {
		return zap.NewCheckedMessage(l, lvl, msg)
	}
	return nil
}
