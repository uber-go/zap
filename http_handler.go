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
	"net/http"
)

// A HTTPHandler takes a logger instance and provides methods to enable
// runtime changes to the logger via http.handler interface.
type HTTPHandler struct {
	logger Logger
}

// NewHTTPHandler constructs a new HTTPHandler.
func NewHTTPHandler(logger Logger) *HTTPHandler {
	return &HTTPHandler{logger}
}

// ChangeLogLevel enables changing the logger's log level at runtime.
// It returns an http.Handler that your application can mount to custom routes.
func (h *HTTPHandler) ChangeLogLevel(lvl Level) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h.logger.SetLevel(lvl)

		var res = "false"
		if h.logger.Enabled(lvl) {
			res = "true"
		}
		w.Write([]byte(res))
	}

	return http.HandlerFunc(fn)
}
