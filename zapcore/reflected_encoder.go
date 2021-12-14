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
	"encoding/json"
	"io"
)

// A ReflectedEncoder handles the serialization of Field which have FieldType = ReflectType and writes them to the underlying data stream
// The underlying data stream is provided by Zap during initialization. See EncoderConfig.NewReflectedEncoder.
type ReflectedEncoder interface {
	// Encode encodes and writes to the underlying data stream.
	Encode(interface{}) error
}

func defaultReflectedEncoder() func(writer io.Writer) ReflectedEncoder {
	return func(writer io.Writer) ReflectedEncoder {
		stdJsonEncoder := json.NewEncoder(writer)
		// For consistency with our custom JSON encoder.
		stdJsonEncoder.SetEscapeHTML(false)
		return &stdReflectedEncoder{
			encoder: stdJsonEncoder,
		}
	}
}

type stdReflectedEncoder struct {
	encoder *json.Encoder
}

func (enc *stdReflectedEncoder) Encode(obj interface{}) error {
	return enc.encoder.Encode(obj)
}
