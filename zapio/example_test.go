// Copyright (c) 2021 Uber Technologies, Inc.
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

package zapio_test

import (
	"io"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func ExampleWriter() {
	logger := zap.NewExample()
	w := &zapio.Writer{Log: logger}

	io.WriteString(w, "starting up\n")
	io.WriteString(w, "running\n")
	io.WriteString(w, "shutting down\n")

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// {"level":"info","msg":"starting up"}
	// {"level":"info","msg":"running"}
	// {"level":"info","msg":"shutting down"}
}
