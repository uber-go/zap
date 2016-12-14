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

	"go.uber.org/zap"
)

// A Log is an encoding-agnostic representation of a log message.
type Log struct {
	Level  zap.Level
	Msg    string
	Fields []zap.Field
}

// A Sink stores Log structs.
type Sink struct {
	mu   sync.Mutex
	logs []Log
}

// WriteLog writes a log message to the LogSink.
func (s *Sink) WriteLog(lvl zap.Level, msg string, fields []zap.Field) {
	s.mu.Lock()
	log := Log{
		Msg:    msg,
		Level:  lvl,
		Fields: fields,
	}
	s.logs = append(s.logs, log)
	s.mu.Unlock()
}

// Logs returns a copy of the sink's accumulated logs.
func (s *Sink) Logs() []Log {
	var logs []Log
	s.mu.Lock()
	logs = append(logs, s.logs...)
	s.mu.Unlock()
	return logs
}

// Facility implements a zap.Facility that captures Log records.
type Facility struct {
	zap.LevelEnabler
	sink    *Sink
	context []zap.Field
}

// With creates a sub spy facility so that all log records recorded under it
// have the given fields attached.
func (sf *Facility) With(fields []zap.Field) zap.Facility {
	return &Facility{
		LevelEnabler: sf.LevelEnabler,
		sink:         sf.sink,
		context:      append(sf.context[:len(sf.context):len(sf.context)], fields...),
	}
}

// Log writes the entry if its level is enabled.
func (sf *Facility) Log(ent zap.Entry, fields []zap.Field) error {
	if sf.Enabled(ent.Level) {
		return sf.Write(ent, fields)
	}
	return nil
}

// Write collects all contextual fields and records a Log record.
func (sf *Facility) Write(ent zap.Entry, fields []zap.Field) error {
	all := make([]zap.Field, 0, len(fields)+len(sf.context))
	all = append(all, sf.context...)
	all = append(all, fields...)
	sf.sink.WriteLog(ent.Level, ent.Message, all)
	return nil
}

// Check adds this spy facility to the CheckedEntry, creating one if necessary,
// if the Entry's level is enabled.
func (sf *Facility) Check(ent zap.Entry, ce *zap.CheckedEntry) *zap.CheckedEntry {
	if sf.Enabled(ent.Level) {
		ce = ce.AddFacility(ent, sf)
	}
	return ce
}

// New creates a new Facility and returns it and its associated Sink.
func New(enab zap.LevelEnabler) (zap.Facility, *Sink) {
	fac := &Facility{
		LevelEnabler: enab,
		sink:         &Sink{},
	}
	return fac, fac.sink
}
