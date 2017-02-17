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

package zapcore

import (
	"sync"

	"go.uber.org/zap/internal/multierror"
)

// A Pusher writes encoded entries at a given level.
type Pusher interface {
	Push(Level, []byte) (int, error)
	Syncer
}

type lockedPusher struct {
	sync.Mutex
	p Pusher
}

// Lock wraps a Pusher in a mutex to make it safe for concurrent use.
// In particular, *os.Files must be locked before use.
func Lock(p Pusher) Pusher {
	if _, ok := p.(*lockedPusher); ok {
		// no need to layer on another lock
		return p
	}
	return &lockedPusher{p: p}
}

func (s *lockedPusher) Push(l Level, bs []byte) (int, error) {
	s.Lock()
	n, err := s.p.Push(l, bs)
	s.Unlock()
	return n, err
}

func (s *lockedPusher) Sync() error {
	s.Lock()
	err := s.p.Sync()
	s.Unlock()
	return err
}

type multiPusher []Pusher

// NewMultiPusher creates a Pusher that duplicates its writes
// and sync calls, much like io.MultiWriter.
func NewMultiPusher(mp ...Pusher) Pusher {
	// Copy to protect against https://github.com/golang/go/issues/7809
	return multiPusher(append([]Pusher(nil), mp...))
}

// See https://golang.org/src/io/multi.go
// When not all underlying syncers write the same number of bytes,
// the smallest number is returned even though Write() is called on
// all of them.
func (mp multiPusher) Push(l Level, p []byte) (int, error) {
	var errs multierror.Error
	nWritten := 0
	for _, pusher := range mp {
		n, err := pusher.Push(l, p)
		errs = errs.Append(err)
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, errs.AsError()
}

func (mp multiPusher) Sync() error {
	var errs multierror.Error
	for _, w := range mp {
		errs = errs.Append(w.Sync())
	}
	return errs.AsError()
}
