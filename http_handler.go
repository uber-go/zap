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
	"net/http"
)

// payload struct defines the format of the request payload received
type payload struct {
	Level string `json:"level"`
}

// NewHTTPHandler takes a logger instance and provides methods to enable
// runtime changes to the logger level via http.handler interface.
func NewHTTPHandler(logger Logger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var currentLevel string

		switch r.Method {
		case "GET":
			currentLevel = logger.Level().String()
		case "PUT":
			decoder := json.NewDecoder(r.Body)
			var p payload

			err := decoder.Decode(&p)
			if err != nil {
				// received data in wrong format
				http.Error(w, "Bad Request", 400)
				return
			}

			var loggerLevel Level
			err = loggerLevel.UnmarshalText([]byte(p.Level))
			if err != nil {
				// unrecognized level provided by the request
				http.Error(w, "Unrecognized Level", 400)
				return
			}

			logger.SetLevel(loggerLevel)
			currentLevel = loggerLevel.String()
		default:
			http.Error(w, "Method Not Allowed", 405)
			return
		}

		res := payload{currentLevel}
		json.NewEncoder(w).Encode(res)
	}

	return http.HandlerFunc(fn)
}
