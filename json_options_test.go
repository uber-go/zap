package zap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageFormatters(t *testing.T) {
	const msg = "foo"

	tests := []struct {
		name      string
		formatter MessageFormatter
		expected  Field
	}{
		{"MessageKey", MessageKey("the-message"), String("the-message", msg)},
		{"Default", defaultMessageF, String("msg", msg)},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(msg), "Unexpected output from MessageFormatter %s.", tt.name)
	}
}

func TestTimeFormatters(t *testing.T) {
	ts := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		formatter TimeFormatter
		expected  Field
	}{
		{"EpochFormatter", EpochFormatter("the-time"), Float64("the-time", 0)},
		{"RFC3339", RFC3339Formatter("ts"), String("ts", "1970-01-01T00:00:00Z")},
		{"NoTime", NoTime(), Skip()},
		{"Default", defaultTimeF, Float64("ts", 0)},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(ts), "Unexpected output from TimeFormatter %s.", tt.name)
	}
}

func TestLevelFormatters(t *testing.T) {
	const lvl = InfoLevel
	tests := []struct {
		name      string
		formatter LevelFormatter
		expected  Field
	}{
		{"LevelString", LevelString("the-level"), String("the-level", "info")},
		{"Default", defaultLevelF, String("level", "info")},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(lvl), "Unexpected output from LevelFormatter %s.", tt.name)
	}
}
