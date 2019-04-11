package zapcore_test

import (
	"github.com/stretchr/testify/assert"
	. "go.uber.org/zap/zapcore"
	"testing"
)

func TestSetConsoleElementDelimiter(t *testing.T) {
	SetConsoleElementDelimiter(' ')
	enc := NewConsoleEncoder(humanEncoderConfig())
	enc.AddString("str", "foo")
	enc.AddInt64("int64-1", 1)

	buf, _ := enc.EncodeEntry(Entry{
		Message: "fake",
		Level:   DebugLevel,
	}, nil)

	assert.Equal(t, `0001-01-01T00:00:00.000Z DEBUG fake {"str": "foo", "int64-1": 1}
`, buf.String())
	buf.Free()
}
