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

package zapcore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeeLogsBoth(t *testing.T) {
	fac1, logs1 := NewObserver(DebugLevel, 10)
	fac2, logs2 := NewObserver(WarnLevel, 10)
	tee := Tee(fac1, fac2)

	debugEntry := Entry{Level: DebugLevel, Message: "log-at-debug"}
	infoEntry := Entry{Level: InfoLevel, Message: "log-at-info"}
	warnEntry := Entry{Level: WarnLevel, Message: "log-at-warn"}
	errorEntry := Entry{Level: ErrorLevel, Message: "log-at-error"}
	for _, ent := range []Entry{debugEntry, infoEntry, warnEntry, errorEntry} {
		if ce := tee.Check(ent, nil); ce != nil {
			ce.Write()
		}
	}

	assert.Equal(t, []ObservedLog{
		{Entry: debugEntry, Context: []Field{}},
		{Entry: infoEntry, Context: []Field{}},
		{Entry: warnEntry, Context: []Field{}},
		{Entry: errorEntry, Context: []Field{}},
	}, logs1.All())

	assert.Equal(t, []ObservedLog{
		{Entry: warnEntry, Context: []Field{}},
		{Entry: errorEntry, Context: []Field{}},
	}, logs2.All())
}
