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
// WriteSyncer. It also returns any error encountered and a function to close
// any opened files.
//
// Passing no paths returns a no-op WriteSyncer. The special paths "stdout" and
// "stderr" are interpreted as os.Stdout and os.Stderr, respectively.
func Open(paths ...string) (zapcore.WriteSyncer, func(), error) {
	if len(paths) == 0 {
		return zapcore.AddSync(ioutil.Discard), func() {}, nil
	}

	writers, close, err := open(paths)
	if len(writers) == 1 {
		return zapcore.Lock(writers[0]), close, err
	}
	return zapcore.Lock(zapcore.NewMultiWriteSyncer(writers...)), close, err
}

func open(paths []string) ([]zapcore.WriteSyncer, func(), error) {
	var errs multierror.Error
	writers := make([]zapcore.WriteSyncer, 0, len(paths))
	files := make([]*os.File, 0, len(paths))
	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
			// Don't close standard out.
			continue
		case "stderr":
			writers = append(writers, os.Stderr)
			// Don't close standard error.
			continue
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		errs = errs.Append(err)
		if err == nil {
			writers = append(writers, f)
			files = append(files, f)
		}
	}
	close := func() {
		for _, f := range files {
			f.Close()
		}
	}
	return writers, close, errs.AsError()
}
