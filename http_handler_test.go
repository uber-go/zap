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
	"testing"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleHTTPClient(t *testing.T, method string, handler http.Handler, reader io.Reader) (int, []byte) {
	ts := httptest.NewServer(handler)
	defer ts.Close()

	client := &http.Client{}

	req, reqErr := http.NewRequest(method, ts.URL, reader)
	require.NoError(t, reqErr, fmt.Sprintf("Error returning new %s request.", method))

	res, resErr := client.Do(req)
	require.NoError(t, resErr, fmt.Sprintf("Error making %s request.", method))

	resString, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.NoError(t, err, "Error reading request body.")

	return res.StatusCode, resString
}

func TestHTTPHandlerGetLevel(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		handler := NewHTTPHandler(jl)

		statusCode, responseStr := sampleHTTPClient(t, "GET", handler, nil)

		assert.Equal(t, http.StatusOK, statusCode, "Unexpected response status code.")
		assert.Equal(t, fmt.Sprintf("%s", responseStr), fmt.Sprintf("{\"level\":\"%v\"}\n", jl.Level().String()), "Unexpected logger level.")
	})
}

func TestHTTPHandlerPutLevel(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		handler := NewHTTPHandler(jl)

		jsonStr := []byte(`{"level":"warn"}`)
		statusCode, responseStr := sampleHTTPClient(t, "PUT", handler, bytes.NewReader(jsonStr))

		assert.Equal(t, http.StatusOK, statusCode, "Unexpected response status code.")
		assert.Equal(t, fmt.Sprintf("%s", responseStr), fmt.Sprintf("{\"level\":\"%v\"}\n", jl.Level().String()), "Unexpected logger level.")
	})
}

func TestHTTPHandlerUnrecognizedLevel(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		handler := NewHTTPHandler(jl)

		jsonStr := []byte(`{"level":"unrecognized-level"}`)
		statusCode, responseStr := sampleHTTPClient(t, "PUT", handler, bytes.NewReader(jsonStr))

		assert.Equal(t, http.StatusBadRequest, statusCode, "Unexpected response status code.")
		assert.Equal(t, fmt.Sprintf("%s", responseStr), "Unrecognized Level\n", "Unexpected response data.")
	})
}

func TestHTTPHandlerBadRequest(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		handler := NewHTTPHandler(jl)

		statusCode, responseStr := sampleHTTPClient(t, "PUT", handler, nil)

		assert.Equal(t, http.StatusBadRequest, statusCode, "Unexpected response status code.")
		assert.Equal(t, fmt.Sprintf("%s", responseStr), "Bad Request\n", "Unexpected response data.")
	})
}

func TestHTTPHandlerMethodNotAllowed(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		handler := NewHTTPHandler(jl)

		statusCode, responseStr := sampleHTTPClient(t, "POST", handler, nil)

		assert.Equal(t, http.StatusMethodNotAllowed, statusCode, "Unexpected response status code.")
		assert.Equal(t, fmt.Sprintf("%s", responseStr), "Method Not Allowed\n", "Unexpected response data.")
	})
}
