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

package zapcore_test

import (
	"testing"

	"go.uber.org/zap/internal/observer"
	. "go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

func withTee(f func(core Core, debugLogs, warnLogs *observer.ObservedLogs)) {
	var debugLogs, warnLogs observer.ObservedLogs
	tee := NewTee(
		observer.New(DebugLevel, debugLogs.Add, true),
		observer.New(WarnLevel, warnLogs.Add, true),
	)
	f(tee, &debugLogs, &warnLogs)
}

func TestTeeUnusualInput(t *testing.T) {
	// Verify that Tee handles receiving one and no inputs correctly.
	t.Run("one input", func(t *testing.T) {
		obs := observer.New(DebugLevel, nil, true)
		assert.Equal(t, obs, NewTee(obs), "Expected to return single inputs unchanged.")
	})
	t.Run("no input", func(t *testing.T) {
		assert.Equal(t, NewNopCore(), NewTee(), "Expected to return NopCore.")
	})
}

func TestTeeCheck(t *testing.T) {
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		debugEntry := Entry{Level: DebugLevel, Message: "log-at-debug"}
		infoEntry := Entry{Level: InfoLevel, Message: "log-at-info"}
		warnEntry := Entry{Level: WarnLevel, Message: "log-at-warn"}
		errorEntry := Entry{Level: ErrorLevel, Message: "log-at-error"}
		for _, ent := range []Entry{debugEntry, infoEntry, warnEntry, errorEntry} {
			if ce := tee.Check(ent, nil); ce != nil {
				ce.Write()
			}
		}

		assert.Equal(t, []observer.LoggedEntry{
			{Entry: debugEntry, Context: []Field{}},
			{Entry: infoEntry, Context: []Field{}},
			{Entry: warnEntry, Context: []Field{}},
			{Entry: errorEntry, Context: []Field{}},
		}, debugLogs.All())

		assert.Equal(t, []observer.LoggedEntry{
			{Entry: warnEntry, Context: []Field{}},
			{Entry: errorEntry, Context: []Field{}},
		}, warnLogs.All())
	})
}

func TestTeeWrite(t *testing.T) {
	// Calling the tee's Write method directly should always log, regardless of
	// the configured level.
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		debugEntry := Entry{Level: DebugLevel, Message: "log-at-debug"}
		warnEntry := Entry{Level: WarnLevel, Message: "log-at-warn"}
		for _, ent := range []Entry{debugEntry, warnEntry} {
			tee.Write(ent, nil)
		}

		for _, logs := range []*observer.ObservedLogs{debugLogs, warnLogs} {
			assert.Equal(t, []observer.LoggedEntry{
				{Entry: debugEntry, Context: []Field{}},
				{Entry: warnEntry, Context: []Field{}},
			}, logs.All())
		}
	})
}

func TestTeeWith(t *testing.T) {
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		f := makeInt64Field("k", 42)
		tee = tee.With([]Field{f})
		ent := Entry{Level: WarnLevel, Message: "log-at-warn"}
		if ce := tee.Check(ent, nil); ce != nil {
			ce.Write()
		}

		for _, logs := range []*observer.ObservedLogs{debugLogs, warnLogs} {
			assert.Equal(t, []observer.LoggedEntry{
				{Entry: ent, Context: []Field{f}},
			}, logs.All())
		}
	})
}

func TestTeeEnabled(t *testing.T) {
	tee := NewTee(
		observer.New(InfoLevel, nil, false),
		observer.New(WarnLevel, nil, false),
	)
	tests := []struct {
		lvl     Level
		enabled bool
	}{
		{DebugLevel, false},
		{InfoLevel, true},
		{WarnLevel, true},
		{ErrorLevel, true},
		{DPanicLevel, true},
		{PanicLevel, true},
		{FatalLevel, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.enabled, tee.Enabled(tt.lvl), "Unexpected Enabled result for level %s.", tt.lvl)
	}
}
