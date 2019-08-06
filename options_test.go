package zap

import (
	"bytes"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/require"
)

func TestSetLoggerName(t *testing.T) {
	buf := &bytes.Buffer{}
	encoder := zapcore.NewConsoleEncoder(NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := New(core, SetLoggerName("zap"))

	require.Equal(t, logger.name, "zap")
}
