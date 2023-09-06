// Copyright (c) 2023 Uber Technologies, Inc.
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
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestCheckedEntryIllegalReuse(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer

	testCore := zaptest.NewLogger(t).Core()
	ce := testCore.Check(zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "hello",
	}, nil)
	ce.ErrorOutput = zapcore.AddSync(&errOut)

	// The first write should succeed.
	ce.Write(zap.String("k", "v"), zap.Int("n", 42))
	assert.Empty(t, errOut.String(), "Expected no errors on first write.")

	// The second write should fail.
	ce.Write(zap.String("foo", "bar"), zap.Int("x", 1))
	assert.Contains(t, errOut.String(), "Unsafe CheckedEntry re-use near Entry",
		"Expected error logged on second write.")
}
