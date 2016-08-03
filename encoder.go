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
	"io"
	"time"
)

// Encoder is a format-agnostic interface for all log entry marshalers. Since
// log encoders don't need to support the same wide range of use cases as
// general-purpose marshalers, it's possible to make them much faster and
// lower-allocation.
//
// Implementations of the KeyValue interface's methods can, of course, freely
// modify the receiver. However, the Clone and WriteEntry methods will be
// called concurrently and shouldn't modify the receiver.
type Encoder interface {
	KeyValue

	// Copy the encoder, ensuring that adding fields to the copy doesn't affect
	// the original.
	Clone() Encoder
	// Return the encoder to the appropriate sync.Pool. Unpooled encoder
	// implementations can no-op this method.
	Free()
	// Write the supplied message, level, and timestamp to the writer, along with
	// any accumulated context.
	WriteEntry(io.Writer, string, Level, time.Time) error
}
