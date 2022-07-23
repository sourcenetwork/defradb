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
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

const (
	stderrPath  = "stderr"
	testLogger1 = "testLogger1"
	testLogger2 = "testLogger2"
	testLogger3 = "testLogger3"
)

var (
	log1 = logging.MustNewLogger(testLogger1)
	log2 = logging.MustNewLogger(testLogger2)
	log3 = logging.MustNewLogger(testLogger3)
)

// todo - add test asserting that logger logs to file by default

func TestCLILogsToStderrGivenNamedLogLevel(t *testing.T) {
	ctx := context.Background()
	logLines := captureLogLines(
		t,
		func() {
			// set the log levels
			os.Args = append(os.Args, "--loglevel")
			// general: error
			// testLogger1: debug
			// testLogger2: info
			os.Args = append(os.Args, fmt.Sprintf("%s,%s=debug,%s=info", "error", testLogger1, testLogger2))
		},
		func() {
			log1.Error(ctx, "error")
			log1.Debug(ctx, "debug")
			log2.Info(ctx, "info")
			log3.Debug(ctx, "info")
		},
	)

	assert.Len(t, logLines, 3)
}

func captureLogLines(t *testing.T, setup func(), predicate func()) []string {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
	}()

	directory := t.TempDir()

	// Set the default logger output path to a file in the temp dir
	// so that production logs don't polute and confuse the tests
	os.Args = append(os.Args, "--logoutput", directory+"/log.txt")
	os.Args = append(os.Args, "init", directory)

	setup()
	cli.Execute()
	predicate()
	log1.Flush()
	log2.Flush()
	log3.Flush()

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
