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
	"io"
	"os"
)

type (
	EncoderFormat       = int8
	EncoderFormatOption struct {
		EncoderFormat EncoderFormat
		HasValue      bool
	}
)

func NewEncoderFormatOption(v EncoderFormat) EncoderFormatOption {
	return EncoderFormatOption{
		EncoderFormat: v,
		HasValue:      true,
	}
}

const (
	stderr = "stderr"

	JSON EncoderFormat = iota
	CSV
)

type (
	LogLevel       = int8
	LogLevelOption struct {
		LogLevel LogLevel
		HasValue bool
	}
)

func NewLogLevelOption(v LogLevel) LogLevelOption {
	return LogLevelOption{
		LogLevel: v,
		HasValue: true,
	}
}

const (
	Debug LogLevel = -1
	Info  LogLevel = 0
	Warn  LogLevel = 1
	Error LogLevel = 2
	Fatal LogLevel = 5
)

type EnableStackTraceOption struct {
	EnableStackTrace bool
	HasValue         bool
}

type EnableCallerOption struct {
	EnableCaller bool
	HasValue     bool
}

func NewEnableStackTraceOption(enable bool) EnableStackTraceOption {
	return EnableStackTraceOption{
		EnableStackTrace: enable,
		HasValue:         true,
	}
}

func NewEnableCallerOption(enable bool) EnableCallerOption {
	return EnableCallerOption{
		EnableCaller: enable,
		HasValue:     true,
	}
}

type Config struct {
	Level                 LogLevelOption
	EncoderFormat         EncoderFormatOption
	EnableStackTrace      EnableStackTraceOption
	EnableCaller          EnableCallerOption
	OutputPaths           []string
	OverridesByLoggerName map[string]OverrideConfig

	pipe io.Writer // this is used for testing purposes only
}

type OverrideConfig struct {
	Level            LogLevelOption
	EncoderFormat    EncoderFormatOption
	EnableStackTrace EnableStackTraceOption
	EnableCaller     EnableCallerOption
	OutputPaths      []string

	pipe io.Writer // this is used for testing purposes only
}

func (c Config) forLogger(name string) Config {
	loggerConfig := Config{
		Level:            c.Level,
		EnableStackTrace: c.EnableStackTrace,
		EnableCaller:     c.EnableCaller,
		EncoderFormat:    c.EncoderFormat,
		OutputPaths:      c.OutputPaths,
		pipe:             c.pipe,
	}

	if override, hasOverride := c.OverridesByLoggerName[name]; hasOverride {
		if override.Level.HasValue {
			loggerConfig.Level = override.Level
		}
		if override.EnableStackTrace.HasValue {
			loggerConfig.EnableStackTrace = override.EnableStackTrace
		}
		if override.EnableCaller.HasValue {
			loggerConfig.EnableCaller = override.EnableCaller
		}
		if override.EncoderFormat.HasValue {
			loggerConfig.EncoderFormat = override.EncoderFormat
		}
		if len(override.OutputPaths) != 0 {
			loggerConfig.OutputPaths = override.OutputPaths
		}
		if override.pipe != nil {
			loggerConfig.pipe = override.pipe
		}
	}

	return loggerConfig
}

func (c Config) copy() Config {
	overridesByLoggerName := make(map[string]OverrideConfig, len(c.OverridesByLoggerName))
	for k, o := range c.OverridesByLoggerName {
		overridesByLoggerName[k] = OverrideConfig{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EncoderFormat:    o.EncoderFormat,
			OutputPaths:      o.OutputPaths,
			pipe:             o.pipe,
		}
	}

	return Config{
		Level:                 c.Level,
		EnableStackTrace:      c.EnableStackTrace,
		EncoderFormat:         c.EncoderFormat,
		OutputPaths:           c.OutputPaths,
		EnableCaller:          c.EnableCaller,
		OverridesByLoggerName: overridesByLoggerName,
		pipe:                  c.pipe,
	}
}

func (oldConfig Config) with(newConfigOptions Config) Config {
	newConfig := oldConfig.copy()

	if newConfigOptions.Level.HasValue {
		newConfig.Level = newConfigOptions.Level
	}

	if newConfigOptions.EnableStackTrace.HasValue {
		newConfig.EnableStackTrace = newConfigOptions.EnableStackTrace
	}

	if newConfigOptions.EnableCaller.HasValue {
		newConfig.EnableCaller = newConfigOptions.EnableCaller
	}

	if newConfigOptions.EncoderFormat.HasValue {
		newConfig.EncoderFormat = newConfigOptions.EncoderFormat
	}

	if len(newConfigOptions.OutputPaths) != 0 {
		newConfig.OutputPaths = validatePaths(newConfigOptions.OutputPaths)
	}

	if newConfigOptions.pipe != nil {
		newConfig.pipe = newConfigOptions.pipe
	}

	for k, o := range newConfigOptions.OverridesByLoggerName {
		// We fully overwrite overrides to allow for ease of
		// reset/removal (can provide empty to return to default)
		newConfig.OverridesByLoggerName[k] = OverrideConfig{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EnableCaller:     o.EnableCaller,
			EncoderFormat:    o.EncoderFormat,
			OutputPaths:      validatePaths(o.OutputPaths),
			pipe:             o.pipe,
		}
	}

	return newConfig
}

// validatePath ensure that all output paths are valid to avoid zap sync errors
// and also to ensure that the logs are not lost.
func validatePaths(paths []string) []string {
	validatedPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == stderr {
			validatedPaths = append(validatedPaths, p)
			continue
		}

		if f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND, 0666); err != nil {
			log.Info(context.Background(), "cannot use provided path", NewKV("err", err))

		} else {
			err := f.Close()
			if err != nil {
				log.Info(context.Background(), "problem closing file", NewKV("err", err))
			}

			validatedPaths = append(validatedPaths, p)
		}
	}

	return validatedPaths
}

func willOutputToStderr(paths []string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == stderr {
			return true
		}
	}
	return false
}
