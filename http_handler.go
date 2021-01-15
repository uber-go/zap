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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.uber.org/zap/zapcore"
)

// ServeHTTP is a simple JSON endpoint that can report on or change the current
// logging level.
//
// GET
//
// The GET request returns a JSON description of the current logging level like:
//   {"level":"info"}
//
// PUT
//
// The PUT request changes the logging level. It is perfectly safe to change the
// logging level while a program is running. Two content types are supported:
//
//    Content-Type: application/x-www-form-urlencoded
//
// With this content type, the request body is expected to be URL encoded like:
//
//    level=debug
//
// This is the default content type for a curl PUT request. An example curl
// request could look like this:
//
//    curl -X PUT localhost:8080/log/level -d level=debug
//
// For any other content type, the payload is expected to be JSON encoded and
// look like:
//
//   {"level":"info"}
//
// An example curl request could look like this:
//
//    curl -X PUT localhost:8080/log/level -H "Content-Type: application/json" -d '{"level":"debug"}'
//
func (lvl AtomicLevel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type payload struct {
		Level zapcore.Level `json:"level"`
	}

	enc := json.NewEncoder(w)

	switch r.Method {
	case http.MethodGet:
		enc.Encode(payload{Level: lvl.Level()})
	case http.MethodPut:
		requestedLvl, err := decodePutRequest(r.Header.Get("Content-Type"), r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{Error: err.Error()})
			return
		}
		lvl.SetLevel(requestedLvl)
		enc.Encode(payload{Level: lvl.Level()})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		enc.Encode(errorResponse{
			Error: "Only GET and PUT are supported.",
		})
	}
}

// Decodes incoming PUT requests and returns the requested logging level.
func decodePutRequest(contentType string, body io.Reader) (zapcore.Level, error) {
	if contentType == "application/x-www-form-urlencoded" {
		return decodePutURL(body)
	}
	return decodePutJSON(body)
}

func decodePutURL(body io.Reader) (zapcore.Level, error) {
	pld, err := ioutil.ReadAll(body)
	if err != nil {
		return 0, err
	}
	values, err := url.ParseQuery(string(pld))
	if err != nil {
		return 0, err
	}
	lvl := values.Get("level")
	if lvl == "" {
		return 0, fmt.Errorf("must specify logging level")
	}
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(lvl)); err != nil {
		return 0, err
	}
	return l, nil
}

func decodePutJSON(body io.Reader) (zapcore.Level, error) {
	var pld struct {
		Level *zapcore.Level `json:"level"`
	}
	if err := json.NewDecoder(body).Decode(&pld); err != nil {
		return 0, fmt.Errorf("malformed request body: %v", err)
	}
	if pld.Level == nil {
		return 0, fmt.Errorf("must specify logging level")
	}
	return *pld.Level, nil

}
