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
	"time"

	"go.uber.org/zap/zapcore"
)

// SamplingConfig sets a sampling strategy for the logger. Each second, the
// first N entries with a given message are logged; afterwards, only a fraction
// of entries are logged. The counts reset each second. This helps to cap the
// global CPU and I/O load that logging puts on your process while preserving a
// representative subset of your logs.
//
// Keep in mind that zap's sampling implementation is optimized for performance
// at the expense of absolute correctness. In particular, there's a small
// chance of hash collisions over-sampling your logs.
type SamplingConfig struct {
	Initial    int `json:"initial",yaml:"initial"`
	Thereafter int `json:"therafter",yaml:"thereafter"`
}

// Config offers a declarative way to construct a logger.
//
// It doesn't do anything that can't be done with New, Options, and the various
// zapcore.WriteSyncer and zapcore.Facility wrappers, but it's a simpler way to
// toggle common options.
type Config struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config. The zero value is
	// InfoLevel.
	Level AtomicLevel `json:"level",yaml:"level"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development",yaml:"development"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number.
	DisableCaller bool `json:"disable_caller",yaml:"disable_caller"`
	// DisableStacktrace completely disables automatic stacktrace capturing.
	DisableStacktrace bool `json:"disable_stacktrace",yaml:"disable_stacktrace"`
	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *SamplingConfig `json:"sampling",yaml:"sampling"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console".
	Encoding string `json:"encoding",yaml:"encoding"`
	// EncoderConfig sets options for the chosen encoder. See
	// zapcore.EncoderConfig for details.
	EncoderConfig zapcore.EncoderConfig `json:"encoder_config",yaml:"encoder_config"`
	// OutputPaths is a list of paths to write logging output to. See Open for
	// details.
	OutputPaths []string `json:"output_paths",yaml:"output_paths"`
	// ErrorOutputPaths is a list of paths to write internal logger errors to.
	// The default is standard error.
	ErrorOutputPaths []string `json:"error_output_paths",yaml:"error_output_paths"`
	// InitialFields is a collection of fields to add to the root logger.
	InitialFields map[string]interface{} `json:"initial_fields",yaml:"initial_fields"`
}

// NewProductionConfig is the recommended production configuration. Logging is
// enabled at InfoLevel and above.
//
// It uses a JSON encoder, writes logs to stdout and internal errors to stderr,
// and enables sampling. Stacktraces are automatically included on logs of
// ErrorLevel and above.
func NewProductionConfig() Config {
	return Config{
		Level:       DynamicLevel(),
		Development: false,
		Sampling: &SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewDevelopmentConfig is a reasonable development configuration. Logging is
// enabled at DebugLevel and above.
//
// It (obviously) enables development mode, uses a console encoder, writes logs
// to stdout and internal errors to stderr, and disables sampling. Stacktraces
// are automatically included on logs of WarnLevel and above.
func NewDevelopmentConfig() Config {
	dyn := DynamicLevel()
	dyn.SetLevel(DebugLevel)

	return Config{
		Level:       dyn,
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Build constructs a logger from the Config and Options.
func (cfg Config) Build(opts ...Option) (*Logger, error) {
	sink, errSink, err := cfg.openSinks()
	if err != nil {
		return nil, err
	}

	enc, err := cfg.buildEncoder()
	if err != nil {
		return nil, err
	}

	return New(
		zapcore.WriterFacility(enc, sink, cfg.Level),
		cfg.buildOptions(errSink)...,
	).WithOptions(opts...), nil
}

func (cfg Config) buildOptions(errSink zapcore.WriteSyncer) []Option {
	opts := []Option{ErrorOutput(errSink)}

	if cfg.Development {
		opts = append(opts, Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, AddCaller())
	}

	stackLevel := ErrorLevel
	if cfg.Development {
		stackLevel = WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, WrapFacility(func(fac zapcore.Facility) zapcore.Facility {
			return zapcore.Sample(fac, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zapcore.Field, 0, len(cfg.InitialFields))
		for k, v := range cfg.InitialFields {
			fs = append(fs, Any(k, v))
		}
		opts = append(opts, Fields(fs...))
	}

	return opts
}

func (cfg Config) openSinks() (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	sink, err := Open(cfg.OutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	errSink, err := Open(cfg.ErrorOutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	return sink, errSink, nil
}

func (cfg Config) buildEncoder() (zapcore.Encoder, error) {
	switch cfg.Encoding {
	case "json":
		return zapcore.NewJSONEncoder(cfg.EncoderConfig), nil
	case "console":
		return zapcore.NewConsoleEncoder(cfg.EncoderConfig), nil
	}
	return nil, fmt.Errorf("unknown encoding %q", cfg.Encoding)
}
