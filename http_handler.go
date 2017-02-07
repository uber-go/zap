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
	"net/http"
)

// ServeHTTP supports changing logging level with an HTTP request.
//
// GET requests return a JSON description of the current logging level. PUT
// requests change the logging level and expect a payload like:
//   {"level":"info"}
func (lvl AtomicLevel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type payload struct {
		Level *Level `json:"level"`
	}

	enc := json.NewEncoder(w)
	switch r.Method {

	case "GET":
		current := lvl.Level()
		if err := enc.Encode(payload{Level: &current}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("AtomicLevel.ServeHTTP internal error: %v", err))
		}

	case "PUT":
		var req payload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{
				Error: fmt.Sprintf("Request body must be well-formed JSON: %v", err),
			})
			return
		}
		if req.Level == nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{
				Error: "Must specify a logging level.",
			})
			return
		}
		lvl.SetLevel(*req.Level)
		if err := enc.Encode(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("AtomicLevel.ServeHTTP internal error: %v", err))
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		enc.Encode(errorResponse{
			Error: "Only GET and PUT are supported.",
		})
	}
}
