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
	"testing"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"

	raven "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

type fakeClient raven.Client

func TestBadDSN(t *testing.T) {
	_, err := New("123http://whoops")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create sentry client")
}

func TestEmptyDSN(t *testing.T) {
	l, err := New("")
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestWithLevels(t *testing.T) {
	l, err := New("", MinLevel(zap.InfoLevel))
	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.Equal(t, l.minLevel, zap.InfoLevel)
}

func TestExtra(t *testing.T) {
	extra := zwrap.KeyValueMap{
		"requestID":         "123h2eor2039423",
		"someInt":           123,
		"arrayOfManyThings": []int{1, 2, 3},
	}
	l, err := New("", Extra(extra))
	l.Error("error log")
	assert.NoError(t, err)
	assert.Equal(t, l.extra, extra)
}

func TestWith(t *testing.T) {
	l, err := New("", Extra(map[string]interface{}{
		"someInt": 123,
	}))
	expected := zwrap.KeyValueMap{"someInt": 123, "someFloat": float64(10)}
	l = l.With(zap.Float64("someFloat", float64(10))).(*Logger)
	assert.NoError(t, err)
	assert.Equal(t, l.extra, expected)
}

func TestWithTraceDisabled(t *testing.T) {
	_, ps := capturePackets(func(l *Logger) {
		l.Error("some error message", zap.String("foo", "bar"))
		l.Error("another error message")
	}, TraceEnabled(false))

	for _, p := range ps {
		assert.Empty(t, p.Interfaces)
	}
}

func TestTraceCfg(t *testing.T) {
	l, err := New("", TraceCfg(1, 7, []string{"github.com/uber-go/unicorns"}))
	assert.NoError(t, err)
	assert.Equal(t, l.traceSkipFrames, 1)
	assert.Equal(t, l.traceContextLines, 7)
	assert.Equal(t, l.traceAppPrefixes, []string{"github.com/uber-go/unicorns"})
}

func TestLevels(t *testing.T) {
	_, ps := capturePackets(func(l *Logger) {
		l.Log(zap.ErrorLevel, "direct call with error")
		l.Info("info")
		l.Warn("warn")
		l.Error("error")
		l.Panic("panic")
		l.Fatal("fatal")
	}, MinLevel(zap.FatalLevel))

	assert.Equal(t, len(ps), 1, "Only the fatal packet should be present")
}

func TestMeta(t *testing.T) {
	l, _ := New("")
	assert.Nil(t, l.Check(zap.InfoLevel, "info log"))

	c := l.Check(zap.ErrorLevel, "error message")
	assert.NotNil(t, c)
}

func TestErrorCapture(t *testing.T) {
	_, p := capturePacket(func(l *Logger) {
		l.Error("some error message", zap.String("foo", "bar"))
	})

	assert.Equal(t, p.Message, "some error message")
	assert.Equal(t, p.Extra, map[string]interface{}{"foo": "bar"})

	trace := p.Interfaces[0].(*raven.Stacktrace)
	assert.NotNil(t, trace.Frames)
}

func capturePacket(f func(l *Logger), options ...Option) (*Logger, *raven.Packet) {
	l, ps := capturePackets(f, options...)
	if len(ps) != 1 {
		panic("Expected to capture a packet, but didn't")
	}
	return l, ps[0]
}

func capturePackets(f func(l *Logger), options ...Option) (*Logger, []*raven.Packet) {
	c := &memCapturer{}
	options = append(options, WithCapturer(c))

	l, err := New("", options...)
	if err != nil {
		panic("Failed to create the logger")
	}

	f(l)

	return l, c.packets
}
