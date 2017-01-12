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
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	// For JSON-escaping; see jsonEncoder.safeAddString below.
	_hex = "0123456789abcdef"
	// Initial buffer size for encoders.
	_initialBufSize = 1024
)

var (
	// errNilSink signals that Encoder.WriteEntry was called with a nil WriteSyncer.
	errNilSink = errors.New("can't write encoded message a nil WriteSyncer")

	// TODO: use the new buffers package instead of pooling encoders.
	jsonPool = sync.Pool{New: func() interface{} {
		return &jsonEncoder{
			// Pre-allocate a reasonably-sized buffer for each encoder.
			bytes: make([]byte, 0, _initialBufSize),
		}
	}}
)

// A JSONConfig sets configuration options for a JSON encoder.
type JSONConfig struct {
	MessageFormatter func(string) Field
	TimeFormatter    func(time.Time) Field
	LevelFormatter   func(Level) Field
}

type jsonEncoder struct {
	*JSONConfig
	bytes []byte
}

// NewJSONEncoder creates a fast, low-allocation JSON encoder. The encoder
// appropriately escapes all field keys and values.
//
// Note that the encoder doesn't deduplicate keys, so it's possible to produce
// a message like
//   {"foo":"bar","foo":"baz"}
// This is permitted by the JSON specification, but not encouraged. Many
// libraries will ignore duplicate key-value pairs (typically keeping the last
// pair) when unmarshaling, but users should attempt to avoid adding duplicate
// keys.
func NewJSONEncoder(cfg JSONConfig) Encoder {
	enc := jsonPool.Get().(*jsonEncoder)
	enc.truncate()
	enc.JSONConfig = &cfg
	return enc
}

func (enc *jsonEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.bytes = append(enc.bytes, '"')
	enc.safeAddString(val)
	enc.bytes = append(enc.bytes, '"')
}

func (enc *jsonEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *jsonEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendInt(enc.bytes, val, 10)
}

func (enc *jsonEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendUint(enc.bytes, val, 10)
}

func (enc *jsonEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	switch {
	case math.IsNaN(val):
		enc.bytes = append(enc.bytes, `"NaN"`...)
	case math.IsInf(val, 1):
		enc.bytes = append(enc.bytes, `"+Inf"`...)
	case math.IsInf(val, -1):
		enc.bytes = append(enc.bytes, `"-Inf"`...)
	default:
		enc.bytes = strconv.AppendFloat(enc.bytes, val, 'f', -1, 64)
	}
}

func (enc *jsonEncoder) AddObject(key string, obj ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *jsonEncoder) AddArray(key string, arr ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *jsonEncoder) AddReflected(key string, obj interface{}) error {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	enc.bytes = append(enc.bytes, marshaled...)
	return nil
}

func (enc *jsonEncoder) AppendArray(arr ArrayMarshaler) error {
	enc.separateElements()
	enc.bytes = append(enc.bytes, '[')
	err := arr.MarshalLogArray(enc)
	enc.bytes = append(enc.bytes, ']')
	return err
}

func (enc *jsonEncoder) AppendObject(obj ObjectMarshaler) error {
	enc.separateElements()
	enc.bytes = append(enc.bytes, '{')
	err := obj.MarshalLogObject(enc)
	enc.bytes = append(enc.bytes, '}')
	return err
}

func (enc *jsonEncoder) AppendBool(val bool) {
	enc.separateElements()
	enc.bytes = strconv.AppendBool(enc.bytes, val)
}

func (enc *jsonEncoder) Clone() Encoder {
	clone := jsonPool.Get().(*jsonEncoder)
	clone.truncate()
	clone.bytes = append(clone.bytes, enc.bytes...)
	clone.JSONConfig = enc.JSONConfig
	return clone
}

func (enc *jsonEncoder) WriteEntry(sink io.Writer, ent Entry, fields []Field) error {
	if sink == nil {
		return errNilSink
	}

	final := jsonPool.Get().(*jsonEncoder)
	final.truncate()
	final.bytes = append(final.bytes, '{')
	enc.LevelFormatter(ent.Level).AddTo(final)
	enc.TimeFormatter(ent.Time).AddTo(final)
	if ent.Caller.Defined {
		// NOTE: we add the field here for parity compromise with text
		// prepending, while not actually mutating the message string.
		final.AddString("caller", ent.Caller.String())
	}
	enc.MessageFormatter(ent.Message).AddTo(final)
	if len(enc.bytes) > 0 {
		if len(final.bytes) > 1 {
			// All the formatters may have been no-ops.
			final.bytes = append(final.bytes, ',')
		}
		final.bytes = append(final.bytes, enc.bytes...)
	}
	addFields(final, fields)
	if ent.Stack != "" {
		final.AddString("stacktrace", ent.Stack)
	}
	final.bytes = append(final.bytes, '}', '\n')

	expectedBytes := len(final.bytes)
	n, err := sink.Write(final.bytes)
	final.free()
	if err != nil {
		return err
	}
	if n != expectedBytes {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, expectedBytes)
	}
	return nil
}

func (enc *jsonEncoder) truncate() {
	enc.bytes = enc.bytes[:0]
}

func (enc *jsonEncoder) free() {
	jsonPool.Put(enc)
}

func (enc *jsonEncoder) addKey(key string) {
	enc.separateElements()
	enc.bytes = append(enc.bytes, '"')
	enc.safeAddString(key)
	enc.bytes = append(enc.bytes, '"', ':')
}

func (enc *jsonEncoder) separateElements() {
	last := len(enc.bytes) - 1
	if last < 0 {
		return
	}
	switch enc.bytes[last] {
	case '{', '[', ':', ',':
		return
	default:
		enc.bytes = append(enc.bytes, ',')
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *jsonEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '\\' && b != '"' {
				enc.bytes = append(enc.bytes, b)
				continue
			}
			switch b {
			case '\\', '"':
				enc.bytes = append(enc.bytes, '\\', b)
			case '\n':
				enc.bytes = append(enc.bytes, '\\', 'n')
			case '\r':
				enc.bytes = append(enc.bytes, '\\', 'r')
			case '\t':
				enc.bytes = append(enc.bytes, '\\', 't')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				enc.bytes = append(enc.bytes, `\u00`...)
				enc.bytes = append(enc.bytes, _hex[b>>4], _hex[b&0xF])
			}
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			enc.bytes = append(enc.bytes, `\ufffd`...)
			i++
			continue
		}
		enc.bytes = append(enc.bytes, s[i:i+size]...)
		i += size
	}
}
