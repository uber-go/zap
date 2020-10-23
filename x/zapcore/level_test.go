package zapcore

import (
	"testing"

	goflags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type testFlags struct {
	Level Level `long:"log-level"`
}

func TestUnmarshalFlag(t *testing.T) {
	tests := []struct {
		level     string
		wantLevel zapcore.Level
		wantErr   string
	}{
		{
			level:     "debug",
			wantLevel: zapcore.DebugLevel,
		},
		{
			level:   "not-a-level",
			wantErr: `unrecognized level: "not-a-level"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var unmarshaled Level

			err := unmarshaled.UnmarshalFlag(tt.level)
			if tt.wantErr != "" {
				require.Error(t, err, "error expected")
				assert.EqualError(t, err, tt.wantErr, "unexpected error message")
				return
			}

			assert.NoError(t, err, "no err expected")
			assert.Equal(t, tt.wantLevel, unmarshaled.Level, "unexpected unmarshaled level")
		})
	}
}

func TestFlagsParse(t *testing.T) {
	tests := []struct {
		level     string
		wantLevel zapcore.Level
		wantErr   string
	}{
		{
			level:     "warn",
			wantLevel: zapcore.WarnLevel,
		},
		{
			level:   "fake",
			wantErr: "invalid argument for flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var flags testFlags

			_, err := goflags.ParseArgs(&flags, []string{"--log-level", tt.level})
			if tt.wantErr != "" {
				require.Error(t, err, "error expected")
				assert.Contains(t, err.Error(), tt.wantErr, "unexpected error message")
				return
			}

			assert.NoError(t, err, "no err expected")
			assert.Equal(t, tt.wantLevel, flags.Level.Level, "unexpeted parsed level")
		})
	}
}
