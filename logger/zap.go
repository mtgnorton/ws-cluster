package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/getsentry/sentry-go"

	"ws-cluster/config"
	"ws-cluster/shared"
	"ws-cluster/tools/wssentry"

	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLogger struct {
	options *Options
	logger  *zap.SugaredLogger
}

func NewZapLogger(opts ...Option) Logger {
	options := NewOptions(opts...)
	z := &ZapLogger{
		options: options,
	}
	z.logger = z.initLogger()
	return z
}

func (z *ZapLogger) Init(opts ...Option) {
	for _, o := range opts {
		o(z.options)
	}
	z.logger = z.initLogger()
}

func (z ZapLogger) Debug(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Debug(args...)
}

func (z ZapLogger) Debugf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Debugf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) Info(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Info(args...)
}

func (z ZapLogger) Infof(ctx context.Context, template string, args ...interface{}) {
	z.logger.Infof(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) Warn(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Warn(args...)
}

func (z ZapLogger) Warnf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Warnf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) Error(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Error(args...)
}

func (z ZapLogger) Errorf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Errorf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) Fatal(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Fatal(args...)
}

func (z ZapLogger) Fatalf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Fatalf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) DPanic(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.DPanic(args...)
}

func (z ZapLogger) DPanicf(ctx context.Context, template string, args ...interface{}) {
	z.logger.DPanicf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) Panic(ctx context.Context, args ...interface{}) {
	args = append([]interface{}{fmt.Sprintf(" [ServerIP:%s,NodeID:%d] ", shared.GetInternalIP(), shared.GetNodeID())}, args...)
	z.logger.Panic(args...)
}

func (z ZapLogger) Panicf(ctx context.Context, template string, args ...interface{}) {
	z.logger.Panicf(" [ServerIP:%s,NodeID:%d] "+template, append([]interface{}{shared.GetInternalIP(), shared.GetNodeID()}, args...)...)
}

func (z ZapLogger) initLogger() *zap.SugaredLogger {
	writer := partitionWriter(z.options.configLog)
	// errorWriter := errorWriter(config)

	encoder := encoder()

	var level zapcore.Level

	switch z.options.configLog.Level {
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
	//	cores = append(cores, zapcore.NewCore(encoder, partitionWriter, level))
	//}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, writer, level),
	}

	tee := zapcore.NewTee(cores...)

	logger := zap.New(tee, zap.AddCaller(), zap.AddCallerSkip(1))

	if z.options.configSentry.DSN != "" {
		err := wssentry.DefaultSentryInstance.Init()
		if err != nil {
			panic(err)
		}
		return attachSentry(logger, sentry.CurrentHub().Client()).Sugar()
	}
	return logger.Sugar()
}
func encoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

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

func simpleWriter(config config.Log) zapcore.WriteSyncer {
	lc := config

	if lc.Path != "" {
		// 去除最后的/
		if lc.Path[len(lc.Path)-1] == '/' {
			lc.Path = lc.Path[:len(lc.Path)-1]
		}
		lc.Path = lc.Path + "/normal.log"
	}

	simple, err := os.OpenFile(lc.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	writers := []io.Writer{simple}

	if lc.Print {
		writers = append(writers, os.Stdout)
	}

	return zapcore.AddSync(io.MultiWriter(writers...))
}

//golint:ignore
func partitionWriter(config config.Log) zapcore.WriteSyncer {
	lc := config

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

	syncWriter := &zapcore.BufferedWriteSyncer{
		WS:   zapcore.AddSync(lumberJackLogger),
		Size: 4096,
	}

	writers := []io.Writer{syncWriter}

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
