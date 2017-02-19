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

import (
	"bytes"
	"log"
	"os"
)

var (
	// L is a global Logger. It defaults to a no-op implementation but can be
	// replaced using ReplaceGlobals.
	//
	// Both L and S are unsynchronized, so replacing them while they're in
	// use isn't safe.
	L = NewNop()
	// S is a global SugaredLogger, similar to L. It also defaults to a no-op
	// implementation.
	S = L.Sugar()
)

// ReplaceGlobals replaces the global Logger L and the global SugaredLogger S,
// and returns a function to restore the original values.
//
// Note that replacing the global loggers isn't safe while they're being used;
// in practice, this means that only the owner of the application's main
// function should use this method.
func ReplaceGlobals(logger *Logger) func() {
	prev := *L
	L = logger
	S = logger.Sugar()
	return func() { ReplaceGlobals(&prev) }
}

// RedirectStdLog redirects output from the standard library's "log" package to
// the supplied logger at InfoLevel. Since zap already handles caller
// annotations, timestamps, etc., it automatically disables the standard
// library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stdout.
func RedirectStdLog(l *Logger) func() {
	const (
		stdLogDefaultDepth = 4
		loggerWriterDepth  = 1
	)
	flags := log.Flags()
	prefix := log.Prefix()
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&loggerWriter{l.WithOptions(
		AddCallerSkip(stdLogDefaultDepth + loggerWriterDepth),
	)})
	return func() {
		log.SetFlags(flags)
		log.SetPrefix(prefix)
		log.SetOutput(os.Stderr)
	}
}

type loggerWriter struct {
	logger *Logger
}

func (l *loggerWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	l.logger.Info(string(p))
	return len(p), nil
}
