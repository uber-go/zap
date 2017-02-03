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

package zap

import "log"

var (
	// L is the global Logger.
	L = New(nil)
	// S is the global SugaredLogger.
	S = Sugar(L)
)

// SetGlobalLogger sets the global Logger L and the global SugaredLogger S.
func SetGlobalLogger(logger Logger) {
	L = logger
	S = Sugar(L)
}

// RedirectStdLogger redirects logging from the golang "log" package to L.
func RedirectStdLogger() {
	log.SetFlags(0)
	log.SetOutput(&loggerWriter{L})
	log.SetPrefix("")
}

type loggerWriter struct {
	logger Logger
}

func (l *loggerWriter) Write(p []byte) (int, error) {
	l.logger.Info(string(p))
	return len(p), nil
}
