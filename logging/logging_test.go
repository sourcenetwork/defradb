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
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/errors"
)

func TestLogWritesFatalMessageToLogAndKillsProcess(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
		})

		logger.Fatal(logMessage)
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

	logLines, err := getLogLines(t, logPath)
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
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
			c.EnableStackTrace = NewEnableStackTraceOption(true)
		})

		logger.Fatal(logMessage)
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

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "FATAL", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// no stacktrace will be present since no error was sent to the logger.
	assert.NotContains(t, logLines[0], "stacktrace")
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogWritesFatalEMessageToLogAndKillsProcess(t *testing.T) {
	logMessage := "test log message"

	if os.Getenv("OS_EXIT") == "1" {
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
		})

		logger.FatalE(logMessage, errors.New("dummy error"))
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

	logLines, err := getLogLines(t, logPath)
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
		logPath := os.Getenv("LOG_PATH")
		logger, logPath := getLogger(t, func(c *Config) {
			c.OutputPaths = []string{logPath}
			c.EnableStackTrace = NewEnableStackTraceOption(true)
		})

		logger.FatalE(logMessage, errors.New("dummy error"))
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

	logLines, err := getLogLines(t, logPath)
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
	LogFunc          func(Logger, string)
	ExpectedLogLevel string
	WithStackTrace   bool
	ExpectStackTrace bool
	WithCaller       bool
}

func logDebug(l Logger, m string)  { l.Debug(m) }
func logInfo(l Logger, m string)   { l.Info(m) }
func logError(l Logger, m string)  { l.Error(m) }
func logErrorE(l Logger, m string) { l.ErrorE(m, errors.New("test error")) }

func getLogLevelTestCase() []LogLevelTestCase {
	return []LogLevelTestCase{
		{Debug, logDebug, "DEBUG", false, false, true},
		{Debug, logDebug, "DEBUG", false, false, false},
		{Debug, logInfo, "INFO", false, false, false},
		{Debug, logError, "ERROR", false, false, false},
		{Debug, logError, "ERROR", true, false, false},
		{Debug, logErrorE, "ERROR", false, false, false},
		{Debug, logErrorE, "ERROR", true, true, false},
		{Info, logDebug, "", false, false, false},
		{Info, logInfo, "INFO", false, false, true},
		{Info, logInfo, "INFO", false, false, false},
		{Info, logError, "ERROR", false, false, false},
		{Info, logError, "ERROR", true, false, false},
		{Info, logErrorE, "ERROR", false, false, false},
		{Info, logErrorE, "ERROR", true, true, false},
		{Warn, logDebug, "", false, false, false},
		{Warn, logInfo, "", false, false, false},
		{Warn, logError, "ERROR", false, false, false},
		{Warn, logError, "ERROR", true, false, false},
		{Warn, logErrorE, "ERROR", false, false, false},
		{Warn, logErrorE, "ERROR", true, true, false},
		{Error, logDebug, "", false, false, false},
		{Error, logInfo, "", false, false, false},
		{Error, logError, "ERROR", false, false, true},
		{Error, logError, "ERROR", false, false, false},
		{Error, logError, "ERROR", true, false, false},
		{Error, logErrorE, "ERROR", false, false, false},
		{Error, logErrorE, "ERROR", true, true, false},
		{Fatal, logDebug, "", false, false, true},
		{Fatal, logDebug, "", false, false, false},
		{Fatal, logInfo, "", false, false, false},
		{Fatal, logError, "", false, false, false},
		{Fatal, logErrorE, "", false, false, false},
	}
}

func TestLogWritesMessagesToLog(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.EnableStackTrace = NewEnableStackTraceOption(tc.WithStackTrace)
			c.EnableCaller = NewEnableCallerOption(tc.WithCaller)
		})
		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := getLogLines(t, logPath)
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

		clearRegistry("TestLogName")
	}
}

func TestLogWritesMessagesToLogGivenUpdatedLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(Fatal)
		})
		SetConfig(Config{
			Level:            NewLogLevelOption(tc.LogLevel),
			EnableStackTrace: NewEnableStackTraceOption(tc.WithStackTrace),
			EnableCaller:     NewEnableCallerOption(tc.WithCaller),
		})
		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := getLogLines(t, logPath)
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

		clearRegistry("TestLogName")
	}
}

func TestLogWritesMessagesToLogGivenUpdatedContextLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
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

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := getLogLines(t, logPath)
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

		clearRegistry("TestLogName")
	}
}

func TestLogDoesntWriteMessagesToLogGivenNoLogPath(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		b := &bytes.Buffer{}
		logger, _ := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{}
			c.Pipe = b
		})

		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := parseLines(b)
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

		clearRegistry("TestLogName")
	}
}

func TestLogDoesntWriteMessagesToLogGivenNotFoundLogPath(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		b := &bytes.Buffer{}
		logger, _ := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{"/path/not/found"}
			c.Pipe = b
		})

		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := parseLines(b)
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

		clearRegistry("TestLogName")
	}
}

func TestLogDoesntWriteMessagesToLogGivenStderrLogPath(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		b := &bytes.Buffer{}
		logger, _ := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{stderr}
			c.Pipe = b
		})

		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := parseLines(b)
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

		clearRegistry("TestLogName")
	}
}

func TestLogWritesMessagesToLogGivenUpdatedLogPath(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for _, tc := range getLogLevelTestCase() {
		logger, _ := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.OutputPaths = []string{}
		})

		dir := t.TempDir()
		logPath := dir + "/log.txt"
		SetConfig(Config{
			OutputPaths: []string{logPath},
		})
		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := getLogLines(t, logPath)
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

		clearRegistry("TestLogName")
	}
}

func logFeedbackInfo(l Logger, m string)  { l.FeedbackInfo(m) }
func logFeedbackError(l Logger, m string) { l.FeedbackError(m) }
func logFeedbackErrorE(l Logger, m string) {
	l.FeedbackErrorE(m, errors.New("test error"))
}

func getFeedbackLogLevelTestCase() []LogLevelTestCase {
	return []LogLevelTestCase{
		{Debug, logFeedbackInfo, "INFO", false, false, false},
		{Debug, logFeedbackError, "ERROR", false, false, false},
		{Debug, logFeedbackError, "ERROR", true, false, false},
		{Debug, logFeedbackErrorE, "ERROR", false, false, false},
		{Debug, logFeedbackErrorE, "ERROR", true, true, false},
		{Info, logFeedbackInfo, "INFO", false, false, true},
		{Info, logFeedbackInfo, "INFO", false, false, false},
		{Info, logFeedbackError, "ERROR", false, false, false},
		{Info, logFeedbackError, "ERROR", true, false, false},
		{Info, logFeedbackErrorE, "ERROR", false, false, false},
		{Info, logFeedbackErrorE, "ERROR", true, true, false},
		{Warn, logFeedbackInfo, "", false, false, false},
		{Warn, logFeedbackError, "ERROR", false, false, false},
		{Warn, logFeedbackError, "ERROR", true, false, false},
		{Warn, logFeedbackErrorE, "ERROR", false, false, false},
		{Warn, logFeedbackErrorE, "ERROR", true, true, false},
		{Error, logFeedbackInfo, "", false, false, false},
		{Error, logFeedbackError, "ERROR", false, false, true},
		{Error, logFeedbackError, "ERROR", false, false, false},
		{Error, logFeedbackError, "ERROR", true, false, false},
		{Error, logFeedbackErrorE, "ERROR", false, false, false},
		{Error, logFeedbackErrorE, "ERROR", true, true, false},
		{Fatal, logFeedbackInfo, "", false, false, false},
		{Fatal, logFeedbackError, "", false, false, false},
		{Fatal, logFeedbackErrorE, "", false, false, false},
	}
}

func TestLogWritesMessagesToFeedbackLog(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	for i, tc := range getFeedbackLogLevelTestCase() {
		b := &bytes.Buffer{}
		logger, logPath := getLogger(t, func(c *Config) {
			c.Level = NewLogLevelOption(tc.LogLevel)
			c.EnableStackTrace = NewEnableStackTraceOption(tc.WithStackTrace)
			c.EnableCaller = NewEnableCallerOption(tc.WithCaller)
			c.Pipe = b
		})
		logMessage := "test log message"

		tc.LogFunc(logger, logMessage)
		logger.Flush()

		logLines, err := getLogLines(t, logPath)
		if err != nil {
			t.Fatal(err)
		}

		if tc.ExpectedLogLevel == "" {
			assert.Len(t, logLines, 0)
		} else {
			if len(logLines) != 1 {
				t.Fatalf("expecting exactly 1 log line but got %d lines for tc %d", len(logLines), i)
			}

			assert.Equal(t, logMessage, logLines[0]["msg"])
			assert.Equal(t, tc.ExpectedLogLevel, logLines[0]["level"])
			assert.Equal(t, "TestLogName", logLines[0]["logger"])
			_, hasStackTrace := logLines[0]["stacktrace"]
			assert.Equal(t, tc.ExpectStackTrace, hasStackTrace)
			_, hasCaller := logLines[0]["caller"]
			assert.Equal(t, tc.WithCaller, hasCaller)
		}

		if tc.ExpectStackTrace {
			assert.Contains(t, b.String(), logMessage+"\ntest error. Stack:")
		} else {
			assert.Equal(t, logMessage+"\n", b.String())
		}

		clearRegistry("TestLogName")
	}
}

func TestLogWritesMessagesToLogGivenPipeWithValidPath(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	b := &bytes.Buffer{}
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
		c.Pipe = b
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideForAnotherLoggerReducingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
		c.OverridesByLoggerName = map[string]Config{
			"not this logger": {Level: NewLogLevelOption(Info)},
		}
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, logLines, 0)
}

func TestLogWritesMessagesToLogGivenOverrideForLoggerReducingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
		c.OverridesByLoggerName = map[string]Config{
			"TestLogName": {Level: NewLogLevelOption(Info)},
		}
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogWritesMessagesToLogGivenOverrideForLoggerRaisingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
		c.OverridesByLoggerName = map[string]Config{
			"not this logger": {Level: NewLogLevelOption(Fatal)},
		}
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideForLoggerRaisingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
		c.OverridesByLoggerName = map[string]Config{
			"TestLogName": {Level: NewLogLevelOption(Fatal)},
		}
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, logLines, 0)
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideUpdatedForAnotherLoggerReducingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]Config{
			"not this logger": {Level: NewLogLevelOption(Info)},
		},
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, logLines, 0)
}

func TestLogWritesMessagesToLogGivenOverrideUpdatedForLoggerReducingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Fatal)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]Config{
			"TestLogName": {Level: NewLogLevelOption(Info)},
		},
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogWritesMessagesToLogGivenOverrideUpdatedForAnotherLoggerRaisingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]Config{
			"not this logger": {Level: NewLogLevelOption(Fatal)},
		},
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}

	assert.Equal(t, logMessage, logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
	// caller is disabled by default
	assert.NotContains(t, logLines[0], "logging_test.go")
}

func TestLogDoesNotWriteMessagesToLogGivenOverrideUpdatedForLoggerRaisingLogLevel(t *testing.T) {
	defer clearConfig()
	defer clearRegistry("TestLogName")
	logger, logPath := getLogger(t, func(c *Config) {
		c.Level = NewLogLevelOption(Info)
	})
	SetConfig(Config{
		OverridesByLoggerName: map[string]Config{
			"TestLogName": {Level: NewLogLevelOption(Fatal)},
		},
	})
	logMessage := "test log message"

	logger.Info(logMessage)
	logger.Flush()

	logLines, err := getLogLines(t, logPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, logLines, 0)
}

func TestGetGoLogger(t *testing.T) {
	l := GetGoLogger("TestLogName")
	assert.NotNil(t, l.ZapEventLogger)
	assert.NotNil(t, l.logger)
}

func TestGetGoLoggerAndApplyConfig(t *testing.T) {
	l := GetGoLogger("TestLogName")
	assert.NotNil(t, l.ZapEventLogger)
	assert.NotNil(t, l.logger)

	b := &bytes.Buffer{}
	l.ApplyConfig(Config{
		EncoderFormat: NewEncoderFormatOption(JSON),
		Pipe:          b,
	})

	l.ZapEventLogger.Info("some info")

	logLines, err := parseLines(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}
	assert.Equal(t, "some info", logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
}

func TestGetGoLoggerV2(t *testing.T) {
	l := GetGoLoggerV2("TestLogName")
	assert.NotNil(t, l.ZapEventLogger)
	assert.NotNil(t, l.logger)
}

func TestGetGoLoggerV2AndApplyConfig(t *testing.T) {
	l := GetGoLoggerV2("TestLogName")
	assert.NotNil(t, l.ZapEventLogger)
	assert.NotNil(t, l.logger)

	b := &bytes.Buffer{}
	l.ApplyConfig(Config{
		EncoderFormat: NewEncoderFormatOption(JSON),
		Pipe:          b,
	})

	l.ZapEventLogger.Info("some info")

	logLines, err := parseLines(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(logLines) != 1 {
		t.Fatalf("expecting exactly 1 log line but got %d lines", len(logLines))
	}
	assert.Equal(t, "some info", logLines[0]["msg"])
	assert.Equal(t, "INFO", logLines[0]["level"])
	assert.Equal(t, "TestLogName", logLines[0]["logger"])
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
	return logger, getFirstOutputPath(logConfig.OutputPaths)
}

func getFirstOutputPath(outputPaths []string) string {
	if len(outputPaths) == 0 {
		return stderr
	}
	return outputPaths[0]
}

var errloggingToConsole = errors.New("no file to open. Logging to console")

func getLogLines(t *testing.T, logPath string) ([]map[string]any, error) {
	if logPath == stderr {
		return nil, errloggingToConsole
	}

	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	return parseLines(file)
}

func parseLines(r io.Reader) ([]map[string]any, error) {
	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	logLines := []map[string]any{}
	for fileScanner.Scan() {
		loggedLine := make(map[string]any)
		err := json.Unmarshal(fileScanner.Bytes(), &loggedLine)
		if err != nil {
			return nil, err
		}
		logLines = append(logLines, loggedLine)
	}

	return logLines, nil
}

func clearRegistry(name string) {
	for _, logger := range registry[name] {
		logger.Flush()
	}
	registry[name] = []Logger{}
}

func clearConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	cachedConfig = Config{}
}
