package gozap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LocalConfig struct {
	Enabled bool
	LogFile string
}

// local logging option
func WithLocal(cfg LocalConfig) zapcore.Core {
	if !cfg.Enabled {
		return zapcore.NewNopCore()
	}

	path := cfg.LogFile
	if len(path) == 0 {
		path = "./logs/app.log"
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    10, // megabytes
		MaxBackups: 365,
		MaxAge:     365, // days
		LocalTime:  true,
		Compress:   true,
	})

	core := zapcore.NewCore(
		getEncoder(),
		w,
		zap.InfoLevel,
	)

	return core
}

func getEncoder() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(cfg)
}
