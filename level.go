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
	"errors"
	"fmt"
	"math"
)

var errMarshalNilLevel = errors.New("can't marshal a nil *Level to text")

// A Level is a logging priority. Higher levels are more important.
//
// Note that Level satisfies the Option interface, so any Level can be passed to
// NewJSON to override the default logging priority.
type Level int32

const (
	// Debug logs are typically voluminous, and are usually disabled in
	// production.
	Debug Level = iota - 1
	// Info is the default logging priority.
	Info
	// Warn logs are more important than Info, but don't need individual human review.
	Warn
	// Error logs are high-priority. If an application is running smoothly, it
	// shouldn't generate any error-level logs.
	Error
	// Panic logs a message, then panics.
	Panic
	// Fatal logs a message, then calls os.Exit(1).
	Fatal

	// All logs everything.
	All Level = math.MinInt32
	// None silences logging completely.
	None Level = math.MaxInt32
)

// String returns a lower-case ASCII representation of the log level.
func (l Level) String() string {
	switch l {
	case All:
		return "all"
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Panic:
		return "panic"
	case Fatal:
		return "fatal"
	case None:
		return "none"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// MarshalText satisfies text.Marshaler.
func (l *Level) MarshalText() ([]byte, error) {
	if l == nil {
		return nil, errMarshalNilLevel
	}
	return []byte(l.String()), nil
}

// UnmarshalText satisfies text.Unmarshaler.
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	switch string(text) {
	case "all":
		*l = All
	case "debug":
		*l = Debug
	case "info":
		*l = Info
	case "warn":
		*l = Warn
	case "error":
		*l = Error
	case "panic":
		*l = Panic
	case "fatal":
		*l = Fatal
	case "none":
		*l = None
	default:
		return fmt.Errorf("unrecognized level: %v", string(text))
	}
	return nil
}
