package logger

import (
	"github.com/mtgnorton/ws-cluster/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

func NewZapLogger(config config.Config) *zap.SugaredLogger {
	normalWriter := normalWriter(config)
	errorWriter := errorWriter(config)

	encoder := encoder()

	var level zapcore.Level

	switch config.Values().Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, errorWriter, zap.ErrorLevel),
	}
	if level < zapcore.ErrorLevel {
		cores = append(cores, zapcore.NewCore(encoder, normalWriter, level))
	}

	tee := zapcore.NewTee(cores...)

	return zap.New(tee, zap.AddCaller()).Sugar()

}
func encoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func normalWriter(config config.Config) zapcore.WriteSyncer {
	lc := config.Values().Log

	if lc.Path != "" {
		// 去除最后的/
		if lc.Path[len(lc.Path)-1] == '/' {
			lc.Path = lc.Path[:len(lc.Path)-1]
		}
		lc.Path = lc.Path + "/normal.log"
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   lc.Path,
		MaxSize:    lc.MaxSize,
		MaxBackups: lc.MaxBackups,
		MaxAge:     lc.MaxAge,
		Compress:   lc.Compress,
	}

	writers := []io.Writer{lumberJackLogger}

	if lc.Print {
		writers = append(writers, os.Stdout)
	}

	return zapcore.AddSync(io.MultiWriter(writers...))
}

func errorWriter(config config.Config) zapcore.WriteSyncer {
	lc := config.Values().Log

	if lc.Path != "" {
		// 去除最后的/
		if lc.Path[len(lc.Path)-1] == '/' {
			lc.Path = lc.Path[:len(lc.Path)-1]
		}
		lc.Path = lc.Path + "/error.log"
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   lc.Path,
		MaxSize:    lc.MaxSize,
		MaxBackups: lc.MaxBackups,
		MaxAge:     lc.MaxAge,
		Compress:   lc.Compress,
	}

	writers := []io.Writer{lumberJackLogger}

	if lc.Print {
		writers = append(writers, os.Stdout)
	}
	return zapcore.AddSync(io.MultiWriter(writers...))
}
