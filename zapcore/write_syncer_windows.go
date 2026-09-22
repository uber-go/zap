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

//go:build windows

package zapcore

import (
	"errors"
	"os"
	"syscall"
)

// windowsInvalidHandle is returned when FlushFileBuffers is called on an
// invalid standard stream handle.
const windowsInvalidHandle syscall.Errno = 6

type windowsStdioWriteSyncer struct {
	WriteSyncer
}

func (s windowsStdioWriteSyncer) Sync() error {
	err := s.WriteSyncer.Sync()
	if errors.Is(err, windowsInvalidHandle) {
		return nil
	}
	return err
}

func lockWriteSyncer(ws WriteSyncer) WriteSyncer {
	file, ok := ws.(*os.File)
	if !ok || (file != os.Stdout && file != os.Stderr) {
		return ws
	}
	return windowsStdioWriteSyncer{WriteSyncer: ws}
}
