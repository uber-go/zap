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
	"go.uber.org/atomic"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/internal/ztest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewProdReportingConfig() Config {
	reporting := &ReportingSamplingConfig{
		Enabled:    true,
		LoggerName: "t",
		Level:      AtomicLevel{l: atomic.NewInt32(0)},
		Message:    "t",
	}
	config := NewProductionConfig()
	config.Sampling.Reporting = reporting
	return config
}

func TestConfig(t *testing.T) {
	tests := []struct {
		desc             string
		cfg              Config
		expectN          int64
		expectRe         string
		expectSamplingN  int64
		expectSamplingRe string
	}{
		{
			desc:    "production",
			cfg:     NewProductionConfig(),
			expectN: 2 + 100 + 1, // 2 from initial logs, 100 initial sampled logs, 1 from off-by-one in sampler
			expectRe: `{"level":"info","caller":"zap/config_test.go:\d+","msg":"info","k":"v","z":"zz"}` + "\n" +
				`{"level":"warn","caller":"zap/config_test.go:\d+","msg":"warn","k":"v","z":"zz"}` + "\n",
			expectSamplingN:  0,
			expectSamplingRe: ``,
		},
		{
			desc: "production_with_sampling_reporting",
			cfg:  NewProdReportingConfig(),
			// 2 from initial logs, 100 initial sampled logs, 1 from off-by-one in sampler, 1 for tick roll-over
			expectN: 2 + 100 + 1 + 1,
			expectRe: `{"level":"info","caller":"zap/config_test.go:\d+","msg":"info","k":"v","z":"zz"}` + "\n" +
				`{"level":"warn","caller":"zap/config_test.go:\d+","msg":"warn","k":"v","z":"zz"}` + "\n",
			expectSamplingN:  99,
			expectSamplingRe: `{"level":"info","logger":"t","msg":"t","original_level":"info","original_message":"sampling","count":\d+}`,
		},
		{
			desc:    "development",
			cfg:     NewDevelopmentConfig(),
			expectN: 3 + 200, // 3 initial logs, all 200 subsequent logs
			expectRe: "DEBUG\tzap/config_test.go:" + `\d+` + "\tdebug\t" + `{"k": "v", "z": "zz"}` + "\n" +
				"INFO\tzap/config_test.go:" + `\d+` + "\tinfo\t" + `{"k": "v", "z": "zz"}` + "\n" +
				"WARN\tzap/config_test.go:" + `\d+` + "\twarn\t" + `{"k": "v", "z": "zz"}` + "\n" +
				`testing.\w+`,
			expectSamplingN:  0,
			expectSamplingRe: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			temp, err := ioutil.TempFile("", "zap-prod-config-test")
			require.NoError(t, err, "Failed to create temp file.")
			defer os.Remove(temp.Name())

			tt.cfg.OutputPaths = []string{temp.Name()}
			tt.cfg.EncoderConfig.TimeKey = "" // no timestamps in tests
			tt.cfg.InitialFields = map[string]interface{}{"z": "zz", "k": "v"}

			hook, count := makeCountingHook()
			logger, err := tt.cfg.Build(Hooks(hook))
			require.NoError(t, err, "Unexpected error constructing logger.")

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")

			byteContents, err := ioutil.ReadAll(temp)
			require.NoError(t, err, "Couldn't read log contents from temp file.")
			logs := string(byteContents)
			assert.Regexp(t, tt.expectRe, logs, "Unexpected log output.")

			for i := 0; i < 200; i++ {
				logger.Info("sampling")
			}

			// Rolling over tick to produce sampling report.
			if tt.expectSamplingN > 0 {
				ztest.Sleep(time.Second)
				logger.Info("sampling")
				byteContents, err = ioutil.ReadAll(temp)
				require.NoError(t, err, "Couldn't read log contents from temp file.")
				logs = string(byteContents)
				logSlice := strings.Split(logs, "\n")
				samplingRep := logSlice[len(logSlice)-3]
				require.Regexpf(t, tt.expectSamplingRe, samplingRep, "Unexpected sampling report output.")
				re := regexp.MustCompile(`[\d]+`)
				samplingN, err := strconv.Atoi(re.FindString(samplingRep))
				require.NoError(t, err, "Couldn't read number of sampled logs from report.")
				require.EqualValues(t, tt.expectSamplingN, samplingN)
			}

			assert.Equal(t, tt.expectN, count.Load(), "Hook called an unexpected number of times.")
		})
	}
}

func TestConfigWithInvalidPaths(t *testing.T) {
	tests := []struct {
		desc      string
		output    string
		errOutput string
	}{
		{"output directory doesn't exist", "/tmp/not-there/foo.log", "stderr"},
		{"error output directory doesn't exist", "stdout", "/tmp/not-there/foo-errors.log"},
		{"neither output directory exists", "/tmp/not-there/foo.log", "/tmp/not-there/foo-errors.log"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cfg := NewProductionConfig()
			cfg.OutputPaths = []string{tt.output}
			cfg.ErrorOutputPaths = []string{tt.errOutput}
			_, err := cfg.Build()
			assert.Error(t, err, "Expected an error opening a non-existent directory.")
		})
	}
}
