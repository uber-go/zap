/*
Package zapglobal allows zap to be a global logger.
*/
package zapglobal

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// DefaultLogger is the default Logger.
	DefaultLogger = zap.New(nil)

	globalLogger = DefaultLogger
	globalHooks  = make([]GlobalHook, 0)
	globalLock   = &sync.Mutex{}
)

// GlobalHook is a function that handles a change in the global Logger instance.
type GlobalHook func(zap.Logger)

// GlobalLogger returns the global Logger instance.
func GlobalLogger() zap.Logger {
	return globalLogger
}

// SetGlobalLogger sets the global Logger instance.
func SetGlobalLogger(logger zap.Logger) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLogger = logger
	for _, globalHook := range globalHooks {
		globalHook(globalLogger)
	}
}

// AddGlobalHook adds a GlobalHook that will be called any time SetGlobalLogger is called, and calls the GlobalHook.
func AddGlobalHook(globalHook GlobalHook) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalHooks = append(globalHooks, globalHook)
	globalHook(globalLogger)
}

// RedirectStdLogger will redirect logs to golang's standard logger to the global Logger instance.
func RedirectStdLogger() {
	AddGlobalHook(
		func(logger zap.Logger) {
			log.SetFlags(0)
			log.SetOutput(newLoggerWriter(logger))
			log.SetPrefix("")
		},
	)
}

// With calls With on the global Logger instance.
func With(fields ...zapcore.Field) zap.Logger { return globalLogger.With(fields...) }

// Check calls Check on the global Logger instance.
func Check(level zapcore.Level, message string) *zapcore.CheckedEntry {
	return globalLogger.Check(level, message)
}

// Debug calls Debug on the global Logger instance.
func Debug(message string, fields ...zapcore.Field) { globalLogger.Debug(message, fields...) }

// Info calls Info on the global Logger instance.
func Info(message string, fields ...zapcore.Field) { globalLogger.Info(message, fields...) }

// Warn calls Warn on the global Logger instance.
func Warn(message string, fields ...zapcore.Field) { globalLogger.Warn(message, fields...) }

// Error calls Error on the global Logger instance.
func Error(message string, fields ...zapcore.Field) { globalLogger.Error(message, fields...) }

// DPanic calls DPanic on the global Logger instance.
func DPanic(message string, fields ...zapcore.Field) { globalLogger.DPanic(message, fields...) }

// Panic calls Panic on the global Logger instance.
func Panic(message string, fields ...zapcore.Field) { globalLogger.Panic(message, fields...) }

// Fatal calls Fatal on the global Logger instance.
func Fatal(message string, fields ...zapcore.Field) { globalLogger.Fatal(message, fields...) }

// Facility calls Facility on the global Logger instance.
func Facility() zapcore.Facility { return globalLogger.Facility() }
