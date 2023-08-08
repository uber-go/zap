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
	"io"

	"fortio.org/log"
)

type shmLogger struct {
	fields []log.KeyVal
}

var fLog = &shmLogger{}

func (l *shmLogger) Infof(msg string, fields ...interface{}) {
	log.Infof(msg, fields...)
}

func (l *shmLogger) InfoS(msg string, fields ...log.KeyVal) {
	log.S(log.Info, msg, fields...)
}

func (l *shmLogger) InfoContext(msg string) {
	log.S(log.Info, msg, l.fields...)
}

func newFortiolog(fields ...log.KeyVal) *shmLogger {
	log.Config = log.DefaultConfig()
	log.Config.LogFileAndLine = false
	log.Config.JSON = true
	log.SetOutput(io.Discard)
	log.Color = false // should already be but just in case
	fLog.fields = fields
	log.SetLogLevel(log.Info)
	return fLog
}

func newDisabledFortiolog(fields ...log.KeyVal) *shmLogger {
	logger := newFortiolog()
	log.SetLogLevel(log.Error)
	logger.fields = fields
	return logger
}

func fakeFortiologFields() []log.KeyVal {
	return []log.KeyVal{
		log.Attr("int", _tenInts[0]),
		log.Attr("ints", _tenInts),
		log.Str("string", _tenStrings[0]),
		log.Attr("strings", _tenStrings),
		log.Attr("time", _tenTimes[0]),
		log.Attr("times", _tenTimes),
		log.Attr("user1", _oneUser),
		log.Attr("user2", _oneUser),
		log.Attr("users", _tenUsers),
		log.Attr("error", errExample),
	}
}
