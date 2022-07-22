// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

const (
	stderrPath     = "stderr"
	testLoggerName = "testLogger"
)

var (
	log = logging.MustNewLogger(testLoggerName)
)

// todo - add test asserting that logger logs to file by default

func TestCLILogsToStderrGivenNamedLogLevel(t *testing.T) {
	directory := t.TempDir()
	ctx := context.Background()

	logLines := captureLogLines(
		t,
		directory,
		func() {
			// Explicitly set the test logger output to stderr
			//os.Args = append(os.Args, "--loggers")
			//os.Args = append(os.Args, "name="+testLoggerName+",level=stderr")

			cli.Execute()

			log.Error(ctx, "message")
			log.Flush()
		},
	)

	assert.Len(t, logLines, 2)
}

func captureLogLines(t *testing.T, directory string, predicate func()) []string {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
	}()

	os.Args = append(os.Args, "init")
	// Set the db root directory to the given temp dir
	os.Args = append(os.Args, directory)
	// Set the default logger output path to a file in the temp dir
	// so that production logs don't polute and confuse the tests
	os.Args = append(os.Args, "--logoutput")
	os.Args = append(os.Args, directory+"/log.txt")

	predicate()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	logLines, err := parseLines(&buf)
	if err != nil {
		t.Fatal(err)
	}

	return logLines
}

func parseLines(r io.Reader) ([]string, error) {
	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	logLines := []string{}
	for fileScanner.Scan() {
		logLines = append(logLines, fileScanner.Text())
	}

	return logLines, nil
}
