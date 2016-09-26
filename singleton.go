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

import "sync"

var (
	singletonLogger *logger
	initOnce        sync.Once
	configLock      sync.Mutex
)

// Standard returns the standard, singleton logger. Although this behavior is discouraged in many settings
// many projects use it regardless and we'd prefer a reference implementation to each service doing it differently.
func Standard() Logger {
	if singletonLogger == nil {
		// If the singleton logger hasn't been configured via ConfigureStandard, give it some default options
		initOnce.Do(func() {
			ConfigureStandard(NewJSONEncoder())
		})
	}
	return singletonLogger
}

// ConfigureStandard configures the singleton logger
func ConfigureStandard(enc Encoder, options ...Option) Logger {
	configLock.Lock()
	singletonLogger = New(enc, options...).(*logger)
	configLock.Unlock()
	return singletonLogger
}
