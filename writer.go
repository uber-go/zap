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
	"io/ioutil"
	"os"

	"go.uber.org/zap/internal/multierror"
	"go.uber.org/zap/zapcore"
)

// Open is a high-level wrapper that takes a variadic number of paths, opens or
// creates each of the specified files, and combines them into a locked
// WriteSyncer.
//
// Passing no paths returns a no-op WriteSyncer. The special paths "stdout" and
// "stderr" are interpreted as os.Stdout and os.Stderr, respectively.
func Open(paths ...string) (zapcore.WriteSyncer, error) {
	if len(paths) == 0 {
		return zapcore.AddSync(ioutil.Discard), nil
	}
	writers, err := open(paths)
	return zapcore.Lock(zapcore.MultiWriteSyncer(writers...)), err
}

func open(paths []string) ([]zapcore.WriteSyncer, error) {
	var errs multierror.Error
	writers := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
			continue
		case "stderr":
			writers = append(writers, os.Stderr)
			continue
		}
		f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		errs = errs.Append(err)
		if err == nil {
			writers = append(writers, f)
		}
	}
	return writers, errs.AsError()
}
