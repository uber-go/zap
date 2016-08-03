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
	"net/http"
)

type httpPayload struct {
	Level *Level `json:"level"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type levelHandler struct {
	logger Logger
}

// NewHTTPHandler returns an HTTP handler that can atomically change the logging
// level at runtime. Keep in mind that changing a logger's level also affects that
// logger's ancestors and descendants.
//
// GET requests return a JSON description of the current logging level. PUT
// requests change the logging level and expect a payload like:
//   {"level":"info"}
func NewHTTPHandler(logger Logger) http.Handler {
	return &levelHandler{logger: logger}
}

func (h *levelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getLevel(w, r)
	case "PUT":
		h.putLevel(w, r)
	default:
		h.error(w, "Only GET and PUT are supported.", http.StatusMethodNotAllowed)
	}
}

func (h *levelHandler) getLevel(w http.ResponseWriter, r *http.Request) {
	current := h.logger.Level()
	json.NewEncoder(w).Encode(httpPayload{Level: &current})
}

func (h *levelHandler) putLevel(w http.ResponseWriter, r *http.Request) {
	var p httpPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		h.error(
			w,
			fmt.Sprintf("Request body must be well-formed JSON: %v", err),
			http.StatusBadRequest,
		)
		return
	}
	if p.Level == nil {
		h.error(w, "Must specify a logging level.", http.StatusBadRequest)
		return
	}
	h.logger.SetLevel(*p.Level)
	json.NewEncoder(w).Encode(p)
}

func (h *levelHandler) error(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
