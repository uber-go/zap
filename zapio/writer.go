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

// Package zapio provides tools for interacting with IO streams through Zap.
package zapio

import (
	"bytes"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Writer is an io.Writer that writes to the provided Zap logger, splitting log
// messages on line boundaries. The Writer will buffer writes in memory until
// it encounters a newline, or the caller calls Sync or Close.
//
// Use the Writer with packages like os/exec where an io.Writer is required,
// and you want to log the output using your existing logger configuration. For
// example,
//
//	writer := &zapio.Writer{Log: logger, Level: zap.DebugLevel}
//	defer writer.Close()
//
//	cmd := exec.CommandContext(ctx, ...)
//	cmd.Stdout = writer
//	cmd.Stderr = writer
//	if err := cmd.Run(); err != nil {
//	    return err
//	}
//
// Writer must be closed when finished to flush buffered data to the logger.
type Writer struct {
	// Log specifies the logger to which the Writer will write messages.
	//
	// The Writer will panic if Log is unspecified.
	Log *zap.Logger

	// Log level for the messages written to the provided logger.
	//
	// If unspecified, defaults to Info.
	Level zapcore.Level

	buff bytes.Buffer
}

var (
	_ zapcore.WriteSyncer = (*Writer)(nil)
	_ io.Closer           = (*Writer)(nil)
)

// Write writes the provided bytes to the underlying logger at the configured
// log level and returns the length of the bytes.
//
// Write will split the input on newlines and post each line as a new log entry
// to the logger.
func (w *Writer) Write(bs []byte) (n int, err error) {
	// Skip all checks if the level isn't enabled.
	if !w.Log.Core().Enabled(w.Level) {
		return len(bs), nil
	}

	n = len(bs)
	for len(bs) > 0 {
		bs = w.writeLine(bs)
	}

	return n, nil
}

// writeLine writes a single line from the input, returning the remaining,
// unconsumed bytes.
func (w *Writer) writeLine(line []byte) (remaining []byte) {
	idx := bytes.IndexByte(line, '\n')
	if idx < 0 {
		// If there are no newlines, buffer the entire string.
		w.buff.Write(line)
		return nil
	}

	// Split on the newline, buffer and flush the left.
	line, remaining = line[:idx], line[idx+1:]

	// Fast path: if we don't have a partial message from a previous write
	// in the buffer, skip the buffer and log directly.
	if w.buff.Len() == 0 {
		w.log(line)
		return
	}

	w.buff.Write(line)

	// Log empty messages in the middle of the stream so that we don't lose
	// information when the user writes "foo\n\nbar".
	w.flush(true /* allowEmpty */)

	return remaining
}

// Close closes the writer, flushing any buffered data in the process.
//
// Always call Close once you're done with the Writer to ensure that it flushes
// all data.
func (w *Writer) Close() error {
	return w.Sync()
}

// Sync flushes buffered data to the logger as a new log entry even if it
// doesn't contain a newline.
func (w *Writer) Sync() error {
	// Don't allow empty messages on explicit Sync calls or on Close
	// because we don't want an extraneous empty message at the end of the
	// stream -- it's common for files to end with a newline.
	w.flush(false /* allowEmpty */)
	return nil
}

// flush flushes the buffered data to the logger, allowing empty messages only
// if the bool is set.
func (w *Writer) flush(allowEmpty bool) {
	if allowEmpty || w.buff.Len() > 0 {
		w.log(w.buff.Bytes())
	}
	w.buff.Reset()
}

func (w *Writer) log(b []byte) {
	if ce := w.Log.Check(w.Level, string(b)); ce != nil {
		ce.Write()
	}
}
