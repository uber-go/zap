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

package zbark_test

import (
	"os"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zbark"

	"github.com/Sirupsen/logrus"
	"github.com/uber-common/bark"
)

func ExampleBarkify() {
	zapLogger := zap.New(zap.NewJSONEncoder(
		zap.NoTime(), // discard timestamps in tests
	))

	// Wrap our structured logger to mimic bark.Logger.
	logger := zbark.Barkify(zapLogger)

	// The wrapped logger has bark's usual fluent API.
	logger.WithField("errors", 0).Infof("%v accepts arbitrary types.", "Bark")

	// Output:
	// {"level":"info","msg":"Bark accepts arbitrary types.","errors":0}
}

func ExampleDebarkify() {
	logrusLogger := logrus.New()
	logrusLogger.Out = os.Stdout
	logrusLogger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: "lies",
	}
	barkLogger := bark.NewLoggerFromLogrus(logrusLogger).WithField("errors", 0)
	logger := zbark.Debarkify(barkLogger, zap.DebugLevel)

	// The wrapped logger has zap's actually fluent API.
	logger.Info("Zap accepts", zap.String("typed", "fields"))

	// Output:
	// {"errors":0,"level":"info","msg":"Zap accepts","time":"lies","typed":"fields"}
}
