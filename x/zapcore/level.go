package zapcore

import "go.uber.org/zap/zapcore"

type Level struct {
	zapcore.Level
}

func (l *Level) UnmarshalFlag(value string) error {
	return l.UnmarshalText([]byte(value))
}
