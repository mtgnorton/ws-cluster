package logger

import (
	"context"
	"io"
	"os"

	"github.com/getsentry/sentry-go"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/mtgnorton/ws-cluster/config"
	"github.com/mtgnorton/ws-cluster/tools/wssentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(config config.Config) Logger {
	return &ZapLogger{logger: newZapLogger(config)}
}
func (z ZapLogger) Debug(ctx context.Context, args ...interface{}) {
	z.logger.Debug(args...)
}

func (z ZapLogger) Debugf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Debugf(template, args...)
}

func (z ZapLogger) Info(ctx context.Context, args ...interface{}) {
	z.logger.Info(args...)
}

func (z ZapLogger) Infof(ctx context.Context, template string, args ...interface{}) {
	z.logger.Infof(template, args...)
}

func (z ZapLogger) Warn(ctx context.Context, args ...interface{}) {
	z.logger.Warn(args...)
}

func (z ZapLogger) Warnf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Warnf(template, args...)
}

func (z ZapLogger) Error(ctx context.Context, args ...interface{}) {
	z.logger.Error(args...)
}

func (z ZapLogger) Errorf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Errorf(template, args...)
}

func (z ZapLogger) Fatal(ctx context.Context, args ...interface{}) {
	z.logger.Fatal(args...)
}

func (z ZapLogger) Fatalf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Fatalf(template, args...)
}

func (z ZapLogger) DPanic(ctx context.Context, args ...interface{}) {
	z.logger.DPanic(args...)
}

func (z ZapLogger) DPanicf(ctx context.Context, template string, args ...interface{}) {
	z.logger.DPanicf(template, args...)
}

func (z ZapLogger) Panic(ctx context.Context, args ...interface{}) {
	z.logger.Panic(args...)
}

func (z ZapLogger) Panicf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Panicf(template, args...)
}

func newZapLogger(config config.Config) *zap.SugaredLogger {
	normalWriter := normalWriter(config)
	// errorWriter := errorWriter(config)

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

	//cores := []zapcore.Core{
	//	zapcore.NewCore(encoder, errorWriter, zap.ErrorLevel),
	//}
	//if level < zapcore.ErrorLevel {
	//	cores = append(cores, zapcore.NewCore(encoder, normalWriter, level))
	//}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, normalWriter, level),
	}

	tee := zapcore.NewTee(cores...)

	logger := zap.New(tee, zap.AddCaller(), zap.AddCallerSkip(1))
	err := wssentry.DefaultSentryInstance.Init()
	if err != nil {
		panic(err)
	}
	if config.Values().Sentry.DSN != "" {
		return attachSentry(logger, sentry.CurrentHub().Client()).Sugar()
	}
	return logger.Sugar()
}
func encoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func attachSentry(log *zap.Logger, client *sentry.Client) *zap.Logger {
	cfg := zapsentry.Configuration{
		Level:             zapcore.ErrorLevel, //when to send message to sentry
		EnableBreadcrumbs: true,               // enable sending breadcrumbs to Sentry
		BreadcrumbLevel:   zapcore.InfoLevel,  // at what level should we sent breadcrumbs to sentry, this level can't be higher than `Level`
		Tags: map[string]string{
			"component": "system",
		},
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))

	// don't use value if error was returned. Noop core will be replaced to nil soon.
	if err != nil {
		panic(err)
	}

	log = zapsentry.AttachCoreToLogger(core, log)

	// if you have web service, create a new scope somewhere in middleware to have valid breadcrumbs.
	return log.With(zapsentry.NewScope())
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

//func errorWriter(config config.Config) zapcore.WriteSyncer {
//	lc := config.Values().Log
//
//	if lc.Path != "" {
//		// 去除最后的/
//		if lc.Path[len(lc.Path)-1] == '/' {
//			lc.Path = lc.Path[:len(lc.Path)-1]
//		}
//		lc.Path = lc.Path + "/error.log"
//	}
//
//	lumberJackLogger := &lumberjack.Logger{
//		Filename:   lc.Path,
//		MaxSize:    lc.MaxSize,
//		MaxBackups: lc.MaxBackups,
//		MaxAge:     lc.MaxAge,
//		Compress:   lc.Compress,
//	}
//
//	writers := []io.Writer{lumberJackLogger}
//
//	if lc.Print {
//		writers = append(writers, os.Stdout)
//	}
//	return zapcore.AddSync(io.MultiWriter(writers...))
//}