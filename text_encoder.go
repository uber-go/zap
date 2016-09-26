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
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"
)

var textPool = sync.Pool{New: func() interface{} {
	return &textEncoder{
		bytes: make([]byte, 0, _initialBufSize),
	}
}}

type textEncoder struct {
	bytes       []byte
	timeFmt     string
	firstNested bool
}

// NewTextEncoder creates a line-oriented text encoder whose output is optimized
// for human, rather than machine, consumption. By default, the encoder uses
// RFC3339-formatted timestamps.
func NewTextEncoder(options ...TextOption) Encoder {
	enc := textPool.Get().(*textEncoder)
	enc.truncate()
	enc.timeFmt = time.RFC3339
	for _, opt := range options {
		opt.apply(enc)
	}
	return enc
}

func (enc *textEncoder) Free() {
	textPool.Put(enc)
}

func (enc *textEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.bytes = append(enc.bytes, val...)
}

func (enc *textEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.bytes = strconv.AppendBool(enc.bytes, val)
}

func (enc *textEncoder) AddInt(key string, val int) {
	enc.AddInt64(key, int64(val))
}

func (enc *textEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendInt(enc.bytes, val, 10)
}

func (enc *textEncoder) AddUint(key string, val uint) {
	enc.AddUint64(key, uint64(val))
}

func (enc *textEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendUint(enc.bytes, val, 10)
}

func (enc *textEncoder) AddUintptr(key string, val uintptr) {
	enc.addKey(key)
	enc.bytes = append(enc.bytes, "0x"...)
	enc.bytes = strconv.AppendUint(enc.bytes, uint64(val), 16)
}

func (enc *textEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.bytes = strconv.AppendFloat(enc.bytes, val, 'f', -1, 64)
}

func (enc *textEncoder) AddMarshaler(key string, obj LogMarshaler) error {
	enc.addKey(key)
	enc.firstNested = true
	enc.bytes = append(enc.bytes, '{')
	err := obj.MarshalLog(enc)
	enc.bytes = append(enc.bytes, '}')
	enc.firstNested = false
	return err
}

func (enc *textEncoder) AddObject(key string, obj interface{}) error {
	enc.AddString(key, fmt.Sprintf("%+v", obj))
	return nil
}

func (enc *textEncoder) Clone() Encoder {
	clone := textPool.Get().(*textEncoder)
	clone.truncate()
	clone.bytes = append(clone.bytes, enc.bytes...)
	clone.timeFmt = enc.timeFmt
	clone.firstNested = enc.firstNested
	return clone
}

func (enc *textEncoder) WriteEntry(sink io.Writer, msg string, lvl Level, t time.Time) error {
	if sink == nil {
		return errNilSink
	}

	final := textPool.Get().(*textEncoder)
	final.truncate()
	enc.addLevel(final, lvl)
	enc.addTime(final, t)
	enc.addMessage(final, msg)

	if len(enc.bytes) > 0 {
		final.bytes = append(final.bytes, ' ')
		final.bytes = append(final.bytes, enc.bytes...)
	}
	final.bytes = append(final.bytes, '\n')

	expectedBytes := len(final.bytes)
	n, err := sink.Write(final.bytes)
	final.Free()
	if err != nil {
		return err
	}
	if n != expectedBytes {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, expectedBytes)
	}
	return nil
}

func (enc *textEncoder) truncate() {
	enc.bytes = enc.bytes[:0]
}

func (enc *textEncoder) addKey(key string) {
	lastIdx := len(enc.bytes) - 1
	if lastIdx >= 0 && !enc.firstNested {
		enc.bytes = append(enc.bytes, ' ')
	} else {
		enc.firstNested = false
	}
	enc.bytes = append(enc.bytes, key...)
	enc.bytes = append(enc.bytes, '=')
}

func (enc *textEncoder) addLevel(final *textEncoder, lvl Level) {
	final.bytes = append(final.bytes, '[')
	switch lvl {
	case DebugLevel:
		final.bytes = append(final.bytes, 'D')
	case InfoLevel:
		final.bytes = append(final.bytes, 'I')
	case WarnLevel:
		final.bytes = append(final.bytes, 'W')
	case ErrorLevel:
		final.bytes = append(final.bytes, 'E')
	case PanicLevel:
		final.bytes = append(final.bytes, 'P')
	case FatalLevel:
		final.bytes = append(final.bytes, 'F')
	default:
		final.bytes = strconv.AppendInt(final.bytes, int64(lvl), 10)
	}
	final.bytes = append(final.bytes, ']')
}

func (enc *textEncoder) addTime(final *textEncoder, t time.Time) {
	if enc.timeFmt == "" {
		return
	}
	final.bytes = append(final.bytes, ' ')
	final.bytes = t.AppendFormat(final.bytes, enc.timeFmt)
}

func (enc *textEncoder) addMessage(final *textEncoder, msg string) {
	final.bytes = append(final.bytes, ' ')
	final.bytes = append(final.bytes, msg...)
}

// A TextOption is used to set options for a text encoder.
type TextOption interface {
	apply(*textEncoder)
}

type textOptionFunc func(*textEncoder)

func (opt textOptionFunc) apply(enc *textEncoder) {
	opt(enc)
}

// TextTimeFormat sets the format for log timestamps, using the same layout
// strings supported by time.Parse.
func TextTimeFormat(layout string) TextOption {
	return textOptionFunc(func(enc *textEncoder) {
		enc.timeFmt = layout
	})
}

// TextNoTime omits timestamps from the serialized log entries.
func TextNoTime() TextOption {
	return TextTimeFormat("")
}
