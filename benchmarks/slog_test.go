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

package benchmarks

import (
	"io"
	"log/slog"
)

func newSlog(fields ...slog.Attr) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil).WithAttrs(fields))
}

func newDisabledSlog(fields ...slog.Attr) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}).WithAttrs(fields))
}

func fakeSlogFields() []slog.Attr {
	return []slog.Attr{
		slog.Int("int", _tenInts[0]),
		slog.Any("ints", _tenInts),
		slog.String("string", _tenStrings[0]),
		slog.Any("strings", _tenStrings),
		slog.Time("time", _tenTimes[0]),
		slog.Any("times", _tenTimes),
		slog.Any("user1", _oneUser),
		slog.Any("user2", _oneUser),
		slog.Any("users", _tenUsers),
		slog.Any("error", errExample),
	}
}

func fakeSlogArgs() []any {
	return []any{
		"int", _tenInts[0],
		"ints", _tenInts,
		"string", _tenStrings[0],
		"strings", _tenStrings,
		"time", _tenTimes[0],
		"times", _tenTimes,
		"user1", _oneUser,
		"user2", _oneUser,
		"users", _tenUsers,
		"error", errExample,
	}
}
