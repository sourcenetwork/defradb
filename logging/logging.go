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
)

var log = MustNewLogger("logging")

// KV is a key-value pair used to pass structured data to loggers.
type KV struct {
	key   string
	value any
}

// NewKV creates a new KV key-value pair.
func NewKV(key string, value any) KV {
	return KV{
		key:   key,
		value: value,
	}
}

type Logger interface {
	// Debug logs a message at debug log level. Key-value pairs can be added.
	Debug(ctx context.Context, message string, keyvals ...KV)
	// Info logs a message at info log level. Key-value pairs can be added.
	Info(ctx context.Context, message string, keyvals ...KV)
	// Error logs a message at error log level. Key-value pairs can be added.
	Error(ctx context.Context, message string, keyvals ...KV)
	// ErrorErr logs a message and an error at error log level. Key-value pairs can be added.
	ErrorE(ctx context.Context, message string, err error, keyvals ...KV)
	// Fatal logs a message at fatal log level. Key-value pairs can be added.
	Fatal(ctx context.Context, message string, keyvals ...KV)
	// FatalE logs a message and an error at fatal log level. Key-value pairs can be added.
	FatalE(ctx context.Context, message string, err error, keyvals ...KV)

	// Feedback prefixed method ensure that messsages reach a user in case the logs are sent to a file.

	// FeedbackInfo calls Info and sends the message to stderr if logs are sent to a file.
	FeedbackInfo(ctx context.Context, message string, keyvals ...KV)
	// FeedbackError calls Error and sends the message to stderr if logs are sent to a file.
	FeedbackError(ctx context.Context, message string, keyvals ...KV)
	// FeedbackErrorE calls ErrorE and sends the message to stderr if logs are sent to a file.
	FeedbackErrorE(ctx context.Context, message string, err error, keyvals ...KV)
	// FeedbackFatal calls Fatal and sends the message to stderr if logs are sent to a file.
	FeedbackFatal(ctx context.Context, message string, keyvals ...KV)
	// FeedbackFatalE calls FatalE and sends the message to stderr if logs are sent to a file.
	FeedbackFatalE(ctx context.Context, message string, err error, keyvals ...KV)

	// Flush flushes any buffered log entries.
	Flush() error
	// ApplyConfig updates the logger with a new config.
	ApplyConfig(config Config)
}

// MustNewLogger creates and registers a new logger with the given name, and panics if there is an error.
func MustNewLogger(name string) Logger {
	logger := mustNewLogger(name)
	register(name, logger)
	return logger
}

// SetConfig updates all registered loggers with the given config.
func SetConfig(newConfig Config) {
	updatedConfig := setConfig(newConfig)
	updateLoggers(updatedConfig)
}
