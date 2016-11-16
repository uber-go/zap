package zap_test

import (
	"testing"

	"github.com/uber-go/atomic"
	"github.com/uber-go/zap"
)

func BenchmarkCheckedMessage_Chain(b *testing.B) {
	infoLog := zap.New(
		zap.NullEncoder(),
		zap.InfoLevel,
		zap.DiscardOutput,
	)
	errorLog := zap.New(
		zap.NullEncoder(),
		zap.ErrorLevel,
		zap.DiscardOutput,
	)

	data := []struct {
		lvl zap.Level
		msg string
	}{
		{zap.DebugLevel, "meh"},
		{zap.InfoLevel, "fwiw"},
		{zap.ErrorLevel, "hey!"},
	}

	p := atomic.NewInt32(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		j := p.Inc()
		myInfoLog := infoLog.With(zap.Int("p", int(j)))
		myErrorLog := errorLog.With(zap.Int("p", int(j)))
		for pb.Next() {
			d := data[i%len(data)]
			cm := myInfoLog.Check(d.lvl, d.msg)
			cm = cm.Chain(myErrorLog, d.lvl, d.msg)
			if cm.OK() {
				cm.Write(zap.Int("i", i))
			}
			i++
		}
	})
}

func BenchmarkCheckedMessage_Chain_sliceLoggers(b *testing.B) {
	logs := []zap.Logger{
		zap.New(
			zap.NullEncoder(),
			zap.InfoLevel,
			zap.DiscardOutput,
		),
		zap.New(
			zap.NullEncoder(),
			zap.ErrorLevel,
			zap.DiscardOutput,
		),
	}

	data := []struct {
		lvl zap.Level
		msg string
	}{
		{zap.DebugLevel, "meh"},
		{zap.InfoLevel, "fwiw"},
		{zap.ErrorLevel, "hey!"},
	}

	p := atomic.NewInt32(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		j := p.Inc()
		myLogs := make([]zap.Logger, len(logs))
		for i, log := range logs {
			myLogs[i] = log.With(zap.Int("p", int(j)))
		}
		i := 0
		for pb.Next() {
			d := data[i%len(data)]
			var cm *zap.CheckedMessage
			for _, log := range myLogs {
				cm = cm.Chain(log, d.lvl, d.msg)
			}
			if cm.OK() {
				cm.Write(zap.Int("i", i))
			}
			i++
		}
	})
}
