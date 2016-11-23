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
	"io"
	"io/ioutil"
	"sync"
)

var (
	// Discard is a convenience wrapper around ioutil.Discard.
	Discard = AddSync(ioutil.Discard)
	// DiscardOutput is an Option that discards logger output.
	DiscardOutput = Output(Discard)
)

// A WriteFlusher is an io.Writer that can also flush any buffered data.
type WriteFlusher interface {
	io.Writer
	Flush() error
}

// A WriteSyncer is an io.Writer that can also flush any buffered data. Note
// that *os.File (and thus, os.Stderr and os.Stdout) implement WriteSyncer.
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// AddSync converts an io.Writer to a WriteSyncer. It attempts to be
// intelligent: if the concrete type of the io.Writer implements WriteSyncer or
// WriteFlusher, we'll use the existing Sync or Flush methods. If it doesn't,
// we'll add a no-op Sync method.
func AddSync(w io.Writer) WriteSyncer {
	switch w := w.(type) {
	case WriteSyncer:
		return w
	case WriteFlusher:
		return flusherWrapper{w}
	default:
		return writerWrapper{w}
	}
}

type lockedWriteSyncer struct {
	sync.Mutex
	ws WriteSyncer
}

func newLockedWriteSyncer(ws WriteSyncer) WriteSyncer {
	return &lockedWriteSyncer{ws: ws}
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.ws.Write(bs)
	s.Unlock()
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Sync() error {
	return nil
}

type flusherWrapper struct {
	WriteFlusher
}

func (f flusherWrapper) Sync() error {
	return f.Flush()
}

// MultiWriteSyncer creates a WriteSyncer that duplicates its writes
// and sync calls, similarly to to io.MultiWriter.
func MultiWriteSyncer(writeSyncers ...WriteSyncer) WriteSyncer {
	// Copy to protect against https://github.com/golang/go/issues/7809
	ws := make([]WriteSyncer, len(writeSyncers))
	copy(ws, writeSyncers)
	return &teeWriteSyncer{
		writeSyncers: ws,
	}
}

// See https://golang.org/src/io/multi.go
// In the case where not all underlying syncs writer all bytes, we return the smallest number of bytes wtirren
// but still call Write() on all the underlying syncs.
func (t *teeWriteSyncer) Write(p []byte) (int, error) {
	var errs multiError
	nWritten := 0
	for _, w := range t.writeSyncers {
		n, err := w.Write(p)
		if err != nil {
			errs = append(errs, err)
		}
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, errs.asError()
}

func (t *teeWriteSyncer) Sync() error {
	return wrapMutiError(t.writeSyncers...)
}

// Run a series of `f`s, collecting and aggregating errors if presents
func wrapMutiError(fs ...WriteSyncer) error {
	var errs multiError
	for _, f := range fs {
		if err := f.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs.asError()
}

type multiError []error

func (m multiError) asError() error {
	if len(m) > 0 {
		return m
	}
	return nil
}

func (m multiError) Error() string {
	sb := bytes.Buffer{}
	for _, err := range m {
		sb.WriteString(err.Error())
		sb.WriteString(" ")
	}
	return sb.String()
}

type teeWriteSyncer struct {
	writeSyncers []WriteSyncer
}
