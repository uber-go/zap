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
	"os"

	"github.com/uber-go/atomic"
)

// Config is implementation-agnostic configuration for Loggers. Most Logger
// implementations can reduce the required boilerplate by embedding a *Config.
//
// Note that while Level and SetLevel are safe for concurrent use, direct field
// access and modification is not.
type Config struct {
	Development bool
	Encoder     encoder
	Hooks       []hook
	Output      WriteSyncer
	ErrorOutput WriteSyncer

	lvl *atomic.Int32
}

// NewConfig returns a new configuration, initialized with the default values:
// logging at InfoLevel, a JSON encoder, development mode off, and writing to
// standard error and standard out.
func NewConfig() *Config {
	return &Config{
		lvl:         atomic.NewInt32(int32(InfoLevel)),
		Encoder:     newJSONEncoder(),
		Output:      os.Stdout,
		ErrorOutput: os.Stderr,
	}
}

// Level returns the minimum enabled log level. It's safe to call concurrently.
func (c *Config) Level() Level {
	return Level(c.lvl.Load())
}

// SetLevel atomically alters the the logging level for this configuration and
// all its clones.
func (c *Config) SetLevel(lvl Level) {
	c.lvl.Store(int32(lvl))
}

// Clone creates a copy of the configuration. It deep-copies the encoder, but
// not the hooks (since they rarely change).
func (c *Config) Clone() *Config {
	return &Config{
		lvl:         c.lvl,
		Encoder:     c.Encoder.Clone(),
		Development: c.Development,
		Output:      c.Output,
		ErrorOutput: c.ErrorOutput,
		Hooks:       c.Hooks,
	}
}
