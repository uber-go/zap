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
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockStdioSyncIgnoresInvalidHandle(t *testing.T) {
	for _, tt := range []struct {
		name string
		file **os.File
	}{
		{name: "stdout", file: &os.Stdout},
		{name: "stderr", file: &os.Stderr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			original := *tt.file
			t.Cleanup(func() { *tt.file = original })

			reader, writer, err := os.Pipe()
			require.NoError(t, err)
			t.Cleanup(func() { _ = reader.Close() })
			t.Cleanup(func() { _ = writer.Close() })

			*tt.file = writer
			require.NoError(t, syscall.CloseHandle(syscall.Handle(writer.Fd())))
			require.NoError(t, Lock(*tt.file).Sync())
		})
	}
}

func TestLockRegularFileSyncPropagatesInvalidHandle(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "zap-sync")
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })

	require.NoError(t, syscall.CloseHandle(syscall.Handle(file.Fd())))
	require.ErrorIs(t, Lock(file).Sync(), windowsInvalidHandle)
}
