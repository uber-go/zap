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

package zap_test

import (
	"testing"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/spy"

	"github.com/stretchr/testify/assert"
)

func TestTeeLoggerLogsBoth(t *testing.T) {
	log1, sink1 := spy.New()
	log2, sink2 := spy.New()
	log1.SetLevel(zap.DebugLevel)
	log2.SetLevel(zap.WarnLevel)
	log := zap.TeeLogger(log1, log2)

	log.Log(zap.InfoLevel, "log @info")
	log.Log(zap.WarnLevel, "log @warn")

	log.Debug("log-dot-debug")
	log.Info("log-dot-info")
	log.Warn("log-dot-warn")
	log.Error("log-dot-error")

	assert.Equal(t, []spy.Log{
		{
			Level:  zap.InfoLevel,
			Msg:    "log @info",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.WarnLevel,
			Msg:    "log @warn",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.DebugLevel,
			Msg:    "log-dot-debug",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.InfoLevel,
			Msg:    "log-dot-info",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.WarnLevel,
			Msg:    "log-dot-warn",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.ErrorLevel,
			Msg:    "log-dot-error",
			Fields: []zap.Field{},
		},
	}, sink1.Logs())

	assert.Equal(t, []spy.Log{
		{
			Level:  zap.WarnLevel,
			Msg:    "log @warn",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.WarnLevel,
			Msg:    "log-dot-warn",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.ErrorLevel,
			Msg:    "log-dot-error",
			Fields: []zap.Field{},
		},
	}, sink2.Logs())
}
