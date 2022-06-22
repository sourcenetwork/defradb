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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0], "stacktrace")
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	assert.Contains(t, logLines[0], "stacktrace")
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

type LogLevelTestCase struct {
	LogLevel         LogLevel
	LogFunc          func(Logger, context.Context, string)
	ExpectedLogLevel string
	WithStackTrace   bool
	ExpectStackTrace bool
	WithCaller       bool
}

func logDebug(l Logger, c context.Context, m string)  { l.Debug(c, m) }
func logInfo(l Logger, c context.Context, m string)   { l.Info(c, m) }
func logWarn(l Logger, c context.Context, m string)   { l.Warn(c, m) }
func logError(l Logger, c context.Context, m string)  { l.Error(c, m) }
func logErrorE(l Logger, c context.Context, m string) { l.ErrorE(c, m, fmt.Errorf("test error")) }

func getLogLevelTestCase() []LogLevelTestCase {
	return []LogLevelTestCase{
		{Debug, logDebug, "DEBUG", false, false, true},
		{Debug, logDebug, "DEBUG", false, false, false},
		{Debug, logInfo, "INFO", false, false, false},
		{Debug, logWarn, "WARN", false, false, false},
		{Debug, logError, "ERROR", false, false, false},
		{Debug, logError, "ERROR", true, true, false},
		{Debug, logErrorE, "ERROR", false, false, false},
		{Debug, logErrorE, "ERROR", true, true, false},
		{Info, logDebug, "", false, false, false},
		{Info, logInfo, "INFO", false, false, true},
		{Info, logInfo, "INFO", false, false, false},
		{Info, logWarn, "WARN", false, false, false},
		{Info, logError, "ERROR", false, false, false},
		{Info, logError, "ERROR", true, true, false},
		{Info, logErrorE, "ERROR", false, false, false},
		{Info, logErrorE, "ERROR", true, true, false},
		{Warn, logDebug, "", false, false, false},
		{Warn, logInfo, "", false, false, false},
		{Warn, logWarn, "WARN", false, false, true},
		{Warn, logWarn, "WARN", false, false, false},
		{Warn, logError, "ERROR", false, false, false},
		{Warn, logError, "ERROR", true, true, false},
		{Warn, logErrorE, "ERROR", false, false, false},
		{Warn, logErrorE, "ERROR", true, true, false},
		{Error, logDebug, "", false, false, false},
		{Error, logInfo, "", false, false, false},
		{Error, logWarn, "", false, false, false},
		{Error, logError, "ERROR", false, false, true},
		{Error, logError, "ERROR", false, false, false},
		{Error, logError, "ERROR", true, true, false},
		{Error, logErrorE, "ERROR", false, false, false},
		{Error, logErrorE, "ERROR", true, true, false},
		{Fatal, logDebug, "", false, false, true},
		{Fatal, logDebug, "", false, false, false},
		{Fatal, logInfo, "", false, false, false},
		{Fatal, logWarn, "", false, false, false},
		{Fatal, logError, "", false, false, false},
		{Fatal, logErrorE, "", false, false, false},
	}
}

func TestLogWritesMessagesToLog(t *testing.T) {
	for _, tc := range getLogLevelTestCase() {
		ctx := context.Background()
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.EnableStackTrace = NewEnableStackTraceOption(tc.WithStackTrace)
			c.EnableCaller = NewEnableCallerOption(tc.WithCaller)
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)
		if err != nil {
			t.Fatal(err)
		}

		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			if len(logLines) != 1 {
				t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
			}

			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			_, hasCaller := logLines[0]["caller"]
			assert.Equal(t, tc.WithCaller, hasCaller)
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
			EnableCaller:     NewEnableCallerOption(tc.WithCaller),
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)
		if err != nil {
			t.Fatal(err)
		}

		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			if len(logLines) != 1 {
				t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
			}

			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			_, hasCaller := logLines[0]["caller"]
			assert.Equal(t, tc.WithCaller, hasCaller)
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
			EnableCaller:     NewEnableCallerOption(tc.WithCaller),
		})
		logMessage := "test log message"

		tc.LogFunc(logger, ctx, logMessage)
		logger.Flush()

		logLines, err := getLogLines(logPath)
		if err != nil {
			t.Fatal(err)
		}

		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			if len(logLines) != 1 {
				t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
			}

			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			_, hasCaller := logLines[0]["caller"]
			assert.Equal(t, tc.WithCaller, hasCaller)
		}
	}
}

// This test is largely a sanity check for `TestLogWritesMessagesToLogGivenUpdatedLogPath`
func TestLogDoesntWriteMessagesToLogGivenNoLogPath(t *testing.T) {
	// making it clear that we are setting the config with an invalid path
	logConfig := Config{
		EncoderFormat: NewEncoderFormatOption(JSON),
		OutputPaths:   []string{"/path/not/found"},
	}
	setConfig(logConfig)
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
		if err != nil {
			t.Fatal(err)
		}

		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			if len(logLines) != 1 {
				t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
			}

			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
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
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "WARN", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
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
	if err != nil {
		t.Fatal(err)
	}

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
	defer file.Close()

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
