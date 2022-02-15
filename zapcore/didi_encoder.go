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
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/internal/bufferpool"
)

const (
	defaultLoggerSkipLevel = 4
	vendorPath = "/vendor/"
)

type DiDiLogEncoder struct {
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
func NewDiDiLogEncoder(cfg EncoderConfig) Encoder {
	if cfg.ConsoleSeparator == "" {
		// Use a default delimiter of '||' for backwards compatibility
		cfg.ConsoleSeparator = "||"
	}
	return DiDiLogEncoder{newJSONEncoder(cfg, true)}
}

func (c DiDiLogEncoder) Clone() Encoder {
	return DiDiLogEncoder{c.jsonEncoder.Clone().(*jsonEncoder)}
}

func (c DiDiLogEncoder) EncodeEntry(ent Entry, fields []Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	// We don't want the entry's metadata to be quoted and escaped (if it's
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	//
	// If this ever becomes a performance bottleneck, we can implement
	// ArrayEncoder for our plain-text format.
	arr := getSliceEncoder()
	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, arr)
	}
	if c.TimeKey != "" && c.EncodeTime != nil {
		c.EncodeTime(ent.Time, arr)
	}

	for i := range arr.elems {
		line.AppendByte('[')
		fmt.Fprint(line, arr.elems[i])
		line.AppendByte(']')
	}
	putSliceEncoder(arr)

	// Add caller code path
	var pc uintptr
	if p, _, _, ok := runtime.Caller(defaultLoggerSkipLevel); ok {
		pc = p
	}
	var file, name string
	var lineNo int
	f := runtime.FuncForPC(pc)
	file, lineNo = f.FileLine(pc)
	name = f.Name()

	if idx := strings.LastIndex(name, vendorPath); idx >= 0 {
		name = name[idx+len(vendorPath):]
	}
	line.WriteByte('[')
	line.WriteString(path.Base(file))
	line.WriteByte(':')
	line.WriteString(strconv.Itoa(lineNo))
	line.WriteByte('@')
	line.WriteString(name)
	line.WriteByte(']')

	// Add the message itself.
	if c.MessageKey != "" {
		line.AppendString(ent.Message)
	}

	// Add any structured context.
	c.writeContext(line, fields)

	// If there's no stacktrace key, honor that; this allows users to force
	// single-line output.
	if ent.Stack != "" && c.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(ent.Stack)
	}

	line.AppendString(c.LineEnding)
	return line, nil
}

func (c DiDiLogEncoder) writeContext(line *buffer.Buffer, extra []Field) {
	context := c.jsonEncoder.Clone().(*jsonEncoder)
	defer func() {
		// putJSONEncoder assumes the buffer is still used, but we write out the buffer so
		// we can free it.
		context.buf.Free()
		putJSONEncoder(context)
	}()

	addFields(context, extra)
	context.closeOpenNamespaces()
	if context.buf.Len() == 0 {
		return
	}

	c.addSeparatorIfNecessary(line)
	line.AppendByte('{')
	line.Write(context.buf.Bytes())
	line.AppendByte('}')
}

func (c DiDiLogEncoder) addSeparatorIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}
}
