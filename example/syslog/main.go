package main

import (
	"log/syslog"
	"net"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	enc := zapcore.NewSyslogEncoder(
		zapcore.SyslogEncoderConfig{
			Facility: syslog.LOG_LOCAL0,
		},
	)
	sink, err := zapcore.NewConnSyncer(
		zapcore.ConnSyncerConfig{
			Dial: func() (net.Conn, error) {
				return net.DialUnix("unixgram", nil, &net.UnixAddr{
					Name: "/var/tmp/dsocket",
					Net:  "unixgram",
				})
			},
		},
	)
	if err != nil {
		panic(err)
	}
	core := zapcore.NewCore(enc, sink, zap.InfoLevel)
	logger := zap.New(core)
	logger.Info("what?",
		zap.String("hi", "there"),
	)
}
