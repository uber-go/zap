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
//
// Allows for a variety of implementations of how to send Sentry packets.
// For more performance sensitive systems, it might make sense to batch
// rather than opening up a connection on each send.
type Capturer interface {
	Capture(p *raven.Packet)
}

type memCapturer struct {
	packets []*raven.Packet
}

func (m *memCapturer) Capture(p *raven.Packet) {
	m.packets = append(m.packets, p)
}

// NonBlockingCapturer does not wait for the result of Sentry packet sending.
type NonBlockingCapturer struct {
	*raven.Client
}

// Capture will fire off a packet without checking the error channel.
func (s *NonBlockingCapturer) Capture(p *raven.Packet) {
	s.Client.Capture(p, nil)
}
