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

package zapcore

import "go.uber.org/zap/internal/bufferpool"

// Facility is a destination for log entries. It can have pervasive fields
// added with With().
type Facility interface {
	LevelEnabler

	With([]Field) Facility
	Check(Entry, *CheckedEntry) *CheckedEntry
	Write(Entry, []Field) error
}

type nopFacility struct{}

// NopFacility returns a no-op Facility.
func NopFacility() Facility                                       { return nopFacility{} }
func (nopFacility) Enabled(Level) bool                            { return false }
func (n nopFacility) With([]Field) Facility                       { return n }
func (nopFacility) Check(_ Entry, ce *CheckedEntry) *CheckedEntry { return ce }
func (nopFacility) Write(Entry, []Field) error                    { return nil }

// WriterFacility creates a facility that writes logs to a WriteSyncer. By
// default, if w is nil, os.Stdout is used.
func WriterFacility(enc Encoder, ws WriteSyncer, enab LevelEnabler) Facility {
	return &ioFacility{
		LevelEnabler: enab,
		enc:          enc,
		out:          Lock(ws),
	}
}

type ioFacility struct {
	LevelEnabler
	enc Encoder
	out WriteSyncer
}

func (iof *ioFacility) With(fields []Field) Facility {
	clone := iof.clone()
	addFields(clone.enc, fields)
	return clone
}

func (iof *ioFacility) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if iof.Enabled(ent.Level) {
		return ce.AddFacility(ent, iof)
	}
	return ce
}

func (iof *ioFacility) Write(ent Entry, fields []Field) error {
	buf, err := iof.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	_, err = iof.out.Write(buf.Bytes())
	bufferpool.Put(buf)
	if err != nil {
		return err
	}
	if ent.Level > ErrorLevel {
		// Since we may be crashing the program, sync the output.
		return iof.out.Sync()
	}
	return nil
}

func (iof *ioFacility) clone() *ioFacility {
	return &ioFacility{
		LevelEnabler: iof.LevelEnabler,
		enc:          iof.enc.Clone(),
		out:          iof.out,
	}
}
