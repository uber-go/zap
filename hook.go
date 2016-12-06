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

import "errors"

var (
	errHookNilEntry = errors.New("can't call a hook on a nil *Entry")
	// Skip Caller, Logger.log, and the leveled Logger method when using
	// runtime.Caller.
	_callerSkip  = 2
	_callersSkip = 3 // see golang docs for differences between runtime.Caller(s)
)

// A Hook is executed each time the logger writes an Entry. It can modify the
// entry (including adding context to Entry.Fields()), but must not retain
// references to the entry or any of its contents. Returned errors are written to
// the logger's error output.
//
// Hooks implement the Option interface.
type Hook interface {
	hook(*Entry) error
}

// AddCaller configures the Logger to annotate each log with
// a "caller" field with the values as the filename,
// line number and program counter.
// The caller is the line where the leveled logger method is called.
func AddCaller() Option {
	return &addCallerHook{
		callerSkip: _callerSkip,
	}
}

// AddCallers configures the Logger to annotate each log with
// a "callers" field with the the caller's call stack for all messages
// at or above a given level.
// The caller is the line where the leveled logger method is called.
func AddCallers(lvl Level) Option {
	return &addCallersHook{
		callerSkip: _callersSkip,
		lvl:        lvl,
	}
}

// AddStacks configures the Logger to annotate each log with a stack trace
// for all messages at or above a given level. Keep in mind that this is
// (relatively speaking) quite expensive.
func AddStacks(lvl Level) Option {
	return &addStacksHook{
		lvl: lvl,
	}
}

// SetCallerSkip updates the callerSkip value for the existing AddCaller option
func SetCallerSkip(callerSkip int) Option {
	return &setCallerSkip{
		callerSkip: callerSkip,
	}
}

// SetCallersSkip updates the callerSkip and lvl values for the existing AddCallers option
func SetCallersSkip(callerSkip int, lvl Level) Option {
	return &setCallersSkip{
		callerSkip: callerSkip,
		lvl:        lvl,
	}
}

////////////////////
//   Add Caller   //
////////////////////

// Hook adds caller information as a field with "caller" key
func (h *addCallerHook) hook(e *Entry) error {
	if e == nil {
		return errHookNilEntry
	}

	var field Field
	var err error
	if field, err = Caller(h.callerSkip); err != nil {
		return err
	}

	field.AddTo(e.Fields())
	return nil
}

// apply implements the Option interface
func (h *addCallerHook) apply(m *Meta) {
	m.Hooks = append(m.Hooks, h)
}

// addCallerHook implements both the Hook and Option interfaces
type addCallerHook struct {
	callerSkip int
}

////////////////////
//   Add Callers  //
////////////////////

// Hook adds callers information as a field with "callers" key
func (h *addCallersHook) hook(e *Entry) error {
	if e == nil {
		return errHookNilEntry
	}
	if e.Level >= h.lvl {
		Callers(h.callerSkip).AddTo(e.Fields())
	}
	return nil
}

// apply implements the Option interface
func (h *addCallersHook) apply(m *Meta) {
	m.Hooks = append(m.Hooks, h)
}

// addCallersHook implements both the Hook and Option interfaces
type addCallersHook struct {
	callerSkip int
	lvl        Level
}

////////////////////
//   Add Stacks   //
////////////////////

// Hook adds stack trace information as a field with "stacktrace" key
func (h *addStacksHook) hook(e *Entry) error {
	if e == nil {
		return errHookNilEntry
	}
	if e.Level >= h.lvl {
		Stack().AddTo(e.Fields())
	}
	return nil
}

// apply implements the Option interface
func (h *addStacksHook) apply(m *Meta) {
	m.Hooks = append(m.Hooks, h)
}

// addStacksHook implements both the Hook and Option interfaces
type addStacksHook struct {
	lvl Level
}

/////////////////////////
//   Set Caller Skip   //
/////////////////////////

// apply implements the Option interface
func (s *setCallerSkip) apply(m *Meta) {
	for _, hd := range m.Hooks {
		if h, ok := hd.(*addCallerHook); ok {
			h.callerSkip = s.callerSkip
		}
	}
}

// setCallerSkip implements the Option interfaces
type setCallerSkip struct {
	callerSkip int
}

/////////////////////////
//   Set Callers Skip  //
/////////////////////////

// apply implements the Option interface
func (s *setCallersSkip) apply(m *Meta) {
	for _, hd := range m.Hooks {
		if h, ok := hd.(*addCallersHook); ok {
			h.callerSkip = s.callerSkip
		}
	}
}

// setCallersSkip implements the Option interfaces
type setCallersSkip struct {
	callerSkip int
	lvl        Level
}
