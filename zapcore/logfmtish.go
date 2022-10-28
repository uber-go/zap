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

package zapcore

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/internal/bufferpool"
)

type logfmtishEncoder struct {
	ctx *buffer.Buffer
	buf *buffer.Buffer

	idx int64 // used only if encoding an array
}

func NewLogfmtishEncoder() Encoder {
	return &logfmtishEncoder{
		ctx: bufferpool.Get(),
		buf: bufferpool.Get(),
	}
}

func (enc *logfmtishEncoder) addKey(key string) {
	if enc.buf.Len() > 0 {
		enc.buf.WriteByte(' ')
	}
	if enc.ctx.Len() > 0 {
		enc.buf.Write(enc.ctx.Bytes())
		enc.buf.AppendByte('.')
	}
	enc.buf.AppendString(key)
	enc.buf.AppendByte('=')
}

func (enc *logfmtishEncoder) AddArray(key string, v ArrayMarshaler) error {
	arrenc := logfmtishEncoder{ctx: bufferpool.Get(), buf: enc.buf}
	if enc.ctx.Len() > 0 {
		arrenc.ctx.Write(enc.ctx.Bytes())
		arrenc.ctx.AppendByte('.')
	}
	arrenc.ctx.AppendString(key)
	return v.MarshalLogArray(&arrenc)
}

func (enc *logfmtishEncoder) AddObject(key string, v ObjectMarshaler) error {
	objenc := logfmtishEncoder{ctx: bufferpool.Get(), buf: enc.buf}
	if enc.ctx.Len() > 0 {
		objenc.ctx.Write(enc.ctx.Bytes())
		objenc.ctx.AppendByte('.')
	}
	objenc.ctx.AppendString(key)
	return v.MarshalLogObject(&objenc)
}

func (enc *logfmtishEncoder) AddBinary(k string, v []byte) {
	enc.AddString(k, base64.StdEncoding.EncodeToString(v))
}

func (enc *logfmtishEncoder) AddByteString(k string, v []byte) {
	enc.addKey(k)
	enc.buf.AppendString(string(v)) // TODO: handle special charactesr
}

func (enc *logfmtishEncoder) AddBool(k string, v bool) {
	enc.addKey(k)
	enc.buf.AppendBool(v)
}

func (enc *logfmtishEncoder) AddComplex128(k string, v complex128) {
	enc.addKey(k)
	enc.appendComplex(v, 64)
}

func (enc *logfmtishEncoder) AddComplex64(k string, v complex64) {
	enc.addKey(k)
	enc.appendComplex(complex128(v), 32)
}

func (enc *logfmtishEncoder) AddDuration(k string, v time.Duration) {
	enc.addKey(k)
	enc.appendInt(v.Milliseconds()) // TODO: customize
}

func (enc *logfmtishEncoder) AddFloat64(k string, v float64) {
	enc.addKey(k)
	enc.appendFloat(v, 64)
}

func (enc *logfmtishEncoder) AddFloat32(k string, v float32) {
	enc.addKey(k)
	enc.appendFloat(float64(v), 32)
}

func (enc *logfmtishEncoder) AddInt64(k string, v int64) {
	enc.addKey(k)
	enc.appendInt(v)
}

func (enc *logfmtishEncoder) AddInt(k string, v int)     { enc.AddInt64(k, int64(v)) }
func (enc *logfmtishEncoder) AddInt32(k string, v int32) { enc.AddInt64(k, int64(v)) }
func (enc *logfmtishEncoder) AddInt16(k string, v int16) { enc.AddInt64(k, int64(v)) }
func (enc *logfmtishEncoder) AddInt8(k string, v int8)   { enc.AddInt64(k, int64(v)) }

func (enc *logfmtishEncoder) AddString(k, v string) {
	enc.addKey(k)
	enc.buf.AppendString(v) // TODO: special characters
}

func (enc *logfmtishEncoder) AddTime(k string, v time.Time) {
	enc.addIndex()
	// TODO: customizable encoder
	enc.appendInt(v.UnixMilli())
}

func (enc *logfmtishEncoder) AddUint64(k string, v uint64) {
	enc.addKey(k)
	enc.buf.AppendUint(v)
}

func (enc *logfmtishEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *logfmtishEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *logfmtishEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *logfmtishEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *logfmtishEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }

func (enc *logfmtishEncoder) AddReflected(k string, v interface{}) error {
	// TODO: implement our own reflected encoder
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	v = nil
	if err := json.Unmarshal(bs, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case bool:
		enc.AddBool(k, v)
	case float64:
		enc.AddFloat64(k, v)
	case string:
		enc.AddString(k, v)
	case []any:
		enc.AddArray(k, anyArray(v))
	case map[string]any:
		enc.AddObject(k, anyObject(v))
	case nil:
		enc.AddString(k, "nil")
	}

	return nil
}

func (enc *logfmtishEncoder) OpenNamespace(k string) {
	if enc.ctx.Len() > 0 {
		enc.ctx.AppendByte('.')
	}
	enc.ctx.AppendString(k)
}

func (enc *logfmtishEncoder) Clone() Encoder {
	clone := logfmtishEncoder{
		ctx: bufferpool.Get(),
		buf: bufferpool.Get(),
		idx: enc.idx,
	}
	clone.ctx.Write(enc.ctx.Bytes())
	clone.buf.Write(enc.buf.Bytes())
	return &clone
}

func (enc *logfmtishEncoder) EncodeEntry(ent Entry, fs []Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	// TODO: time
	line.WriteString(ent.Level.CapitalString())
	line.WriteByte('\t')
	line.WriteString(ent.Message)
	if len(fs) > 0 {
		line.WriteByte('\t')
	}

	enc = enc.Clone().(*logfmtishEncoder)
	for _, f := range fs {
		f.AddTo(enc)
	}

	line.Write(enc.buf.Bytes())
	line.AppendByte('\n')
	return line, nil
}

type anyObject map[string]any

func (ms anyObject) MarshalLogObject(enc ObjectEncoder) error {
	for k, v := range ms {
		switch v := v.(type) {
		case bool:
			enc.AddBool(k, v)
		case float64:
			enc.AddFloat64(k, v)
		case string:
			enc.AddString(k, v)
		case []any:
			return enc.AddArray(k, anyArray(v))
		case map[string]any:
			if err := enc.AddObject(k, anyObject(v)); err != nil {
				return err
			}
		case nil:
			enc.AddString(k, "nil")
		}
	}
	return nil
}

func (enc *logfmtishEncoder) addIndex() {
	if enc.buf.Len() > 0 {
		enc.buf.WriteByte(' ')
	}
	enc.buf.Write(enc.ctx.Bytes())
	enc.buf.AppendByte('[')
	enc.buf.AppendInt(enc.idx)
	enc.buf.AppendString("]=")

	enc.idx++
}

func (enc *logfmtishEncoder) AppendBool(v bool) {
	enc.addIndex()
	enc.buf.AppendBool(v)
}

func (enc *logfmtishEncoder) AppendByteString(v []byte) {
	enc.addIndex()
	enc.buf.AppendString(string(v)) // TOOD: handle special chars
}

func (enc *logfmtishEncoder) appendComplex(val complex128, precision int) {
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, precision)
	// If imaginary part is less than 0, minus (-) sign is added by default
	// by AppendFloat.
	if i >= 0 {
		enc.buf.AppendByte('+')
	}
	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')
}

func (enc *logfmtishEncoder) AppendComplex128(v complex128) {
	enc.addIndex()
	enc.appendComplex(v, 64)
}

func (enc *logfmtishEncoder) AppendComplex64(v complex64) {
	enc.addIndex()
	enc.appendComplex(complex128(v), 32)
}

func (enc *logfmtishEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`NaN`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`+Inf`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`-Inf`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

func (enc *logfmtishEncoder) AppendFloat64(v float64) {
	enc.addIndex()
	enc.appendFloat(v, 64)
}

func (enc *logfmtishEncoder) AppendFloat32(v float32) {
	enc.addIndex()
	enc.appendFloat(float64(v), 32)
}

func (enc *logfmtishEncoder) appendInt(v int64) {
	enc.buf.AppendInt(v)
}

func (enc *logfmtishEncoder) AppendInt64(v int64) {
	enc.addIndex()
	enc.appendInt(v)
}

func (enc *logfmtishEncoder) AppendInt(v int)     { enc.AppendInt64(int64(v)) }
func (enc *logfmtishEncoder) AppendInt32(v int32) { enc.AppendInt64(int64(v)) }
func (enc *logfmtishEncoder) AppendInt16(v int16) { enc.AppendInt64(int64(v)) }
func (enc *logfmtishEncoder) AppendInt8(v int8)   { enc.AppendInt64(int64(v)) }

func (enc *logfmtishEncoder) AppendString(v string) {
	enc.addIndex()
	enc.buf.AppendString(v) // TODO: handle special characters
}

func (enc *logfmtishEncoder) AppendUint64(v uint64) {
	enc.addIndex()
	enc.buf.AppendUint(v)
}

func (enc *logfmtishEncoder) AppendUint(v uint)       { enc.AppendUint64(uint64(v)) }
func (enc *logfmtishEncoder) AppendUint32(v uint32)   { enc.AppendUint64(uint64(v)) }
func (enc *logfmtishEncoder) AppendUint16(v uint16)   { enc.AppendUint64(uint64(v)) }
func (enc *logfmtishEncoder) AppendUint8(v uint8)     { enc.AppendUint64(uint64(v)) }
func (enc *logfmtishEncoder) AppendUintptr(v uintptr) { enc.AppendUint64(uint64(v)) }

func (enc *logfmtishEncoder) AppendDuration(v time.Duration) {
	enc.addIndex()
	// TODO: customizable encoder
	enc.AppendInt64(v.Milliseconds())
}

func (enc *logfmtishEncoder) AppendTime(v time.Time) {
	enc.addIndex()
	// TODO: customizable encoder
	enc.AppendInt64(v.UnixMilli())
}

func (enc *logfmtishEncoder) AppendArray(v ArrayMarshaler) error {
	arrenc := logfmtishEncoder{ctx: bufferpool.Get(), buf: enc.buf}
	arrenc.ctx.Write(enc.ctx.Bytes())
	arrenc.ctx.AppendByte('[')
	arrenc.ctx.AppendInt(enc.idx)
	arrenc.ctx.AppendByte(']')
	return v.MarshalLogArray(&arrenc)
}

func (enc *logfmtishEncoder) AppendObject(v ObjectMarshaler) error {
	objenc := logfmtishEncoder{ctx: bufferpool.Get(), buf: enc.buf}
	objenc.ctx.Write(enc.ctx.Bytes())
	objenc.ctx.AppendByte('[')
	objenc.ctx.AppendInt(enc.idx)
	objenc.ctx.AppendByte(']')
	return v.MarshalLogObject(&objenc)
}

func (enc *logfmtishEncoder) AppendReflected(v interface{}) error {
	// TODO: implement our own reflected encoder
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	v = nil
	if err := json.Unmarshal(bs, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case bool:
		enc.AppendBool(v)
	case float64:
		enc.AppendFloat64(v)
	case string:
		enc.AppendString(v)
	case []any:
		enc.AppendArray(anyArray(v))
	case map[string]any:
		enc.AppendObject(anyObject(v))
	case nil:
		enc.AppendString("nil")
	}

	return nil
}

type anyArray []any

func (vs anyArray) MarshalLogArray(enc ArrayEncoder) error {
	for _, v := range vs {
		switch v := v.(type) {
		case bool:
			enc.AppendBool(v)
		case float64:
			enc.AppendFloat64(v)
		case string:
			enc.AppendString(v)
		case []any:
			return enc.AppendArray(anyArray(v))
		case map[string]any:
			if err := enc.AppendObject(anyObject(v)); err != nil {
				return err
			}
		case nil:
			enc.AppendString("nil")
		}
	}
	return nil
}
