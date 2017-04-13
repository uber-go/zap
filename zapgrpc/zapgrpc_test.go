package zapgrpc

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"
)

func TestLoggerInfoExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.InfoLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, []LoggerOption{WithDebug()}, zapcore.DebugLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugSuppressed(t *testing.T) {
	checkMessages(t, zapcore.InfoLevel, []LoggerOption{WithDebug()}, zapcore.DebugLevel, nil, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerFatalExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.FatalLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Fatal("hello")
		logger.Fatalf("world")
		logger.Fatalln("foo")
	})
}

func checkMessages(
	t testing.TB,
	levelEnabler zapcore.LevelEnabler,
	loggerOptions []LoggerOption,
	expectedLevel zapcore.Level,
	expectedMessages []string,
	f func(*Logger),
) {
	if expectedLevel == zapcore.FatalLevel {
		expectedLevel = zapcore.WarnLevel
	}
	withLogger(levelEnabler, loggerOptions, func(logger *Logger, observedLogs *observer.ObservedLogs) {
		f(logger)
		logEntries := observedLogs.All()
		require.Equal(t, len(expectedMessages), len(logEntries))
		for i, logEntry := range logEntries {
			require.Equal(t, expectedLevel, logEntry.Level)
			require.Equal(t, expectedMessages[i], logEntry.Message)
		}
	})
}

func withLogger(
	levelEnabler zapcore.LevelEnabler,
	loggerOptions []LoggerOption,
	f func(*Logger, *observer.ObservedLogs),
) {
	core, observedLogs := observer.New(levelEnabler)
	f(NewLogger(zap.New(core), append(loggerOptions, withWarn())...), observedLogs)
}

// withWarn redirects the fatal level to the warn level.
//
// This is used for testing.
func withWarn() LoggerOption {
	return func(logger *Logger) {
		logger.fatalFunc = (*zap.SugaredLogger).Warn
		logger.fatalfFunc = (*zap.SugaredLogger).Warnf
	}
}
