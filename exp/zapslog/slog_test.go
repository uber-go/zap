package zapslog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/slog"
)

func TestAddSource(t *testing.T) {
	r := require.New(t)
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(HandlerOptions{
		AddSource: true,
	}.New(fac))
	sl.Info("msg")

	r.Len(logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	r.Equal("msg", entry.Entry.Message, "Unexpected message")
	r.Regexp(
		`/slog_test.go:\d+$`,
		entry.Caller.String(),
		"Unexpected caller annotation.",
	)
}
