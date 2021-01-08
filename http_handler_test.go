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

package zap_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicLevelServeHTTP(t *testing.T) {
	tests := map[string]struct {
		Method        string
		ContentType   string
		Body          string
		ExpectedCode  int
		ExpectedLevel zapcore.Level
	}{
		"GET": {
			Method:        http.MethodGet,
			ExpectedCode:  http.StatusOK,
			ExpectedLevel: zap.InfoLevel,
		},
		"PUT JSON": {
			Method:        http.MethodPut,
			ExpectedCode:  http.StatusOK,
			ExpectedLevel: zap.WarnLevel,
			Body:          `{"level":"warn"}`,
		},
		"PUT URL encoded": {
			Method:        http.MethodPut,
			ExpectedCode:  http.StatusOK,
			ExpectedLevel: zap.WarnLevel,
			ContentType:   "application/x-www-form-urlencoded",
			Body:          "level=warn",
		},
		"PUT JSON unrecognized": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			Body:         `{"level":"unrecognized"}`,
		},
		"PUT URL encoded unrecognized": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			ContentType:  "application/x-www-form-urlencoded",
			Body:         "level=unrecognized",
		},
		"PUT JSON malformed": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			Body:         `{"level":"warn`,
		},
		"PUT URL encoded malformed": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			ContentType:  "application/x-www-form-urlencoded",
			Body:         "level",
		},
		"PUT JSON unspecified": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			Body:         `{}`,
		},
		"PUT URL encoded unspecified": {
			Method:       http.MethodPut,
			ExpectedCode: http.StatusBadRequest,
			ContentType:  "application/x-www-form-urlencoded",
			Body:         "",
		},
		"POST JSON": {
			Method:       http.MethodPost,
			ExpectedCode: http.StatusMethodNotAllowed,
			Body:         `{"level":"warn"}`,
		},
		"POST URL": {
			Method:       http.MethodPost,
			ExpectedCode: http.StatusMethodNotAllowed,
			ContentType:  "application/x-www-form-urlencoded",
			Body:         "level=warn",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lvl := zap.NewAtomicLevel()
			lvl.SetLevel(zapcore.InfoLevel)

			ts := httptest.NewServer(lvl)
			defer ts.Close()

			req, err := http.NewRequest(test.Method, ts.URL, strings.NewReader(test.Body))
			require.NoError(t, err, "Error constructing %s request.", req.Method)
			if test.ContentType != "" {
				req.Header.Set("Content-Type", test.ContentType)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "Error making %s request.", req.Method)
			defer res.Body.Close()

			require.Equal(t, test.ExpectedCode, res.StatusCode, "Unexpected status code.")
			if test.ExpectedCode != http.StatusOK {
				// Don't need to test exact error message, but one should be present.
				var pld struct {
					Error string `json:"error"`
				}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&pld), "Decoding response body")
				assert.NotEmpty(t, pld.Error, "Expected an error message")
				return
			}

			var pld struct {
				Level zapcore.Level `json:"level"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&pld), "Decoding response body")
			assert.Equal(t, test.ExpectedLevel, pld.Level, "Unexpected logging level returned")
		})
	}
}
