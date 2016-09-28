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

import "io"

// Tee creates a WriteSyncer that duplicates its writes and
// sync calls. It is similar to io.MultiWriter
func Tee(writeSyncers ...WriteSyncer) WriteSyncer {
	// Copy to protect against https://github.com/golang/go/issues/7809
	ws := make([]WriteSyncer, len(writeSyncers))
	copy(ws, writeSyncers)
	return &teeWriteSyncer{
		writeSyncers: ws,
	}
}

// See https://golang.org/src/io/multi.go
func (t *teeWriteSyncer) Write(p []byte) (int, error) {
	for _, w := range t.writeSyncers {
		n, err := w.Write(p)
		if err != nil {
			return 0, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

func (t *teeWriteSyncer) Sync() error {
	for _, w := range t.writeSyncers {
		if err := w.Sync(); err != nil {
			return err
		}
	}
	return nil
}

type teeWriteSyncer struct {
	writeSyncers []WriteSyncer
}
