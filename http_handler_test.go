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
	"errors"
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
	tests := []struct {
		desc          string
		method        string
		query         string
		contentType   string
		body          string
		expectedCode  int
		expectedLevel zapcore.Level
	}{
		{
			desc:          "GET",
			method:        http.MethodGet,
			expectedCode:  http.StatusOK,
			expectedLevel: zap.InfoLevel,
		},
		{
			desc:          "PUT JSON",
			method:        http.MethodPut,
			expectedCode:  http.StatusOK,
			expectedLevel: zap.WarnLevel,
			body:          `{"level":"warn"}`,
		},
		{
			desc:          "PUT URL encoded",
			method:        http.MethodPut,
			expectedCode:  http.StatusOK,
			expectedLevel: zap.WarnLevel,
			contentType:   "application/x-www-form-urlencoded",
			body:          "level=warn",
		},
		{
			desc:          "PUT query parameters",
			method:        http.MethodPut,
			query:         "?level=warn",
			expectedCode:  http.StatusOK,
			expectedLevel: zap.WarnLevel,
			contentType:   "application/x-www-form-urlencoded",
		},
		{
			desc:          "body takes precedence over query",
			method:        http.MethodPut,
			query:         "?level=info",
			expectedCode:  http.StatusOK,
			expectedLevel: zap.WarnLevel,
			contentType:   "application/x-www-form-urlencoded",
			body:          "level=warn",
		},
		{
			desc:          "JSON ignores query",
			method:        http.MethodPut,
			query:         "?level=info",
			expectedCode:  http.StatusOK,
			expectedLevel: zap.WarnLevel,
			body:          `{"level":"warn"}`,
		},
		{
			desc:         "PUT JSON unrecognized",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			body:         `{"level":"unrecognized"}`,
		},
		{
			desc:         "PUT URL encoded unrecognized",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			contentType:  "application/x-www-form-urlencoded",
			body:         "level=unrecognized",
		},
		{
			desc:         "PUT JSON malformed",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			body:         `{"level":"warn`,
		},
		{
			desc:         "PUT URL encoded malformed",
			method:       http.MethodPut,
			query:        "?level=%",
			expectedCode: http.StatusBadRequest,
			contentType:  "application/x-www-form-urlencoded",
		},
		{
			desc:         "PUT Query parameters malformed",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			contentType:  "application/x-www-form-urlencoded",
			body:         "level=%",
		},
		{
			desc:         "PUT JSON unspecified",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			body:         `{}`,
		},
		{
			desc:         "PUT URL encoded unspecified",
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			contentType:  "application/x-www-form-urlencoded",
			body:         "",
		},
		{
			desc:         "POST JSON",
			method:       http.MethodPost,
			expectedCode: http.StatusMethodNotAllowed,
			body:         `{"level":"warn"}`,
		},
		{
			desc:         "POST URL",
			method:       http.MethodPost,
			expectedCode: http.StatusMethodNotAllowed,
			contentType:  "application/x-www-form-urlencoded",
			body:         "level=warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			lvl := zap.NewAtomicLevel()
			lvl.SetLevel(zapcore.InfoLevel)

			server := httptest.NewServer(lvl)
			defer server.Close()

			req, err := http.NewRequest(tt.method, server.URL+tt.query, strings.NewReader(tt.body))
			require.NoError(t, err, "Error constructing %s request.", req.Method)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "Error making %s request.", req.Method)
			defer func() {
				assert.NoError(t, res.Body.Close(), "Error closing response body.")
			}()

			require.Equal(t, tt.expectedCode, res.StatusCode, "Unexpected status code.")
			if tt.expectedCode != http.StatusOK {
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
			assert.Equal(t, tt.expectedLevel, pld.Level, "Unexpected logging level returned")
		})
	}
}

func TestAtomicLevelServeHTTPBrokenWriter(t *testing.T) {
	t.Parallel()

	lvl := zap.NewAtomicLevel()

	request, err := http.NewRequest(http.MethodGet, "http://localhost:1234/log/level", nil)
	require.NoError(t, err, "Error constructing request.")

	recorder := httptest.NewRecorder()
	lvl.ServeHTTP(&brokenHTTPResponseWriter{
		ResponseWriter: recorder,
	}, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Unexpected status code.")
}

type brokenHTTPResponseWriter struct {
	http.ResponseWriter
}

func (w *brokenHTTPResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("great sadness")
}
