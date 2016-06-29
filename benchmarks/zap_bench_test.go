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

package benchmarks

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"
)

var errExample = errors.New("fail")

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u user) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("name", u.Name)
	kv.AddString("email", u.Email)
	kv.AddInt64("created_at", u.CreatedAt.UnixNano())
	return nil
}

var _jane = user{
	Name:      "Jane Doe",
	Email:     "jane@test.com",
	CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

func fakeFields() []zap.Field {
	return []zap.Field{
		zap.Int("int", 1),
		zap.Int64("int64", 2),
		zap.Float64("float", 3.0),
		zap.String("string", "four!"),
		zap.Bool("bool", true),
		zap.Time("time", time.Unix(0, 0)),
		zap.Error(errExample),
		zap.Duration("duration", time.Second),
		zap.Marshaler("user-defined type", _jane),
		zap.String("another string", "done!"),
	}
}

func fakeMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("Test logging, but use a somewhat realistic message length. (#%v)", i)
	}
	return messages
}

func BenchmarkZapDisabledLevelsWithoutFields(b *testing.B) {
	logger := zap.NewJSON(zap.ErrorLevel, zap.Output(zap.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapDisabledLevelsAccumulatedContext(b *testing.B) {
	context := fakeFields()
	logger := zap.NewJSON(zap.ErrorLevel, zap.Output(zap.Discard), zap.Fields(context...))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapDisabledLevelsAddingFields(b *testing.B) {
	logger := zap.NewJSON(zap.ErrorLevel, zap.Output(zap.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.", fakeFields()...)
		}
	})
}

func BenchmarkZapDisabledLevelsCheckAddingFields(b *testing.B) {
	logger := zap.NewJSON(zap.ErrorLevel, zap.Output(zap.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if m := logger.Check(zap.InfoLevel, "Should be discarded."); m.OK() {
				m.Write(fakeFields()...)
			}
		}
	})
}

func BenchmarkZapAddingFields(b *testing.B) {
	logger := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.", fakeFields()...)
		}
	})
}

func BenchmarkZapWithAccumulatedContext(b *testing.B) {
	context := fakeFields()
	logger := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard), zap.Fields(context...))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkZapWithoutFields(b *testing.B) {
	logger := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.")
		}
	})
}

func BenchmarkZapSampleWithoutFields(b *testing.B) {
	messages := fakeMessages(1000)
	base := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	logger := zwrap.Sample(base, time.Second, 10, 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			logger.Info(messages[i%1000])
		}
	})
}

func BenchmarkZapSampleAddingFields(b *testing.B) {
	messages := fakeMessages(1000)
	base := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	logger := zwrap.Sample(base, time.Second, 10, 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			logger.Info(messages[i%1000], fakeFields()...)
		}
	})
}

func BenchmarkZapSampleCheckWithoutFields(b *testing.B) {
	messages := fakeMessages(1000)
	base := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	logger := zwrap.Sample(base, time.Second, 10, 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if cm := logger.Check(zap.InfoLevel, messages[i%1000]); cm.OK() {
				cm.Write()
			}
		}
	})
}

func BenchmarkZapSampleCheckAddingFields(b *testing.B) {
	messages := fakeMessages(1000)
	base := zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard))
	logger := zwrap.Sample(base, time.Second, 10, 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if m := logger.Check(zap.InfoLevel, messages[i%1000]); m.OK() {
				m.Write(fakeFields()...)
			}
		}
	})
}
