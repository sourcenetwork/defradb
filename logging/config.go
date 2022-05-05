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

func NewEnableStackTraceOption(enable bool) EnableStackTraceOption {
	return EnableStackTraceOption{
		EnableStackTrace: enable,
		HasValue:         true,
	}
}

type Config struct {
	Level                 LogLevelOption
	EnableStackTrace      EnableStackTraceOption
	EncoderFormat         EncoderFormatOption
	OutputPaths           []string
	OverridesByLoggerName map[string]OverrideConfig
}

type OverrideConfig struct {
	Level            LogLevelOption
	EnableStackTrace EnableStackTraceOption
	EncoderFormat    EncoderFormatOption
	OutputPaths      []string
}

func (c Config) forLogger(name string) Config {
	loggerConfig := Config{
		Level:            c.Level,
		EnableStackTrace: c.EnableStackTrace,
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

	if newConfigOptions.EncoderFormat.HasValue {
		newConfig.EncoderFormat = newConfigOptions.EncoderFormat
	}

	if len(newConfigOptions.OutputPaths) != 0 {
		newConfig.OutputPaths = newConfigOptions.OutputPaths
	}

	for k, o := range newConfigOptions.OverridesByLoggerName {
		// We fully overwrite overrides to allow for ease of
		// reset/removal (can provide empty to return to default)
		newConfig.OverridesByLoggerName[k] = OverrideConfig{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EncoderFormat:    o.EncoderFormat,
			OutputPaths:      o.OutputPaths,
		}
	}

	return newConfig
}
