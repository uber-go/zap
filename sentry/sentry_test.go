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

package sentry

import (
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"

	raven "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

type fakeClient raven.Client

func TestBadDSN(t *testing.T) {
	_, err := New("123http://whoops")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create sentry client")
}

func TestEmptyDSN(t *testing.T) {
	l, err := New("")
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestWithLevels(t *testing.T) {
	l, err := New("", MinLevel(zap.InfoLevel))
	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.Equal(t, l.minLevel, zap.InfoLevel)
}

func TestExtra(t *testing.T) {
	extra := zwrap.KeyValueMap{
		"requestID":         "123h2eor2039423",
		"someInt":           123,
		"arrayOfManyThings": []int{1, 2, 3},
	}
	l, err := New("", Extra(extra))
	l.Error("error log", zap.String("foo", "bar"))
	assert.NoError(t, err)
	assert.Equal(t, l.extra, extra)
}

func TestWith(t *testing.T) {
	l, err := New("", Extra(map[string]interface{}{
		"someInt": 123,
	}))
	expected := zwrap.KeyValueMap{"someInt": 123, "someFloat": float64(10)}
	l = l.With(zap.Float64("someFloat", float64(10))).(*Logger)
	assert.NoError(t, err)
	assert.Equal(t, l.extra, expected)
}

func TestWithDoesNotMutate(t *testing.T) {
	l1, err := New("", Extra(map[string]interface{}{
		"someInt": 123,
	}))
	assert.NoError(t, err)

	var _ = l1.With(zap.Float64("someFloat", float64(10))).(*Logger)
	assert.Equal(t, zwrap.KeyValueMap{"someInt": 123}, l1.extra)
}

func TestWithTraceDisabled(t *testing.T) {
	_, ps := capturePackets(func(l *Logger) {
		l.Error("some error message", zap.String("foo", "bar"))
		l.Error("another error message")
	}, TraceEnabled(false))

	for _, p := range ps {
		assert.Empty(t, p.Interfaces)
	}
}

func TestTraceCfg(t *testing.T) {
	l, err := New("", TraceCfg(1, 7, []string{"github.com/uber-go/unicorns"}))
	assert.NoError(t, err)
	assert.Equal(t, l.traceSkipFrames, 1)
	assert.Equal(t, l.traceContextLines, 7)
	assert.Equal(t, l.traceAppPrefixes, []string{"github.com/uber-go/unicorns"})
}

func TestLevels(t *testing.T) {
	_, ps := capturePackets(func(l *Logger) {
		l.Log(zap.DebugLevel, "direct call at Debug level")
		l.Info("info")
		l.Warn("warn")
		l.Error("error")
	}, MinLevel(zap.WarnLevel))

	assert.Equal(t, len(ps), 2, "only Warn and Error packets should be collected")
}

func TestMeta(t *testing.T) {
	l, _ := New("")
	assert.Nil(t, l.Check(zap.InfoLevel, "info log"))

	c := l.Check(zap.ErrorLevel, "error message")
	assert.NotNil(t, c)
}

func TestErrorCapture(t *testing.T) {
	_, p := capturePacket(func(l *Logger) {
		l.Error("some error message", zap.String("foo", "bar"))
	})

	assert.Equal(t, p.Message, "some error message")
	assert.Equal(t, p.Extra, map[string]interface{}{"foo": "bar"})

	trace := p.Interfaces[0].(*raven.Stacktrace)
	assert.NotNil(t, trace.Frames)
}

func TestPacketSending(t *testing.T) {
	withTestSentry(t, func(dsn string, ch <-chan *raven.Packet) {
		sl, err := New(dsn)
		defer sl.Close()

		if err != nil {
			panic("Failed to create sentry client")
		}
		sl.Error("my error message", zap.String("mykey1", "myvalue1"))

		p := <-ch

		assert.Equal(t, p.Message, "my error message")
		assert.Equal(t, map[string]interface{}{"mykey1": "myvalue1"}, p.Extra)
	})
}

func capturePacket(f func(l *Logger), options ...Option) (*Logger, *raven.Packet) {
	l, ps := capturePackets(f, options...)
	if len(ps) != 1 {
		panic("Expected to capture a packet, but didn't")
	}
	return l, ps[0]
}

func capturePackets(f func(l *Logger), options ...Option) (*Logger, []*raven.Packet) {
	l, err := New("", options...)
	if err != nil {
		panic("Failed to create the logger")
	}

	c := &memCapturer{}
	l.Capturer = c

	f(l)

	return l, c.packets
}

func withTestSentry(t *testing.T, f func(string, <-chan *raven.Packet)) {
	ch := make(chan *raven.Packet)
	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		contentType := req.Header.Get("Content-Type")
		var bodyReader io.Reader = req.Body
		// underlying client will compress and encode payload above certain size
		if contentType == "application/octet-stream" {
			bodyReader = base64.NewDecoder(base64.StdEncoding, bodyReader)
			bodyReader, _ = zlib.NewReader(bodyReader)
		}

		d := json.NewDecoder(bodyReader)
		p := &raven.Packet{}
		err := d.Decode(p)
		if err != nil {
			ch <- nil
			t.Fatal(err.Error())
		}
		ch <- p
	})
	s := httptest.NewServer(h)
	defer s.Close()

	fragments := strings.SplitN(s.URL, "://", 2)
	dsn := fmt.Sprintf(
		"%s://public:secret@%s/sentry/project-id",
		fragments[0],
		fragments[1],
	)

	f(dsn, ch)
}
