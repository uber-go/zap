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
					// Actually I got the following result on machine whose TZ=JST+9
					// "{\"created_at\":\"2017-05-04T06:09:11.980+9000\"}\n",
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
