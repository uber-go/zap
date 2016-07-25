package zap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func stubNow(afterEpoch time.Duration) func() {
	prev := _timeNow
	t := time.Unix(0, int64(afterEpoch))
	_timeNow = func() time.Time { return t }
	return func() { _timeNow = prev }
}

func TestNewEntry(t *testing.T) {
	defer stubNow(0)()
	e := newEntry(DebugLevel, "hello", nil)
	assert.Equal(t, DebugLevel, e.Level, "Unexpected log level.")
	assert.Equal(t, time.Unix(0, 0).UTC(), e.Time, "Unexpected time.")
	assert.Nil(t, e.Fields(), "Unexpected fields.")
}
