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

	"go.uber.org/zap"
	"go.uber.org/zap/spy"

	"github.com/stretchr/testify/assert"
)

func TestTeeLogsBoth(t *testing.T) {
	fac1, sink1 := spy.New(zap.DebugLevel)
	fac2, sink2 := spy.New(zap.WarnLevel)
	log := zap.New(zap.Tee(fac1, fac2))

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

func TestTee_Panic(t *testing.T) {
	fac1, sink1 := spy.New(zap.DebugLevel)
	fac2, sink2 := spy.New(zap.WarnLevel)
	log := zap.New(zap.Tee(fac1, fac2))

	assert.Panics(t, func() { log.Panic("foo") }, "tee logger.Panic panics")
	assert.Panics(t, func() { log.Check(zap.PanicLevel, "bar").Write() }, "tee logger.Check(PanicLevel).Write() panics")
	assert.NotPanics(t, func() { log.Log(zap.PanicLevel, "baz") }, "tee logger.Log(PanicLevel) does not panic")

	assert.Equal(t, []spy.Log{
		{
			Level:  zap.PanicLevel,
			Msg:    "foo",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.PanicLevel,
			Msg:    "bar",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.PanicLevel,
			Msg:    "baz",
			Fields: []zap.Field{},
		},
	}, sink1.Logs())

	assert.Equal(t, []spy.Log{
		{
			Level:  zap.PanicLevel,
			Msg:    "foo",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.PanicLevel,
			Msg:    "bar",
			Fields: []zap.Field{},
		},
		{
			Level:  zap.PanicLevel,
			Msg:    "baz",
			Fields: []zap.Field{},
		},
	}, sink2.Logs())
}

// XXX: we cannot presently write `func TestTee_Fatal(t *testing.T)`,
// because we can't have both a spy logger and an exit stub without a
// dependency cycle.
