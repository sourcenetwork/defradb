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

var log = MustNewLogger("defra.logging")

type KV struct {
	key   string
	value interface{}
}

func NewKV(key string, value interface{}) KV {
	return KV{
		key:   key,
		value: value,
	}
}

type Logger interface {
	Debug(ctx context.Context, message string, keyvals ...KV)
	Info(ctx context.Context, message string, keyvals ...KV)
	Warn(ctx context.Context, message string, keyvals ...KV)
	Error(ctx context.Context, message string, keyvals ...KV)
	ErrorE(ctx context.Context, message string, err error, keyvals ...KV)
	Fatal(ctx context.Context, message string, keyvals ...KV)
	FatalE(ctx context.Context, message string, err error, keyvals ...KV)
	Flush() error
	ApplyConfig(config Config)
}

func MustNewLogger(name string) Logger {
	logger := mustNewLogger(name)
	register(name, logger)
	return logger
}

func SetConfig(newConfig Config) {
	updatedConfig := setConfig(newConfig)
	updateLoggers(updatedConfig)
}
