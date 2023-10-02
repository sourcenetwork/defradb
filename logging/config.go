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
	"io"
	"os"
)

type (
	// EncoderFormat is the format of the log output (JSON, CSV, ...).
	EncoderFormat       = int8
	EncoderFormatOption struct {
		EncoderFormat EncoderFormat
		HasValue      bool
	}
)

// NewEncoderFormatOption creates a new EncoderFormatOption with the given value.
func NewEncoderFormatOption(v EncoderFormat) EncoderFormatOption {
	return EncoderFormatOption{
		EncoderFormat: v,
		HasValue:      true,
	}
}

const (
	stderr = "stderr"
	stdout = "stdout"

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

// Log levels.
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

type DisableColorOption struct {
	DisableColor bool
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

func NewDisableColorOption(disable bool) DisableColorOption {
	return DisableColorOption{
		DisableColor: disable,
		HasValue:     true,
	}
}

type Config struct {
	Level                 LogLevelOption
	EncoderFormat         EncoderFormatOption
	EnableStackTrace      EnableStackTraceOption
	EnableCaller          EnableCallerOption
	DisableColor          DisableColorOption
	OutputPaths           []string
	OverridesByLoggerName map[string]Config

	Pipe io.Writer // this is used for testing purposes only
}

func (c Config) forLogger(name string) Config {
	loggerConfig := Config{
		Level:            c.Level,
		EnableStackTrace: c.EnableStackTrace,
		DisableColor:     c.DisableColor,
		EnableCaller:     c.EnableCaller,
		EncoderFormat:    c.EncoderFormat,
		OutputPaths:      c.OutputPaths,
		Pipe:             c.Pipe,
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
		if override.DisableColor.HasValue {
			loggerConfig.DisableColor = override.DisableColor
		}
		if override.EncoderFormat.HasValue {
			loggerConfig.EncoderFormat = override.EncoderFormat
		}
		if len(override.OutputPaths) != 0 {
			loggerConfig.OutputPaths = override.OutputPaths
		}
		if override.Pipe != nil {
			loggerConfig.Pipe = override.Pipe
		}
	}

	return loggerConfig
}

func (c Config) copy() Config {
	overridesByLoggerName := make(map[string]Config, len(c.OverridesByLoggerName))
	for k, o := range c.OverridesByLoggerName {
		overridesByLoggerName[k] = Config{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EncoderFormat:    o.EncoderFormat,
			EnableCaller:     o.EnableCaller,
			DisableColor:     o.DisableColor,
			OutputPaths:      o.OutputPaths,
			Pipe:             o.Pipe,
		}
	}

	return Config{
		Level:                 c.Level,
		EnableStackTrace:      c.EnableStackTrace,
		EncoderFormat:         c.EncoderFormat,
		OutputPaths:           c.OutputPaths,
		EnableCaller:          c.EnableCaller,
		DisableColor:          c.DisableColor,
		OverridesByLoggerName: overridesByLoggerName,
		Pipe:                  c.Pipe,
	}
}

// Create a new Config given new config options. Each updated Config field is handled.
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

	if newConfigOptions.DisableColor.HasValue {
		newConfig.DisableColor = newConfigOptions.DisableColor
	}

	if newConfigOptions.EncoderFormat.HasValue {
		newConfig.EncoderFormat = newConfigOptions.EncoderFormat
	}

	if len(newConfigOptions.OutputPaths) != 0 {
		newConfig.OutputPaths = validatePaths(newConfigOptions.OutputPaths)
	}

	if newConfigOptions.Pipe != nil {
		newConfig.Pipe = newConfigOptions.Pipe
	}

	for k, o := range newConfigOptions.OverridesByLoggerName {
		// We fully overwrite overrides to allow for ease of
		// reset/removal (can provide empty to return to default)
		newConfig.OverridesByLoggerName[k] = Config{
			Level:            o.Level,
			EnableStackTrace: o.EnableStackTrace,
			EnableCaller:     o.EnableCaller,
			DisableColor:     o.DisableColor,
			EncoderFormat:    o.EncoderFormat,
			OutputPaths:      validatePaths(o.OutputPaths),
			Pipe:             o.Pipe,
		}
	}

	return newConfig
}

// validatePath ensure that all output paths are valid to avoid zap sync errors
// and also to ensure that the logs are not lost.
func validatePaths(paths []string) []string {
	validatedPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == stderr || p == stdout {
			validatedPaths = append(validatedPaths, p)
			continue
		}

		if f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND, 0644); err != nil {
			log.Info("cannot use provided path", NewKV("err", err))
		} else {
			err := f.Close()
			if err != nil {
				log.Info("problem closing file", NewKV("err", err))
			}

			validatedPaths = append(validatedPaths, p)
		}
	}

	return validatedPaths
}

func willOutputToStderrOrStdout(paths []string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == stderr || p == stdout {
			return true
		}
	}
	return false
}
