// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package logging

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"sync"

	golog "github.com/ipfs/go-log"
	gologV2 "github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	name          string
	logger        *zap.Logger
	consoleLogger *stdlog.Logger
	syncLock      sync.RWMutex
}

var _ Logger = (*logger)(nil)

func mustNewLogger(name string) *logger {
	l, err := buildZapLogger(name, Config{})
	if err != nil {
		panic(err)
	}

	return &logger{
		name:   name,
		logger: l,
	}
}

func (l *logger) Debug(ctx context.Context, message string, keyvals ...KV) {
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Debug(message, toZapFields(keyvals)...)
}

func (l *logger) Info(ctx context.Context, message string, keyvals ...KV) {
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Info(message, toZapFields(keyvals)...)
}

func (l *logger) Error(ctx context.Context, message string, keyvals ...KV) {
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Error(message, toZapFields(keyvals)...)
}

func (l *logger) ErrorE(ctx context.Context, message string, err error, keyvals ...KV) {
	kvs := keyvals
	kvs = append(kvs, NewKV("Error", err.Error()))
	kvs = withStackTrace(err, kvs)

	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Error(message, toZapFields(kvs)...)
}

func (l *logger) Fatal(ctx context.Context, message string, keyvals ...KV) {
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Fatal(message, toZapFields(keyvals)...)
}

func (l *logger) FatalE(ctx context.Context, message string, err error, keyvals ...KV) {
	kvs := keyvals
	kvs = append(kvs, NewKV("Error", err.Error()))
	kvs = withStackTrace(err, kvs)

	l.syncLock.RLock()
	defer l.syncLock.RUnlock()

	l.logger.Fatal(message, toZapFields(kvs)...)
}

func (l *logger) FeedbackInfo(ctx context.Context, message string, keyvals ...KV) {
	l.Info(ctx, message, keyvals...)
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()
	if l.consoleLogger != nil {
		l.consoleLogger.Println(message)
	}
}

func (l *logger) FeedbackError(ctx context.Context, message string, keyvals ...KV) {
	l.Error(ctx, message, keyvals...)
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()
	if l.consoleLogger != nil {
		l.consoleLogger.Println(message)
	}
}

func (l *logger) FeedbackErrorE(ctx context.Context, message string, err error, keyvals ...KV) {
	l.ErrorE(ctx, message, err, keyvals...)
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()
	if l.consoleLogger != nil {
		l.consoleLogger.Println(message)
		if stack, hasStack := getStackTrace(err); hasStack {
			l.consoleLogger.Println(stack)
		}
	}
}

func (l *logger) FeedbackFatal(ctx context.Context, message string, keyvals ...KV) {
	l.Fatal(ctx, message, keyvals...)
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()
	if l.consoleLogger != nil {
		l.consoleLogger.Println(message)
	}
}

func (l *logger) FeedbackFatalE(ctx context.Context, message string, err error, keyvals ...KV) {
	l.FatalE(ctx, message, err, keyvals...)
	l.syncLock.RLock()
	defer l.syncLock.RUnlock()
	if l.consoleLogger != nil {
		l.consoleLogger.Println(message)
		if stack, hasStack := getStackTrace(err); hasStack {
			l.consoleLogger.Println(stack)
		}
	}
}

func (l *logger) Flush() error {
	return l.logger.Sync()
}

func toZapFields(keyvals []KV) []zap.Field {
	result := make([]zap.Field, len(keyvals))
	for i, kv := range keyvals {
		result[i] = zap.Any(kv.key, kv.value)
	}
	return result
}

func (l *logger) ApplyConfig(config Config) {
	newLogger, err := buildZapLogger(l.name, config)
	if err != nil {
		l.logger.Error("Error applying config to logger", zap.Error(err))
		return
	}

	l.syncLock.Lock()
	defer l.syncLock.Unlock()

	// We need sync the old log before swapping it out
	_ = l.logger.Sync()
	l.logger = newLogger

	if !willOutputToStderrOrStdout(config.OutputPaths) {
		if config.Pipe != nil { // for testing purposes only
			l.consoleLogger = stdlog.New(config.Pipe, "", 0)
		} else {
			l.consoleLogger = stdlog.New(os.Stderr, "", 0)
		}
	} else {
		l.consoleLogger = nil
	}
}

func withStackTrace(err error, keyvals []KV) []KV {
	if stack, hasStack := getStackTrace(err); hasStack {
		return append(keyvals, NewKV("stacktrace", stack))
	}

	return keyvals
}

func getStackTrace(err error) (string, bool) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if cachedConfig.EnableStackTrace.EnableStackTrace {
		return fmt.Sprintf("%+v", err), true
	}

	return "", false
}

func buildZapLogger(name string, config Config) (*zap.Logger, error) {
	const (
		encodingTypeConsole string = "console"
		encodingTypeJSON    string = "json"
	)
	defaultConfig := zap.NewProductionConfig()
	defaultConfig.Encoding = encodingTypeConsole
	defaultConfig.EncoderConfig.ConsoleSeparator = ", "
	defaultConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	defaultConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	defaultConfig.DisableStacktrace = true
	defaultConfig.DisableCaller = true

	if config.Level.HasValue {
		defaultConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(config.Level.LogLevel))
	}

	if config.DisableColor.HasValue && config.DisableColor.DisableColor {
		defaultConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	if config.EnableCaller.HasValue {
		defaultConfig.DisableCaller = !config.EnableCaller.EnableCaller
	}

	if config.EncoderFormat.HasValue {
		if config.EncoderFormat.EncoderFormat == JSON {
			defaultConfig.Encoding = encodingTypeJSON
			defaultConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		} else if config.EncoderFormat.EncoderFormat == CSV {
			defaultConfig.Encoding = encodingTypeConsole
		}
	}

	if len(config.OutputPaths) != 0 {
		defaultConfig.OutputPaths = config.OutputPaths[:]
	}

	// We must skip the first caller, as this will always be our wrapper
	newLogger, err := defaultConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	if willOutputToStderrOrStdout(defaultConfig.OutputPaths) && config.Pipe != nil {
		newLogger = newLogger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
			cfg := zap.NewProductionEncoderConfig()
			cfg.ConsoleSeparator = defaultConfig.EncoderConfig.ConsoleSeparator
			cfg.EncodeTime = defaultConfig.EncoderConfig.EncodeTime
			cfg.EncodeLevel = defaultConfig.EncoderConfig.EncodeLevel
			return zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg),
				zapcore.Lock(zapcore.AddSync(config.Pipe)),
				zap.NewAtomicLevelAt(zapcore.Level(config.Level.LogLevel)),
			)
		}))
	}

	return newLogger.Named(name), nil
}

/*
	The following are wrappers for external packages loggers that are compatible with
	our own logger (i.e. Zap based). They offer a way to access the internal logger stores
	and apply our configuration. They should implement ApplyConfig.
*/

// goLogger is a wrapper for a go-log logger
// Used by github.com/ipfs/go-ipfs-provider
type goLogger struct {
	*logger
	*golog.ZapEventLogger
}

func GetGoLogger(name string) *goLogger {
	l := mustNewLogger(name)
	gl := golog.Logger(name)
	return &goLogger{
		logger:         l,
		ZapEventLogger: gl,
	}
}

func (l *goLogger) ApplyConfig(config Config) {
	l.logger.ApplyConfig(config)
	l.ZapEventLogger.SugaredLogger = *l.logger.logger.Sugar()
}

// goLoggerV2 is a wrapper for a go-log V2 logger
// Used by github.com/sourcenetwork/defradb/datastore/badger/v4
type goLoggerV2 struct {
	*logger
	*gologV2.ZapEventLogger
}

func GetGoLoggerV2(name string) *goLoggerV2 {
	l := mustNewLogger(name)
	gl := gologV2.Logger(name)
	return &goLoggerV2{
		logger:         l,
		ZapEventLogger: gl,
	}
}

func (l *goLoggerV2) ApplyConfig(config Config) {
	l.logger.ApplyConfig(config)
	l.ZapEventLogger.SugaredLogger = *l.logger.logger.Sugar()
}
