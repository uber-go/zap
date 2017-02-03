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
	"bytes"
	"fmt"
	"strconv"

	"go.uber.org/zap/internal/buffers"
)

type consoleEncoder struct {
	*jsonEncoder
}

// NewConsoleEncoder creates an encoder whose output is designed for human -
// rather than machine - consumption. It serializes the core log entry data
// (message, level, timestamp, etc.) in a plain-text format and leaves the
// structured context as JSON.
//
// Note that although the console encoder doesn't use the keys specified in the
// encoder configuration, it will omit any element whose key is set to the empty
// string.
func NewConsoleEncoder(cfg EncoderConfig) Encoder {
	j := NewJSONEncoder(cfg).(*jsonEncoder)
	j.spaced = true
	return consoleEncoder{j}
}

func (c consoleEncoder) Clone() Encoder {
	return consoleEncoder{c.jsonEncoder.Clone().(*jsonEncoder)}
}

func (c consoleEncoder) EncodeEntry(ent Entry, fields []Field) ([]byte, error) {
	line := bytes.NewBuffer(buffers.Get())

	// We don't want the date and level to be quoted and escaped (if they're
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	arr := &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	if c.TimeKey != "" {
		c.EncodeTime(ent.Time, arr)
	}
	if c.LevelKey != "" {
		c.EncodeLevel(ent.Level, arr)
	}
	for i := range arr.elems {
		if i > 0 {
			line.WriteByte('\t')
		}
		fmt.Fprint(line, arr.elems[i])
	}

	// Compose the logger name and caller info into a call site and add it.
	c.writeCallSite(line, ent.LoggerName, ent.Caller)

	// Add the message itself.
	if c.MessageKey != "" {
		if line.Len() > 0 {
			line.WriteByte('\t')
		}
		line.WriteString(ent.Message)
	}

	// Add any structured context.
	c.writeContext(line, fields)

	// If there's no stacktrace key, honor that; this allows users to force
	// single-line output.
	if ent.Stack != "" && c.StacktraceKey != "" {
		line.WriteByte('\n')
		line.WriteString(ent.Stack)
	}

	line.WriteByte('\n')
	return line.Bytes(), nil
}

func (c consoleEncoder) writeCallSite(line *bytes.Buffer, name string, caller EntryCaller) {
	shouldWriteName := name != "" && c.NameKey != ""
	shouldWriteCaller := caller.Defined && c.CallerKey != ""
	if !shouldWriteName && !shouldWriteCaller {
		return
	}
	if line.Len() > 0 {
		line.WriteByte('\t')
	}
	if shouldWriteName {
		line.WriteString(name)
		if shouldWriteCaller {
			line.WriteByte('@')
		}
	}
	if shouldWriteCaller {
		line.WriteString(caller.File)
		line.WriteByte(':')
		line.WriteString(strconv.FormatInt(int64(caller.Line), 10))
	}
}

func (c consoleEncoder) writeContext(line *bytes.Buffer, extra []Field) {
	context := c.jsonEncoder.Clone().(*jsonEncoder)
	defer buffers.Put(context.bytes)

	for i := range extra {
		extra[i].AddTo(context)
	}
	context.closeOpenNamespaces()
	if len(context.bytes) == 0 {
		return
	}

	if line.Len() > 0 {
		line.WriteByte('\t')
	}
	line.WriteByte('{')
	line.Write(context.bytes)
	line.WriteByte('}')
}
