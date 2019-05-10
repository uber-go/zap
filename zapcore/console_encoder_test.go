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
package zapcore_test

import (
	"github.com/stretchr/testify/assert"
	. "go.uber.org/zap/zapcore"
	"testing"
)

var (
	testEntry = Entry{
		LoggerName: "main",
		Level:      InfoLevel,
		Message:    `hello`,
		Time:       _epoch,
		Stack:      "fake-stack",
		Caller:     EntryCaller{Defined: true, File: "foo.go", Line: 42},
	}
)

func TestConsoleSeparator(t *testing.T) {

	tests := []struct {
		desc            string
		cfg             EncoderConfig
		expectedConsole string
	}{
		{
			desc:            "space console separator",
			cfg:             encoderTestEncoderConfig(' '),
			expectedConsole: "0 info main foo.go:42 hello\nfake-stack\n",
		},
		{
			desc:            "default console separator",
			cfg:             testEncoderConfig(),
			expectedConsole: "0\tinfo\tmain\tfoo.go:42\thello\nfake-stack\n",
		},
		{
			desc:            "default console separator",
			cfg:             encoderTestEncoderConfig('\t'),
			expectedConsole: "0\tinfo\tmain\tfoo.go:42\thello\nfake-stack\n",
		},
		{
			desc:            "dash console separator",
			cfg:             encoderTestEncoderConfig('-'),
			expectedConsole: "0-info-main-foo.go:42-hello\nfake-stack\n",
		},
	}

	for i, tt := range tests {
		console := NewConsoleEncoder(tt.cfg)
		entry := testEntry
		consoleOut, consoleErr := console.EncodeEntry(entry, nil)
		if assert.NoError(t, consoleErr, "Unexpected error console-encoding entry in case #%d.", i) {
			assert.Equal(
				t,
				tt.expectedConsole,
				consoleOut.String(),
				"Unexpected console output: expected to %v.", tt.desc,
			)
		}
	}
}

func encoderTestEncoderConfig(separator byte) EncoderConfig {
	testEncoder := testEncoderConfig()
	testEncoder.ConsoleSeparator = separator
	return testEncoder
}
