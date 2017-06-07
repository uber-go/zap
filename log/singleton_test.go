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

package log

import (
	"bytes"
	"testing"

	"github.com/uber-go/zap"

	"github.com/stretchr/testify/assert"
)

func TestUseSingleton_WithoutConfig(t *testing.T) {
	l := Standard()
	assert.Equal(t, zap.Level(0), l.Level())
}

func TestConfigureSingleton_AfterUser(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = Standard()

	l := ConfigureStandard(zap.NewTextEncoder(zap.TextNoTime()), zap.Output(zap.AddSync(buf)))
	l.Info("Hi")
	assert.Equal(t, "[I] Hi\n", string(buf.Bytes()))
}
