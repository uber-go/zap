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
	"time"

	"github.com/uber-go/zap"
)

type Auth struct {
	ExpiresAt time.Time `json:"expires_at"`
	// Since we'll need to send the token to the browser, we include it in the
	// struct's JSON representation.
	Token string `json:"token"`
}

func (a Auth) MarshalLog(kv zap.KeyValue) error {
	kv.AddInt64("expires_at", a.ExpiresAt.UnixNano())
	// We don't want to log sensitive data.
	kv.AddString("token", "---")
	return nil
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Auth Auth   `auth:"auth"`
}

func (u User) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("name", u.Name)
	kv.AddInt("age", u.Age)
	return kv.AddMarshaler("auth", u.Auth)
}

func ExampleMarshaler() {
	jane := User{
		Name: "Jane Doe",
		Age:  42,
		Auth: Auth{
			ExpiresAt: time.Unix(0, 100),
			Token:     "super secret",
		},
	}

	logger := zap.New(zap.NewJSONEncoder(zap.NoTime()))
	logger.Info("Successful login.", zap.Marshaler("user", jane))

	// Output:
	// {"level":"info","msg":"Successful login.","user":{"name":"Jane Doe","age":42,"auth":{"expires_at":100,"token":"---"}}}
}
