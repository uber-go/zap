// Copyright (c) 2021 Uber Technologies, Inc.
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
	"bufio"
	"sync"
	"time"

	"go.uber.org/multierr"
)

// A BufferedWriteSyncer is a WriteSyncer that buffers writes in-memory before
// flushing them to a wrapped WriteSyncer after reaching some limit, or at some
// fixed interval--whichever comes first.
//
// BufferedWriteSyncer is safe for concurrent use. You don't need to use
// zapcore.Lock for WriteSyncers with BufferedWriteSyncer.
type BufferedWriteSyncer struct {
	// WriteSyncer is the WriteSyncer wrapped by BufferedWriteSyncer.
	//
	// This field is required.
	WriteSyncer

	// Size specifies the maximum amount of data the writer will buffered
	// before flushing.
	//
	// Defaults to 256 kB if unspecified.
	Size int

	// FlushInterval specifies how often the writer should flush data if
	// there have been no writes.
	//
	// Defaults to 30 seconds if unspecified.
	FlushInterval time.Duration

	// Clock, if specified, provides control of the source of time for the
	// writer.
	//
	// Defaults to the system clock.
	Clock Clock

	// unexported fields for state
	ws          WriteSyncer
	mu          sync.Mutex
	writer      *bufio.Writer
	ticker      *time.Ticker
	stop        chan struct{}
	initialized bool
}

const (
	// _defaultBufferSize specifies the default size used by Buffer.
	_defaultBufferSize = 256 * 1024 // 256 kB

	// _defaultFlushInterval specifies the default flush interval for
	// Buffer.
	_defaultFlushInterval = 30 * time.Second
)

func (s *BufferedWriteSyncer) initialize() {
	s.ws = s.WriteSyncer

	size := s.Size
	if size == 0 {
		size = _defaultBufferSize
	}

	flushInterval := s.FlushInterval
	if flushInterval == 0 {
		flushInterval = _defaultFlushInterval
	}

	if s.Clock == nil {
		s.Clock = DefaultClock
	}
	s.ticker = s.Clock.NewTicker(flushInterval)

	writer := s.WriteSyncer
	if w, ok := writer.(*lockedWriteSyncer); ok {
		writer = w.ws
	} // don't double lock
	s.writer = bufio.NewWriterSize(writer, size)

	s.stop = make(chan struct{})
	s.initialized = true
	go s.flushLoop()
}

// Write writes log data into buffer syncer directly, multiple Write calls will be batched,
// and log data will be flushed to disk when the buffer is full or periodically.
func (s *BufferedWriteSyncer) Write(bs []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.initialize()
	}

	// To avoid partial writes from being flushed, we manually flush the existing buffer if:
	// * The current write doesn't fit into the buffer fully, and
	// * The buffer is not empty (since bufio will not split large writes when the buffer is empty)
	if len(bs) > s.writer.Available() && s.writer.Buffered() > 0 {
		if err := s.writer.Flush(); err != nil {
			return 0, err
		}
	}

	return s.writer.Write(bs)
}

// Sync flushes buffered log data into disk directly.
func (s *BufferedWriteSyncer) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return multierr.Append(s.writer.Flush(), s.ws.Sync())
}

// flushLoop flushes the buffer at the configured interval until Close is
// called.
func (s *BufferedWriteSyncer) flushLoop() {
	for {
		select {
		case <-s.ticker.C:
			// we just simply ignore error here
			// because the underlying bufio writer stores any errors
			// and we return any error from Sync() as part of the close
			_ = s.Sync()
		case <-s.stop:
			return
		}
	}
}

// Close closes the buffer, cleans up background goroutines, and flushes
// remaining, unwritten data. This will not close the underlying WriteSyncer.
func (s *BufferedWriteSyncer) Close() error {
	s.ticker.Stop()
	close(s.stop)
	return s.Sync()
}