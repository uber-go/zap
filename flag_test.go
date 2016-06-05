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
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelFlag(t *testing.T) {
	tests := []struct {
		args      []string
		wantLevel Level
		wantErr   bool
	}{
		{
			args:      nil,
			wantLevel: InfoLevel,
		},
		{
			args:    []string{"--level", "unknown"},
			wantErr: true,
		},
		{
			args:      []string{"--level", "error"},
			wantLevel: ErrorLevel,
		},
	}

	origCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = origCommandLine }()

	for _, tt := range tests {
		flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
		flag.CommandLine.SetOutput(ioutil.Discard)
		level := LevelFlag("level", InfoLevel, "")

		err := flag.CommandLine.Parse(tt.args)
		if tt.wantErr {
			assert.Error(t, err, "Parse(%v) should fail", tt.args)
			continue
		}

		if assert.NoError(t, err, "Parse(%v) shouldn't fail", tt.args) {
			assert.Equal(t, tt.wantLevel, *level, "Level mismatch")
		}
	}
}
