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
// THE SOFTWARE

package sentry

import raven "github.com/getsentry/raven-go"

// Capturer knows what to do with a Sentry packet.
type Capturer interface {
	Capture(p *raven.Packet) error
	Close()
}

type memCapturer struct {
	packets []*raven.Packet
}

func (m *memCapturer) Close() {}

func (m *memCapturer) Capture(p *raven.Packet) error {
	m.packets = append(m.packets, p)
	return nil
}

type nonBlockingCapturer struct {
	*raven.Client
}

func (nb *nonBlockingCapturer) Close() {
	nb.Client.Close()
}

// Capture will fire off a packet without checking the error channel.
func (nb *nonBlockingCapturer) Capture(p *raven.Packet) error {
	nb.Client.Capture(p, nil)
	return nil
}
