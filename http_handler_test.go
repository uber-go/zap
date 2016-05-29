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

	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

func TestHTTPHandlerChangeLogLevel(t *testing.T) {
	withJSONLogger(t, nil, func(jl *jsonLogger, _ func() []string) {
		assert.Equal(t, All, jl.Level(), "Unexpected initial level.")
		assert.NotEqual(t, Debug, jl.Level(), "Unexpected initial level.")

		h := NewHTTPHandler(jl)

		// test server to handle the http.Handler
		ts := httptest.NewServer(h.ChangeLogLevel(Debug))
		defer ts.Close()

		res, err := http.Get(ts.URL)
		if err != nil {
			jl.Fatal("Error in GET request")
		}

		resString, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			jl.Fatal("Error reading response body")
		}

		// check equality of the response data
		assert.Equal(t, "true", fmt.Sprintf("%s", resString), "Unexpected http.handler response")

		// log level changed to 'Debug' as expected
		assert.Equal(t, Debug, jl.Level(), "Unexpected level after ChangeLogLevel.")
	})
}