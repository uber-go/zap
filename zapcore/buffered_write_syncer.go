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
)

// A BufferedWriteSyncer is a WriteSyncer that can also flush any buffered data
// with the ability to change the buffer size, flush interval and Clock.
// The default values are; Size 256kb, FlushInterval 30s.
type BufferedWriteSyncer struct {
	WriteSyncer

	Size          int
	FlushInterval time.Duration
	Clock         Clock

	// unexported fields for state
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

func (s *BufferedWriteSyncer) loadConfig() {
	size := s.Size
	if size == 0 {
		size = _defaultBufferSize
	}

	flushInterval := s.FlushInterval
	if flushInterval == 0 {
		flushInterval = _defaultFlushInterval
	}

	if s.Clock != nil {
		s.ticker = s.Clock.NewTicker(flushInterval)
	} else {
		s.ticker = DefaultClock.NewTicker(flushInterval)
	}

	s.writer = bufio.NewWriterSize(s.WriteSyncer, size)
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
		s.loadConfig()
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

	return s.writer.Flush()
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
// remaining, unwritten data.
func (s *BufferedWriteSyncer) Close() error {
	s.ticker.Stop()
	close(s.stop)
	return s.Sync()
}
