package logger

import (
	"context"
)

var DefaultLogger Logger = NewZapLogger()

type LogLevel string

const (
	DebugLevel  LogLevel = "debug"
	InfoLevel   LogLevel = "info"
	WarnLevel   LogLevel = "warn"
	ErrorLevel  LogLevel = "error"
	DPanicLevel LogLevel = "dpanic"
	PanicLevel  LogLevel = "panic"
	FatalLevel  LogLevel = "fatal"
)

type Logger interface {
	Init(opts ...Option)
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
