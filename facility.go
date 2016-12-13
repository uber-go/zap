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
	"io"
	"os"
)

// Facility is a destination for log entries. It can have pervasive fields
// added with With().
type Facility interface {
	LevelEnabler

	With(...Field) Facility
	Log(Entry, ...Field) error
	Check(Entry, *CheckedEntry) *CheckedEntry
	Write(Entry, []Field) error
}

// WriterFacility creates a facility that writes logs to an io.Writer. By
// default, if w is nil, os.Stdout is used.
func WriterFacility(enc Encoder, w io.Writer, enab LevelEnabler) Facility {
	if w == nil {
		w = os.Stdout
	}
	return ioFacility{
		LevelEnabler: enab,
		enc:          enc,
		out:          newLockedWriteSyncer(AddSync(w)),
	}
}

type ioFacility struct {
	LevelEnabler
	enc Encoder
	out WriteSyncer
}

func (iof ioFacility) With(fields ...Field) Facility {
	iof.enc = iof.enc.Clone()
	addFields(iof.enc, fields)
	return iof
}

func (iof ioFacility) Log(ent Entry, fields ...Field) error {
	if iof.Enabled(ent.Level) {
		return iof.Write(ent, fields)
	}
	return nil
}

func (iof ioFacility) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if iof.Enabled(ent.Level) {
		ce = ce.AddFacility(ent, iof)
	}
	return ce
}

func (iof ioFacility) Write(ent Entry, fields []Field) error {
	if err := iof.enc.WriteEntry(iof.out, ent, fields); err != nil {
		return err
	}
	if ent.Level > ErrorLevel {
		// Sync on Panic and Fatal, since they may crash the program.
		return iof.out.Sync()
	}
	return nil
}
