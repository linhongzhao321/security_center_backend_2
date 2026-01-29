// Package logger 仅出于统一日志格式，规范必填字段而做适当封装，注意避免过度封装
// TODO @林鸿钊 需要根据公司日志规范进行封装
package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"infrastructure/contextextension"
)

var logger *zap.Logger

func InitLogger(options ...Option) error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	for _, option := range options {
		err := option(&config)
		if err != nil {
			return err
		}
	}

	var err error
	logger, err = config.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.WarnLevel))
	return err
}

type Option func(config *zap.Config) error

func WithOutputPaths(paths ...string) Option {
	return func(config *zap.Config) error {
		config.OutputPaths = paths
		return nil
	}
}

func WithErrorOutputPath(paths ...string) Option {
	return func(config *zap.Config) error {
		config.ErrorOutputPaths = paths
		return nil
	}
}

func OutputMinLevel(level zapcore.Level) Option {
	return func(config *zap.Config) error {
		config.Level = zap.NewAtomicLevelAt(level)
		config.Development = true
		return nil
	}
}

func Write(ctx context.Context, level zapcore.Level, msg string, fields ...zapcore.Field) {
	if logger == nil {
		_, _ = fmt.Fprintln(os.Stderr, `logger.Write() error, logger has not been initialized.`)
		return
	}

	ce := logger.Check(level, msg)
	if ce == nil {
		return
	}

	traceID := `-`
	ctxVal := ctx.Value(contextextension.TraceIDKey)
	if ctxVal != nil {
		traceIDString, ok := ctxVal.(string)
		if !ok {
			_, _ = fmt.Fprintln(os.Stderr, `context.TraceID is not string. type:%T`, ctxVal)
		}
		traceID = traceIDString
	}
	fields = append(fields, zap.String(contextextension.TraceIDKey, traceID))
	ce.Write(fields...)
	return
}
