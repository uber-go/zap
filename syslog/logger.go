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

package syslog

import "log/syslog"

var (
	// TODO(pedge): verify all severity levels are covered in a test
	severityToLogFunc = map[Severity]func(*syslog.Writer, string) error{
		EmergSeverity:   (*syslog.Writer).Emerg,
		AlertSeverity:   (*syslog.Writer).Alert,
		CritSeverity:    (*syslog.Writer).Crit,
		ErrSeverity:     (*syslog.Writer).Err,
		WarningSeverity: (*syslog.Writer).Warning,
		NoticeSeverity:  (*syslog.Writer).Notice,
		InfoSeverity:    (*syslog.Writer).Info,
		DebugSeverity:   (*syslog.Writer).Debug,
	}
)

type logger struct {
	delegate *syslog.Writer
}

func newLogger(delegate *syslog.Writer) *logger {
	return &logger{delegate}
}

func (l *logger) Write(severity Severity, message string) error {
	return severityToLogFunc[severity](l.delegate, message)
}

func (l *logger) Close() error {
	return l.delegate.Close()
}
