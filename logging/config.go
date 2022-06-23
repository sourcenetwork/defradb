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

import "os"

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
}

type OverrideConfig struct {
	Level            LogLevelOption
	EncoderFormat    EncoderFormatOption
	EnableStackTrace EnableStackTraceOption
	EnableCaller     EnableCallerOption
	OutputPaths      []string
}

func (c Config) forLogger(name string) Config {
	loggerConfig := Config{
		Level:            c.Level,
		EnableStackTrace: c.EnableStackTrace,
		EnableCaller:     c.EnableCaller,
		EncoderFormat:    c.EncoderFormat,
		OutputPaths:      c.OutputPaths,
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
		}
	}

	return Config{
		Level:                 c.Level,
		EnableStackTrace:      c.EnableStackTrace,
		EncoderFormat:         c.EncoderFormat,
		OutputPaths:           c.OutputPaths,
		EnableCaller:          c.EnableCaller,
		OverridesByLoggerName: overridesByLoggerName,
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

	for k, o := range newConfigOptions.OverridesByLoggerName {
		// We fully overwrite overrides to allow for ease of
		// reset/removal (can provide empty to return to default)
		newConfig.OverridesByLoggerName[k] = OverrideConfig{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EnableCaller:     o.EnableCaller,
			EncoderFormat:    o.EncoderFormat,
			OutputPaths:      validatePaths(o.OutputPaths),
		}
	}

	return newConfig
}

// validatePath ensure that all output paths are valid to avoid zap sync errors
// and also to ensure that the logs are not lost.
func validatePaths(paths []string) []string {
	validatedPaths := paths
	for i := 0; i < len(validatedPaths); i++ {
		if validatedPaths[i] == "stdout" {
			continue
		}

		if f, err := os.Create(validatedPaths[i]); os.IsNotExist(err) {
			validatedPaths[i] = validatedPaths[len(validatedPaths)-1]
			validatedPaths = validatedPaths[:len(validatedPaths)-1]
		} else {
			f.Close()
		}
	}
	return validatedPaths
}
