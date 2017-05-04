// Copyright (c) 2017 Uber Technologies, Inc.
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

package zapcore_test

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestJSONEncoder_EncodeEntry(t *testing.T) {
	type subtestCase struct {
		ent    zapcore.Entry
		fields []zapcore.Field
		want   string
	}

	testCases := []struct {
		subtestName string
		cfg         zapcore.EncoderConfig
		cases       []subtestCase
	}{
		{
			subtestName: "no time, level and msg config",
			cfg: func() zapcore.EncoderConfig {
				cfg := zap.NewProductionEncoderConfig()
				cfg.TimeKey = ""
				cfg.LevelKey = ""
				cfg.MessageKey = ""
				cfg.EncodeTime = zapcore.ISO8601TimeEncoder
				return cfg
			}(),
			cases: []subtestCase{
				{
					ent: zapcore.Entry{},
					fields: []zapcore.Field{
						zap.Time("created_at", time.Date(2017, 5, 3, 21, 9, 11, 980000000, time.UTC)),
					},
					want: "{\"created_at\":\"2017-05-03T21:09:11.980Z\"}\n",
				},
			},
		},
	}
	for _, tc := range testCases {
		enc := zapcore.NewJSONEncoder(tc.cfg)
		t.Run(tc.subtestName, func(t *testing.T) {
			for _, st := range tc.cases {
				buf, err := enc.EncodeEntry(st.ent, st.fields)
				if err != nil {
					t.Fatalf("failed to encode entry; ent=%+v, fields=%+v, err=%+v", st.ent, st.fields, err)
				}
				got := buf.String()
				if got != st.want {
					t.Errorf("got=%q, want=%q, ent=%+v, fields=%+v", got, st.want, st.ent, st.fields)
				}
			}
		})
	}
}
