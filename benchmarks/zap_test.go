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
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

var (
	errExample = errors.New("fail")

	_messages = fakeMessages(1000)
)

func fakeMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("Test logging, but use a somewhat realistic message length. (#%v)", i)
	}
	return messages
}

func getMessage(iter int) string {
	return _messages[iter%1000]
}

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u user) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", u.Name)
	enc.AddString("email", u.Email)
	enc.AddInt64("created_at", u.CreatedAt.UnixNano())
	return nil
}

var _jane = user{
	Name:      "Jane Doe",
	Email:     "jane@test.com",
	CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

func newZapLogger(lvl zapcore.Level) *zap.Logger {
	// use the canned production encoder configuration
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	return zap.New(zapcore.NewCore(
		enc,
		&zaptest.Discarder{},
		lvl,
	))
}

func newSampledLogger(lvl zapcore.Level) *zap.Logger {
	return zap.New(zapcore.NewSampler(
		newZapLogger(zap.DebugLevel).Core(),
		100*time.Millisecond,
		10, // first
		10, // thereafter
	))
}

func fakeFields() []zapcore.Field {
	return []zapcore.Field{
		zap.Int("int", 1),
		zap.Int64("int64", 2),
		zap.Float64("float", 3.0),
		zap.String("string", "four!"),
		zap.Bool("bool", true),
		zap.Time("time", time.Unix(0, 0)),
		zap.Error(errExample),
		zap.Duration("duration", time.Second),
		zap.Object("user-defined type", _jane),
		zap.String("another string", "done!"),
	}
}

func fakeSugarFields() []interface{} {
	return []interface{}{
		"int", 1,
		"int64", 2,
		"float", 3.0,
		"string", "four!",
		"bool", true,
		"time", time.Unix(0, 0),
		"error", errExample,
		"duration", time.Second,
		"user-defined type", _jane,
		"another string", "done!",
	}
}
