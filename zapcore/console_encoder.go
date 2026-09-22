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

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/internal/bufferpool"
	"go.uber.org/zap/internal/pool"
)

var _sliceEncoderPool = pool.New(func() *sliceArrayEncoder {
	return &sliceArrayEncoder{
		elems: make([]interface{}, 0, 2),
	}
})

// OrderField represents a field that can be included in console encoder output.
type OrderField string

// Console encoder field constants that can be used to configure field order.
const (
	// OrderFieldTime represents the timestamp field
	OrderFieldTime OrderField = "Time"
	// OrderFieldLevel represents the log level field
	OrderFieldLevel OrderField = "Level"
	// OrderFieldName represents the logger name field
	OrderFieldName OrderField = "Name"
	// OrderFieldCallee represents the caller file and line field
	OrderFieldCallee OrderField = "Callee"
	// OrderFieldFunction represents the caller function field
	OrderFieldFunction OrderField = "Function"
	// OrderFieldMessage represents the log message field
	OrderFieldMessage OrderField = "Message"
	// OrderFieldStack represents the stack trace field
	OrderFieldStack OrderField = "Stack"
)

var defaultConsoleOrder = []OrderField{
	OrderFieldTime,
	OrderFieldLevel,
	OrderFieldName,
	OrderFieldCallee,
	OrderFieldFunction,
	OrderFieldMessage,
	OrderFieldStack,
}

func getSliceEncoder() *sliceArrayEncoder {
	return _sliceEncoderPool.Get()
}

func putSliceEncoder(e *sliceArrayEncoder) {
	e.elems = e.elems[:0]
	_sliceEncoderPool.Put(e)
}

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
	if cfg.ConsoleSeparator == "" {
		// Use a default delimiter of '\t' for backwards compatibility
		cfg.ConsoleSeparator = "\t"
	}

	jsonEncoder := newJSONEncoder(cfg, true)

	return consoleEncoder{
		jsonEncoder: jsonEncoder,
	}
}

func (c consoleEncoder) Clone() Encoder {
	return consoleEncoder{
		jsonEncoder: c.jsonEncoder.Clone().(*jsonEncoder),
	}
}

func (c consoleEncoder) EncodeEntry(ent Entry, fields []Field) (*buffer.Buffer, error) {
	line := bufferpool.Get()

	// We don't want the entry's metadata to be quoted and escaped (if it's
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	//
	// If this ever becomes a performance bottleneck, we can implement
	// ArrayEncoder for our plain-text format.
	arr := getSliceEncoder()

	// Process ordered fields
	for _, f := range c.jsonEncoder.ConsoleFieldOrder {
		switch f {
		case OrderFieldTime:
			if c.TimeKey != "" && c.EncodeTime != nil && !ent.Time.IsZero() {
				c.EncodeTime(ent.Time, arr)
			}
		case OrderFieldLevel:
			if c.LevelKey != "" && c.EncodeLevel != nil {
				c.EncodeLevel(ent.Level, arr)
			}
		case OrderFieldName:
			if ent.LoggerName != "" && c.NameKey != "" {
				nameEncoder := c.EncodeName
				if nameEncoder == nil {
					// Fall back to FullNameEncoder for backward compatibility.
					nameEncoder = FullNameEncoder
				}
				nameEncoder(ent.LoggerName, arr)
			}
		case OrderFieldCallee:
			if ent.Caller.Defined {
				if c.CallerKey != "" && c.EncodeCaller != nil {
					c.EncodeCaller(ent.Caller, arr)
				}
			}
		case OrderFieldFunction:
			if ent.Caller.Defined {
				if c.FunctionKey != "" {
					arr.AppendString(ent.Caller.Function)
				}
			}
		case OrderFieldMessage:
			// Add the message itself.
			if c.MessageKey != "" {
				arr.AppendString(ent.Message)
			}

			// Add any structured context.
			contextEncoder := c.jsonEncoder.Clone().(*jsonEncoder)
			defer contextEncoder.buf.Free() // Free the cloned buffer when done.

			// Add fields from the parameter to the cloned encoder.
			for _, f := range fields {
				f.AddTo(contextEncoder)
			}
			// Add fields from the internal buffer too (from With).
			// The clone already includes fields added via With,
			// so no need to explicitly add c.jsonEncoder.buf.Bytes().
			// We just need to ensure namespaces are closed on the clone.
			contextEncoder.closeOpenNamespaces()

			// Check if the cloned encoder has any content.
			if contextEncoder.buf.Len() > 0 {
				// Manually add curly braces because the buffer contains only the inner KVs.
				jsonStr := "{" + contextEncoder.buf.String() + "}"
				arr.AppendString(jsonStr)
			}
		case OrderFieldStack:
			// If there's no stacktrace key, honor that; this allows users to force
			// single-line output.
			if ent.Stack != "" && c.StacktraceKey != "" {
				arr.AppendString("\n" + ent.Stack)
			}
		}
	}

	// Optimized loop to format arr.elems into line buffer
	for i, elem := range arr.elems {
		if i > 0 {
			// Optimized: Avoid adding separator before stack trace if it starts with newline
			needSeparator := true
			if i == len(arr.elems)-1 { // Check if it's the last element
				if str, ok := elem.(string); ok && len(str) > 0 && str[0] == '\n' {
					needSeparator = false
				}
			}
			if needSeparator {
				line.AppendString(c.ConsoleSeparator)
			}
		} // End of if i > 0

		// Optimized writing based on type
		switch e := elem.(type) {
		case string:
			line.AppendString(e)
		case []byte:
			line.Write(e)
		case int:
			line.AppendInt(int64(e))
		case int8:
			line.AppendInt(int64(e))
		case int16:
			line.AppendInt(int64(e))
		case int32:
			line.AppendInt(int64(e))
		case int64:
			line.AppendInt(e)
		case uint:
			line.AppendUint(uint64(e))
		case uint8:
			line.AppendUint(uint64(e))
		case uint16:
			line.AppendUint(uint64(e))
		case uint32:
			line.AppendUint(uint64(e))
		case uint64:
			line.AppendUint(e)
		case float32:
			line.AppendFloat(float64(e), 32)
		case float64:
			line.AppendFloat(e, 64)
		case bool:
			line.AppendBool(e)
		default:
			// Fallback using fmt.Fprint for unexpected or complex types
			// This path might still cause allocations
			_, _ = fmt.Fprint(line, e)
		}
	} // End of loop through arr.elems
	putSliceEncoder(arr)

	line.AppendString(c.LineEnding)
	return line, nil
}

func (c consoleEncoder) writeContext(line *buffer.Buffer, extra []Field) {
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

func (c consoleEncoder) addSeparatorIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.AppendString(c.ConsoleSeparator)
	}
}
