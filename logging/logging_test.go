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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogWritesFatalMessageToLogAndKillsProcess(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		ctx := context.Background()
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
		})

		logger.Fatal(ctx, logMessage)
		return
	}

	dir := t.TempDir()
	logPath := dir + "/log.txt"
	cmd := exec.Command(os.Args[0], "-test.run=TestLogWritesFatalMessageToLogAndKillsProcess")
	cmd.Env = append(os.Environ(), "OS_EXIT=1", "LOG_PATH="+logPath)
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Logger.Fatal failed to kill the process, error: %v", err)
	}

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
	// stacktrace is disabled by default
	assert.NotContains(t, logLines[0], "stacktrace")
}

func TestLogWritesFatalMessageWithStackTraceToLogAndKillsProcessGivenStackTraceEnabled(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		ctx := context.Background()
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
			c.EnableStackTrace = NewEnableStackTraceOption(true)
		})

		logger.Fatal(ctx, logMessage)
		return
	}

	dir := t.TempDir()
	logPath := dir + "/log.txt"
	cmd := exec.Command(os.Args[0], "-test.run=TestLogWritesFatalMessageWithStackTraceToLogAndKillsProcessGivenStackTraceEnabled")
	cmd.Env = append(os.Environ(), "OS_EXIT=1", "LOG_PATH="+logPath)
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Logger.Fatal failed to kill the process, error: %v", err)
	}

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
	assert.Contains(t, logLines[0], "stacktrace")
}

func TestLogWritesFatalEMessageToLogAndKillsProcess(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		ctx := context.Background()
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
		})

		logger.FatalE(ctx, logMessage, fmt.Errorf("dummy error"))
		return
	}

	dir := t.TempDir()
	logPath := dir + "/log.txt"
	cmd := exec.Command(os.Args[0], "-test.run=TestLogWritesFatalEMessageToLogAndKillsProcess")
	cmd.Env = append(os.Environ(), "OS_EXIT=1", "LOG_PATH="+logPath)
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Logger.Fatal failed to kill the process, error: %v", err)
	}

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
	// stacktrace is disabled by default
	assert.NotContains(t, logLines[0], "stacktrace")
}

func TestLogWritesFatalEMessageWithStackTraceToLogAndKillsProcessGivenStackTraceEnabled(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		ctx := context.Background()
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
			c.EnableStackTrace = NewEnableStackTraceOption(true)
		})

		logger.FatalE(ctx, logMessage, fmt.Errorf("dummy error"))
		return
	}

	dir := t.TempDir()
	logPath := dir + "/log.txt"
	cmd := exec.Command(os.Args[0], "-test.run=TestLogWritesFatalEMessageWithStackTraceToLogAndKillsProcessGivenStackTraceEnabled")
	cmd.Env = append(os.Environ(), "OS_EXIT=1", "LOG_PATH="+logPath)
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Logger.Fatal failed to kill the process, error: %v", err)
	}

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
	assert.Contains(t, logLines[0], "stacktrace")
}

type LogLevelTestCase struct {
	LogLevel         LogLevel
	LogFunc          func(Logger, context.Context, string)
	ExpectedLogLevel string
	WithStackTrace   bool
	ExpectStackTrace bool
}

func logDebug(l Logger, c context.Context, m string)  { l.Debug(c, m) }
func logInfo(l Logger, c context.Context, m string)   { l.Info(c, m) }
func logWarn(l Logger, c context.Context, m string)   { l.Warn(c, m) }
func logError(l Logger, c context.Context, m string)  { l.Error(c, m) }
func logErrorE(l Logger, c context.Context, m string) { l.ErrorE(c, m, fmt.Errorf("test error")) }

func getLogLevelTestCase() []LogLevelTestCase {
	return []LogLevelTestCase{
		{Debug, logDebug, "DEBUG", false, false},
		{Debug, logInfo, "INFO", false, false},
		{Debug, logWarn, "WARN", false, false},
		{Debug, logError, "ERROR", false, false},
		{Debug, logError, "ERROR", true, true},
		{Debug, logErrorE, "ERROR", false, false},
		{Debug, logErrorE, "ERROR", true, true},
		{Info, logDebug, "", false, false},
		{Info, logInfo, "INFO", false, false},
		{Info, logWarn, "WARN", false, false},
		{Info, logError, "ERROR", false, false},
		{Info, logError, "ERROR", true, true},
		{Info, logErrorE, "ERROR", false, false},
		{Info, logErrorE, "ERROR", true, true},
		{Warn, logDebug, "", false, false},
		{Warn, logInfo, "", false, false},
		{Warn, logWarn, "WARN", false, false},
		{Warn, logError, "ERROR", false, false},
		{Warn, logError, "ERROR", true, true},
		{Warn, logErrorE, "ERROR", false, false},
		{Warn, logErrorE, "ERROR", true, true},
		{Error, logDebug, "", false, false},
		{Error, logInfo, "", false, false},
		{Error, logWarn, "", false, false},
		{Error, logError, "ERROR", false, false},
		{Error, logError, "ERROR", true, true},
		{Error, logErrorE, "ERROR", false, false},
		{Error, logErrorE, "ERROR", true, true},
		{Fatal, logDebug, "", false, false},
		{Fatal, logInfo, "", false, false},
		{Fatal, logWarn, "", false, false},
		{Fatal, logError, "", false, false},
		{Fatal, logErrorE, "", false, false},
	}
}

func TestLogWritesMessagesToLog(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.EnableStackTrace = NewEnableStackTraceOption(tc.WithStackTrace)
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)

		assert.Nil(t, err)
		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			assert.Len(t, logLines, 1)
			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			assert.Contains(t, logLines[0]["caller"], "logging_test.go")
		}
	}
}

func TestLogWritesMessagesToLogGivenUpdatedLogLevel(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(Fatal)
		})
		SetConfig(Config{
			Level:            NewLogLevelOption(tc.LogLevel),
			EnableStackTrace: NewEnableStackTraceOption(tc.WithStackTrace),
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)

		assert.Nil(t, err)
		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			assert.Len(t, logLines, 1)
			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			assert.Contains(t, logLines[0]["caller"], "logging_test.go")
		}
	}
}

func TestLogWritesMessagesToLogGivenUpdatedContextLogLevel(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(Fatal)
		})
		SetConfig(Config{
			Level: NewLogLevelOption(Error),
		})
		SetConfig(Config{
			Level:            NewLogLevelOption(tc.LogLevel),
			EnableStackTrace: NewEnableStackTraceOption(tc.WithStackTrace),
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)

		assert.Nil(t, err)
		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			assert.Len(t, logLines, 1)
			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			assert.Contains(t, logLines[0]["caller"], "logging_test.go")
		}
	}
}

// This test is largely a sanity check for `TestLogWritesMessagesToLogGivenUpdatedLogPath`
func TestLogDoesntWriteMessagesToLogGivenNoLogPath(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{}
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)

		assert.Errorf(t, err, "PathError")
		assert.Len(t, logLines, 0)
	}
}

func TestLogWritesMessagesToLogGivenUpdatedLogPath(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{}
		})
		SetConfig(Config{
			OutputPaths: []string{logPath},
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)

		assert.Nil(t, err)
		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			assert.Len(t, logLines, 1)
			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			assert.Contains(t, logLines[0]["caller"], "logging_test.go")
		}
	}
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideForAnotherLoggerReducingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
		c.OverridesByLoggerName = map[string]OverrideConfig{
			"not this logger": {Level: NewLogLevelOption(Info)},
		}
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 0)
}

func TestLogWritesMessagesToLogGivenOverrideForLoggerReducingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
		c.OverridesByLoggerName = map[string]OverrideConfig{
			"TestLogName": {Level: NewLogLevelOption(Info)},
		}
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
}

func TestLogWritesMessagesToLogGivenOverrideForLoggerRaisingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
		c.OverridesByLoggerName = map[string]OverrideConfig{
			"not this logger": {Level: NewLogLevelOption(Fatal)},
		}
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideForLoggerRaisingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
		c.OverridesByLoggerName = map[string]OverrideConfig{
			"TestLogName": {Level: NewLogLevelOption(Fatal)},
		}
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 0)
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideUpdatedForAnotherLoggerReducingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]OverrideConfig{
			"not this logger": {Level: NewLogLevelOption(Info)},
		},
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 0)
}

func TestLogWritesMessagesToLogGivenOverrideUpdatedForLoggerReducingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]OverrideConfig{
			"TestLogName": {Level: NewLogLevelOption(Info)},
		},
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
}

func TestLogWritesMessagesToLogGivenOverrideUpdatedForAnotherLoggerRaisingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]OverrideConfig{
			"not this logger": {Level: NewLogLevelOption(Fatal)},
		},
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 1)
	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0]["caller"], "logging_test.go")
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideUpdatedForLoggerRaisingLogLevel(t *testing.T) {
	ctx := context.Background()
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]OverrideConfig{
			"TestLogName": {Level: NewLogLevelOption(Fatal)},
		},
	})
	logMessage := "test log message"

	logger.Warn(ctx, logMessage)
	logger.Flush()

	logLines, err := getLogLines(logPath)

	assert.Nil(t, err)
	assert.Len(t, logLines, 0)
}

type Option = func(*Config)

func getLogger(t *testing.T, options ...Option) (Logger, string) {
	dir := t.TempDir()
	logPath := dir + "/log.txt"
	name := "TestLogName"
	logConfig := Config{
		EncoderFormat: NewEncoderFormatOption(JSON),
		OutputPaths:   []string{logPath},
	}

	for _, o := range options {
		o(&logConfig)
	}

	logger := MustNewLogger(name)
	SetConfig(logConfig)
	return logger, logPath
}

func getLogLines(logPath string) ([]map[string]interface{}, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(file)

	fileScanner.Split(bufio.ScanLines)

	logLines := []map[string]interface{}{}
	for fileScanner.Scan() {
		loggedLine := make(map[string]interface{})
		err = json.Unmarshal(fileScanner.Bytes(), &loggedLine)
		if err != nil {
			return nil, err
		}
		logLines = append(logLines, loggedLine)
	}

	return logLines, nil
}
