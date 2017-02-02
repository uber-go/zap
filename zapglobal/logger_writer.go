package zapglobal

import "go.uber.org/zap"

type loggerWriter struct {
	logger zap.Logger
}

func newLoggerWriter(logger zap.Logger) *loggerWriter {
	return &loggerWriter{logger}
}

func (l *loggerWriter) Write(p []byte) (int, error) {
	l.logger.Info(string(p))
	return len(p), nil
}
