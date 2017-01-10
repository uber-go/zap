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
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func opts(opts ...Option) []Option {
	return opts
}

type stubbedExit struct {
	Status *int
}

func (se *stubbedExit) Unstub() {
	_exit = os.Exit
}

func (se *stubbedExit) AssertNoExit(t testing.TB) {
	assert.Nil(t, se.Status, "Unexpected exit.")
}

func (se *stubbedExit) AssertStatus(t testing.TB, expected int) {
	if assert.NotNil(t, se.Status, "Expected to exit.") {
		assert.Equal(t, expected, *se.Status, "Unexpected exit code.")
	}
}

func stubExit() *stubbedExit {
	stub := &stubbedExit{}
	_exit = func(s int) { stub.Status = &s }
	return stub
}

func withJSONLogger(t testing.TB, enab LevelEnabler, opts []Option, f func(Logger, *testBuffer)) {
	sink := &testBuffer{}
	errSink := &testBuffer{}

	allOpts := make([]Option, 0, 2+len(opts))
	allOpts = append(allOpts, ErrorOutput(errSink))
	allOpts = append(allOpts, opts...)
	logger := New(
		WriterFacility(newJSONEncoder(NoTime()), sink, enab),
		allOpts...)

	f(logger, sink)
	assert.Empty(t, errSink.String(), "Expected error sink to be empty.")
}

func runConcurrently(goroutines, iterations int, wg *sync.WaitGroup, f func()) {
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f()
			}
		}()
	}
}
