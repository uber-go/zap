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

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/internal/observer"
	"go.uber.org/zap/zapcore"
)

func TestReplaceGlobals(t *testing.T) {
	initialL := *L
	initialS := *S

	withLogger(t, DebugLevel, nil, func(l *Logger, logs *observer.ObservedLogs) {
		L.Info("no-op")
		S.Info("no-op")
		assert.Equal(t, 0, logs.Len(), "Expected initial logs to go to default no-op global.")

		defer ReplaceGlobals(l)()

		L.Info("captured")
		S.Info("captured")
		expected := observer.LoggedEntry{
			Entry:   zapcore.Entry{Message: "captured"},
			Context: []zapcore.Field{},
		}
		assert.Equal(
			t,
			[]observer.LoggedEntry{expected, expected},
			logs.AllUntimed(),
			"Unexpected global log output.",
		)
	})

	assert.Equal(t, initialL, *L, "Expected func returned from ReplaceGlobals to restore initial L.")
	assert.Equal(t, initialS, *S, "Expected func returned from ReplaceGlobals to restore initial S.")
}

func TestRedirectStdLog(t *testing.T) {
	initialFlags := log.Flags()
	initialPrefix := log.Prefix()

	withLogger(t, DebugLevel, nil, func(l *Logger, logs *observer.ObservedLogs) {
		defer RedirectStdLog(l)()
		log.Print("redirected")

		assert.Equal(t, []observer.LoggedEntry{{
			Entry:   zapcore.Entry{Message: "redirected"},
			Context: []zapcore.Field{},
		}}, logs.AllUntimed(), "Unexpected global log output.")
	})

	assert.Equal(t, initialFlags, log.Flags(), "Expected to reset initial flags.")
	assert.Equal(t, initialPrefix, log.Prefix(), "Expected to reset initial prefix.")
}
