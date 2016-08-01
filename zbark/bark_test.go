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

package zbark

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/uber-go/zap"

	"github.com/stretchr/testify/assert"
	"github.com/uber-common/bark"
)

type user struct {
	Name string `json:"name"`
}

type noJSON struct{}

func (n noJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("fail")
}

type stringable string

func (s stringable) String() string {
	return string(s)
}

type loggable string

func (l loggable) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("foo", string(l))
	return nil
}

func newBark() (bark.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := zap.New(zap.NewJSONEncoder(), zap.DebugLevel, zap.Output(zap.AddSync(buf)))
	return Barkify(logger), buf
}

func assertLogged(t testing.TB, levelName string, shouldPanic bool, out *bytes.Buffer, log func()) {
	if shouldPanic {
		assert.Panics(t, log, "Expected panic at level %v", levelName)
	} else {
		log()
	}
	logged := out.String()
	assert.Contains(t, logged, fmt.Sprintf(`"level":"%s"`, levelName), "Expected log output to contain log level.")
	assert.Contains(t, logged, `"msg":"foobar"`, `Expected log output to contain message "foobar".`)
}

func TestLeveledLogging(t *testing.T) {
	b, out := newBark()

	tests := []struct {
		levelName   string
		log         func(...interface{})
		logf        func(string, ...interface{})
		shouldPanic bool
	}{
		{"debug", b.Debug, b.Debugf, false},
		{"info", b.Info, b.Infof, false},
		{"warn", b.Warn, b.Warnf, false},
		{"error", b.Error, b.Errorf, false},
		{"panic", b.Panic, b.Panicf, true},
	}

	for _, tt := range tests {
		out.Reset()
		assertLogged(t, tt.levelName, tt.shouldPanic, out, func() { tt.log("foo", "bar") })
		out.Reset()
		assertLogged(t, tt.levelName, tt.shouldPanic, out, func() { tt.logf("foo%s", "bar") })
	}
}

func TestWithField(t *testing.T) {
	tests := []struct {
		val      interface{}
		expected string
	}{
		// Concrete types.
		{true, "true"},
		{float64(3.14), "3.14"},
		{int(42), "42"},
		{int64(42), "42"},
		{"foo", `"foo"`},
		{time.Unix(0, 0), "0"},
		{time.Nanosecond, "1"},
		// Interfaces.
		{loggable("bar"), `{"foo":"bar"}`}, // zap.Marshaler
		{errors.New("foo"), `"foo"`},       // error
		{stringable("foo"), `"foo"`},       // fmt.Stringer
		{user{"fred"}, `{"name":"fred"}`},  // json.Marshaler
	}

	for _, tt := range tests {
		b, out := newBark()
		b.WithField("thing", tt.val).Debug("")
		assert.Contains(
			t,
			out.String(),
			fmt.Sprintf(`,"thing":%s}`, tt.expected),
			"Unexpected fields output. Expected %+v to serialize as %s.", tt.val, tt.expected,
		)
	}
}

func TestWithFieldSerializationError(t *testing.T) {
	b, out := newBark()
	b.WithField("thing", noJSON{}).Debug("")
	assert.Contains(
		t,
		out.String(),
		`,"thingError":"json: error calling MarshalJSON for type zbark.noJSON: fail"}`,
		"Expected JSON serialization errors to be logged.",
	)
}

func TestWithFields(t *testing.T) {
	b, out := newBark()
	b.WithFields(bark.Fields{
		"foo": "bar",
		"baz": 42,
	}).Debug("")

	output := out.String()
	// Map iteration order is random.
	orderSame := `,"foo":"bar","baz":42}`
	orderReversed := `,"baz":42,"foo":"bar"}`
	if !strings.Contains(output, orderSame) {
		assert.Contains(t, output, orderReversed, "Expected output to contain both fields.")
	}
}

func TestFields(t *testing.T) {
	b, _ := newBark()
	fields := b.WithField("foo", 1).WithFields(bark.Fields{
		"bar": 2,
		"baz": 3,
	}).Fields()

	assert.Equal(t, 3, len(fields), "Expected exactly three fields.")
	assert.Equal(t, 1, fields["foo"], "Unexpected value for field.")
	assert.Equal(t, 2, fields["bar"], "Unexpected value for field.")
	assert.Equal(t, 3, fields["baz"], "Unexpected value for field.")
}
