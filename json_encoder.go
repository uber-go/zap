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
	"encoding/json"
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
	// Initial buffer size.
	_initialBufSize = 1024
)

var jsonPool = sync.Pool{
	New: func() interface{} {
		return &jsonEncoder{
			// Pre-allocate a reasonably-sized buffer for each encoder.
			bytes: make([]byte, 0, _initialBufSize),
		}
	},
}

// jsonEncoder is a logging-optimized JSON encoder.
type jsonEncoder struct {
	bytes []byte
}

// newJSONEncoder creates an encoder, re-using one from the pool if possible. The
// returned encoder is initialized and ready for use.
func newJSONEncoder() *jsonEncoder {
	enc := jsonPool.Get().(*jsonEncoder)
	enc.truncate()
	return enc
}

// Free returns the encoder to the pool. Callers should not retain any
// references to the freed object.
func (enc *jsonEncoder) Free() {
	jsonPool.Put(enc)
}

// AddString adds a string key and value to the encoder's fields. Both key and
// value are JSON-escaped.
func (enc *jsonEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.bytes = append(enc.bytes, '"')
	enc.safeAddString(val)
	enc.bytes = append(enc.bytes, '"')
}

// AddBool adds a string key and a boolean value to the encoder's fields. The
// key is JSON-escaped.
func (enc *jsonEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	if val {
		enc.bytes = append(enc.bytes, "true"...)
		return
	}
	enc.bytes = append(enc.bytes, "false"...)
}

// AddInt adds a string key and integer value to the encoder's fields. The key
// is JSON-escaped.
func (enc *jsonEncoder) AddInt(key string, val int) {
	enc.AddInt64(key, int64(val))
}

// AddInt64 adds a string key and int64 value to the encoder's fields. The key
// is JSON-escaped.
func (enc *jsonEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendInt(enc.bytes, val, 10)
}

// AddFloat64 adds a string key and float64 value to the encoder's fields. The
// key is JSON-escaped, and the floating-point value is encoded using
// strconv.FormatFloat's 'g' option (exponential notation for large exponents,
// grade-school notation otherwise).
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
		enc.bytes = strconv.AppendFloat(enc.bytes, val, 'g', -1, 64)
	}
}

// AddMarshaler adds a LogMarshaler to the encoder's fields.
//
// TODO: Encode the error into the message instead of returning.
func (enc *jsonEncoder) AddMarshaler(key string, obj LogMarshaler) error {
	enc.addKey(key)
	enc.bytes = append(enc.bytes, '{')
	err := obj.MarshalLog(enc)
	enc.bytes = append(enc.bytes, '}')
	return err
}

// AddObject uses reflection to add an arbitrary object to the logging context.
func (enc *jsonEncoder) AddObject(key string, obj interface{}) error {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	enc.bytes = append(enc.bytes, marshaled...)
	return nil
}

// Clone copies the current encoder, including any data already encoded.
func (enc *jsonEncoder) Clone() encoder {
	clone := newJSONEncoder()
	clone.bytes = append(clone.bytes, enc.bytes...)
	return clone
}

// AddFields applies the passed fields to this encoder.
func (enc *jsonEncoder) AddFields(fields []Field) {
	addFields(enc, fields)
}

// WriteMessage writes a complete log message to the supplied writer, including
// the encoder's accumulated fields. It doesn't modify or lock the encoder's
// underlying byte slice. It's safe to call from multiple goroutines, but it's
// not safe to call CreateMessage while adding fields.
func (enc *jsonEncoder) WriteMessage(sink io.Writer, lvl string, msg string, ts time.Time) error {
	// Grab an encoder from the pool so that we can re-use the underlying
	// buffer.
	final := newJSONEncoder()
	defer final.Free()

	final.bytes = append(final.bytes, `{"msg":"`...)
	final.safeAddString(msg)
	final.bytes = append(final.bytes, `","level":"`...)
	final.bytes = append(final.bytes, lvl...)
	final.bytes = append(final.bytes, `","ts":`...)
	final.bytes = strconv.AppendInt(final.bytes, ts.UnixNano(), 10)
	final.bytes = append(final.bytes, `,"fields":{`...)
	final.bytes = append(final.bytes, enc.bytes...)
	final.bytes = append(final.bytes, "}}\n"...)

	n, err := sink.Write(final.bytes)
	if err != nil {
		return err
	}
	if n != len(final.bytes) {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, len(final.bytes))
	}
	return nil
}

func (enc *jsonEncoder) truncate() {
	enc.bytes = enc.bytes[:0]
}

func (enc *jsonEncoder) addKey(key string) {
	last := len(enc.bytes) - 1
	// At some point, we'll also want to support arrays.
	if last >= 0 && enc.bytes[last] != '{' {
		enc.bytes = append(enc.bytes, ',')
	}
	enc.bytes = append(enc.bytes, '"')
	enc.safeAddString(key)
	enc.bytes = append(enc.bytes, '"', ':')
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's escaping function, it doesn't attempt to
// protect the user from browser vulnerabilities or JSONP-related problems.
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
