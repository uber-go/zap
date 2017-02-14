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
	"io/ioutil"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
)

func newDisabledApexLog() *log.Logger {
	return &log.Logger{
		Handler: json.New(ioutil.Discard),
		Level:   log.ErrorLevel,
	}
}

func newApexLog() *log.Logger {
	return &log.Logger{
		Handler: json.New(ioutil.Discard),
		Level:   log.DebugLevel,
	}
}

func fakeApexFields() log.Fields {
	return log.Fields{
		"int":               1,
		"int64":             int64(1),
		"float":             3.0,
		"string":            "four!",
		"bool":              true,
		"time":              time.Unix(0, 0),
		"error":             errExample.Error(),
		"duration":          time.Second,
		"user-defined type": _jane,
		"another string":    "done!",
	}
}
