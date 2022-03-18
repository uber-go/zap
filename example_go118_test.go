// Copyright (c) 2022 Uber Technologies, Inc.
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

//go:build go1.18
// +build go1.18

package zap_test

import "go.uber.org/zap"

func ExampleObjects() {
	logger := zap.NewExample()
	defer logger.Sync()

	logger.Debug("opening connections",
		zap.Objects("addrs", []addr{
			{IP: "123.45.67.89", Port: 4040},
			{IP: "127.0.0.1", Port: 4041},
			{IP: "192.168.0.1", Port: 4042},
		}))
	// Output:
	// {"level":"debug","msg":"opening connections","addrs":[{"ip":"123.45.67.89","port":4040},{"ip":"127.0.0.1","port":4041},{"ip":"192.168.0.1","port":4042}]}
}
