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

package zbark

import (
	"bytes"
	"testing"

	"github.com/uber-go/zap"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-common/bark"
)

func newLogrus() (bark.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := logrus.New()
	logger.Out = buf
	logger.Level = logrus.DebugLevel
	logger.Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	return bark.NewLoggerFromLogrus(logger), buf
}

func newDebark(lvl zap.Level) (zap.Logger, *bytes.Buffer) {
	logrus, buf := newLogrus()
	return Debarkify(logrus, lvl), buf
}

func TestLogrusOutputIsTheSame(t *testing.T) {
	logrus, lbuf := newLogrus()
	debark, dbuf := newDebark(zap.DebugLevel)

	zfields := []zap.Field{
		zap.Bool("a", true),
		zap.Float64("b", float64(0.1)),
		zap.Int("c", 42),
		zap.String("d", "bar"),
		zap.Int64("e", int64(42)),
		zap.Stringer("g", stringable("g")),
	}
	const warning = "danger will robinson!"
	logrus.WithFields(zapToBark(zfields)).Warn(warning)
	debark.With(zfields...).Warn(warning)

	require.NotZero(t, lbuf.Len(), "logrus should have logged")
	require.NotZero(t, dbuf.Len(), "debarker should have logged")
	assert.True(t, bytes.Equal(lbuf.Bytes(), dbuf.Bytes()), "output should be the same")
}

func TestDebark_CastNoop(t *testing.T) {
	orig := zap.New(
		zap.NewJSONEncoder(),
		zap.DebugLevel,
		zap.DiscardOutput,
	)
	assert.True(t, orig == Debarkify(Barkify(orig), zap.DebugLevel))
}

func TestBarkify_CastNoop(t *testing.T) {
	b, _ := newLogrus()
	zap := Debarkify(b, zap.DebugLevel)
	b2 := Barkify(zap)

	require.Equal(t, b, b2, "Barkify(Debarkify(bark)) should equal bark")
	assert.True(t, b == b2, "%s", "expected barkify(debarkify(bark)) to return bark: %+v != %+v != Barkify(%+v)", b, b2, zap)
}

var levels = []zap.Level{
	zap.DebugLevel,
	zap.InfoLevel,
	zap.WarnLevel,
	zap.ErrorLevel,
}

func TestDebark_Check(t *testing.T) {
	logger, buf := newDebark(zap.DebugLevel)
	for _, l := range append(levels, zap.PanicLevel, zap.FatalLevel) {
		require.Equal(t, 0, buf.Len(), "buffer must be clean for %v", l)
		lc := logger.Check(l, "msg")
		require.NotNil(t, lc)
		assert.True(t, lc.OK())
		switch l {
		case zap.FatalLevel:
			continue
		case zap.PanicLevel:
			assert.Panics(t, func() { lc.Write() })
		default:
			lc.Write()
		}
		assert.NotEqual(t, 0, buf.Len())
		buf.Reset()
	}

	logger, buf = newDebark(zap.PanicLevel)
	for _, l := range levels {
		assert.Nil(t, logger.Check(l, "msg"))
	}

	// We should still panic even if the level isn't enough to log.
	logger, buf = newDebark(zap.FatalLevel)
	assert.Panics(t, func() {
		logger.Check(zap.PanicLevel, "panic!").Write()
	})
}

func TestDebark_LeveledLogging(t *testing.T) {
	logger, buf := newDebark(zap.DebugLevel)
	for _, l := range levels {
		require.Equal(t, 0, buf.Len(), "buffer not zero")
		logger.Log(l, "ohai")
		assert.NotEqual(t, 0, buf.Len(), "%q did not log", l)
		buf.Reset()
	}

	logger, buf = newDebark(zap.FatalLevel)
	require.Equal(t, 0, buf.Len(), "buffer not zero to begin test")
	for _, l := range append(levels) {
		logger.Log(l, "ohai")
		assert.Equal(t, 0, buf.Len(), "buffer not zero, we should not have logged")
	}

	logger, buf = newDebark(zap.DebugLevel)
	assert.Panics(t, func() { logger.Log(zap.Level(31337), "") })
	assert.Panics(t, func() { logger.Log(zap.PanicLevel, "") })
}

func TestDebark_Methods(t *testing.T) {
	logger, buf := newDebark(zap.DebugLevel)

	funcs := []func(string, ...zap.Field){
		logger.Debug,
		logger.Info,
		logger.Warn,
		logger.Error,
	}

	for i, f := range funcs {
		buf.Reset()
		assert.Equal(t, 0, buf.Len(), "buffer not zero")
		f("ohai")
		assert.NotEqual(t, 0, buf.Len(), "%+v(%d) did not log", f, i)
		buf.Reset()
	}
	assert.Panics(t, func() { logger.Panic("foo") })
	buf.Reset()

	logger, buf = newDebark(zap.FatalLevel)
	for i, f := range funcs {
		f("ohai")
		if !assert.Equal(t, 0, buf.Len(), "%+v(%d) logged, but shouldn't", f, i) {
			buf.Reset()
		}
	}
	assert.Panics(t, func() { logger.Panic("msg") })
}

func TestDebark_Stubs(t *testing.T) {
	logger, _ := newDebark(zap.DebugLevel)
	assert.NotPanics(t, func() { logger.DFatal("msg") })
}

func TestDebark_zapToBarkFields(t *testing.T) {
	logger, _ := newDebark(zap.DebugLevel)
	fields := []zap.Field{
		zap.Bool("a", true),
		zap.Float64("b", float64(0.1)),
		zap.Int("c", 42),
		zap.String("d", "bar"),
		zap.Int64("e", int64(42)),
		zap.Object("f", logger),
		zap.Stringer("g", stringable("g")),
	}
	assert.NotNil(t, logger.With(fields...))

	for _, f := range fields {
		assert.Len(t, zapToBark([]zap.Field{f}).Fields(), 1)
	}
	assert.Len(t, zapToBark(fields).Fields(), len(fields))
}
