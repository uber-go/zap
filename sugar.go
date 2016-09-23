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
	"time"
)

// Sugar is a wrapper around core logger whith less verbose API
type Sugar interface {
	// Check the minimum enabled log level.
	Level() Level
	// Change the level of this logger, as well as all its ancestors and
	// descendants. This makes it easy to change the log level at runtime
	// without restarting your application.
	SetLevel(Level)

	// Create a child logger, and optionally add some context to that logger.
	With(...interface{}) (Sugar, error)

	// Log a message at the given level. Messages include any context that's
	// accumulated on the logger, as well as any fields added at the log site.
	Log(Level, string, ...interface{}) error
	Debug(string, ...interface{}) error
	Info(string, ...interface{}) error
	Warn(string, ...interface{}) error
	Error(string, ...interface{}) error
	Panic(string, ...interface{}) error
	Fatal(string, ...interface{}) error
	// If the logger is in development mode (via the Development option), DFatal
	// logs at the Fatal level. Otherwise, it logs at the Error level.
	DFatal(string, ...interface{}) error
}

type sugar struct {
	core Logger
}

// NewSugar is a constructor for Sugar
func NewSugar(core Logger) Sugar {
	return &sugar{core}
}

func (s *sugar) Level() Level {
	return s.core.Level()
}

func (s *sugar) SetLevel(lvl Level) {
	s.core.SetLevel(lvl)
}

func (s *sugar) With(args ...interface{}) (Sugar, error) {
	fields, err := getSugarFields(args...)
	if err != nil {
		return nil, err
	}
	return NewSugar(s.core.With(fields...)), nil
}

func getSugarFields(args ...interface{}) ([]Field, error) {
	var (
		noErrArgs []interface{}
		fields    []Field

		ii    int
		key   string
		value interface{}
	)

	if len(args) == 0 {
		return fields, nil
	}

	switch args[0].(type) {
	case error:
		fields = append(fields, Error(args[0].(error)))
		noErrArgs = args[1:]
	default:
		noErrArgs = args
	}

	if (len(noErrArgs) % 2) != 0 {
		return nil, errors.New("invalid number of arguments")
	}

	for ii, value = range noErrArgs {
		if (ii % 2) == 0 {
			switch value.(type) {
			case string:
				key = value.(string)
			default:
				return nil, errors.New("field name must be string")
			}
		} else {
			switch value.(type) {
			case bool:
				fields = append(fields, Bool(key, value.(bool)))
			case float64:
				fields = append(fields, Float64(key, value.(float64)))
			case int:
				fields = append(fields, Int(key, value.(int)))
			case int64:
				fields = append(fields, Int64(key, value.(int64)))
			case uint:
				fields = append(fields, Uint(key, value.(uint)))
			case uint64:
				fields = append(fields, Uint64(key, value.(uint64)))
			case string:
				fields = append(fields, String(key, value.(string)))
			case time.Time:
				fields = append(fields, Time(key, value.(time.Time)))
			case time.Duration:
				fields = append(fields, Duration(key, value.(time.Duration)))
			case fmt.Stringer:
				fields = append(fields, Stringer(key, value.(fmt.Stringer)))
			case error:
				return nil, errors.New("error can only be the first argument")
			default:
				return nil, errors.New("invalid argument type")
			}
		}
	}
	return fields, nil
}

// Log ...
func (s *sugar) Log(lvl Level, msg string, args ...interface{}) error {
	fields, err := getSugarFields(args...)
	if err != nil {
		return err
	}
	s.core.Log(lvl, msg, fields...)
	return nil
}

func (s *sugar) Debug(msg string, args ...interface{}) error {
	return s.Log(DebugLevel, msg, args...)
}

func (s *sugar) Info(msg string, args ...interface{}) error {
	return s.Log(InfoLevel, msg, args...)
}

func (s *sugar) Warn(msg string, args ...interface{}) error {
	return s.Log(WarnLevel, msg, args...)
}

func (s *sugar) Error(msg string, args ...interface{}) error {
	return s.Log(ErrorLevel, msg, args...)
}

func (s *sugar) Panic(msg string, args ...interface{}) error {
	return s.Log(PanicLevel, msg, args...)
}

func (s *sugar) Fatal(msg string, args ...interface{}) error {
	return s.Log(FatalLevel, msg, args...)
}

func (s *sugar) DFatal(msg string, args ...interface{}) error {
	fields, err := getSugarFields(args...)
	if err != nil {
		return err
	}
	s.core.DFatal(msg, fields...)
	return nil
}
