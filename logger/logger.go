package logger

import (
	"context"

	"ws-cluster/config"
)

var DefaultLogger Logger = NewZapLogger(config.DefaultConfig)

type Logger interface {
	Debug(ctx context.Context, args ...interface{})
	Debugf(ctx context.Context, template string, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Infof(ctx context.Context, template string, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Warnf(ctx context.Context, template string, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Errorf(ctx context.Context, template string, args ...interface{})
	Fatal(ctx context.Context, args ...interface{})
	Fatalf(ctx context.Context, template string, args ...interface{})
	DPanic(ctx context.Context, args ...interface{})
	DPanicf(ctx context.Context, template string, args ...interface{})
	Panic(ctx context.Context, args ...interface{})
	Panicf(ctx context.Context, template string, args ...interface{})
}
