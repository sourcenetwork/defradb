// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package clitest provides a testing framework for the Defra CLI, along with CLI integration tests.
*/
package clitest

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
)

const COMMAND_TIMEOUT_SECONDS = 2 * time.Second
const SUBCOMMAND_TIME_BUFFER_SECONDS = 200 * time.Millisecond

type DefraNodeConfig struct {
	rootDir  string
	logPath  string
	APIURL   string
	GRPCAddr string
}

func NewDefraNodeDefaultConfig(t *testing.T) DefraNodeConfig {
	t.Helper()
	portAPI, err := findFreePortInRange(t, 49152, 65535)
	if err != nil {
		t.Fatal(err)
	}
	portGRPC, err := findFreePortInRange(t, 49152, 65535)
	if err != nil {
		t.Fatal(err)
	}

	return DefraNodeConfig{
		rootDir:  t.TempDir(),
		logPath:  "",
		APIURL:   fmt.Sprintf("localhost:%d", portAPI),
		GRPCAddr: fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", portGRPC),
	}
}

// runDefraNode runs a defra node in a separate goroutine and returns a stopping function
// which also returns the node's execution log lines.
func runDefraNode(t *testing.T, conf DefraNodeConfig) func() []string {
	t.Helper()

	if conf.logPath == "" {
		conf.logPath = filepath.Join(t.TempDir(), "defra.log")
	}

	var args []string
	if conf.rootDir != "" {
		args = append(args, "--rootdir", conf.rootDir)
	}
	if conf.APIURL != "" {
		args = append(args, "--url", conf.APIURL)
	}
	if conf.GRPCAddr != "" {
		args = append(args, "--tcpaddr", conf.GRPCAddr)
	}
	args = append(args, "--logoutput", conf.logPath)

	cfg := config.DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan struct{})
	go func(ready chan struct{}) {
		defraCmd := cli.NewDefraCommand(cfg)
		defraCmd.RootCmd.SetArgs(
			append([]string{"start"}, args...),
		)
		ready <- struct{}{}
		err := defraCmd.Execute(ctx)
		assert.NoError(t, err)
	}(ready)
	<-ready
	time.Sleep(SUBCOMMAND_TIME_BUFFER_SECONDS)
	cancelAndOutput := func() []string {
		cancel()
		time.Sleep(SUBCOMMAND_TIME_BUFFER_SECONDS)
		lines, err := readLoglines(t, conf.logPath)
		assert.NoError(t, err)
		return lines
	}
	return cancelAndOutput
}

// Runs a defra command and returns the stdout and stderr output.
func runDefraCommand(t *testing.T, conf DefraNodeConfig, args []string) (stdout, stderr []string) {
	t.Helper()
	cfg := config.DefaultConfig()
	args = append([]string{
		"--url", conf.APIURL,
	}, args...)
	if !contains(args, "--rootdir") {
		args = append(args, "--rootdir", t.TempDir())
	}

	ctx, cancel := context.WithTimeout(context.Background(), COMMAND_TIMEOUT_SECONDS)
	defer cancel()

	stdout, stderr = captureOutput(func() {
		defraCmd := cli.NewDefraCommand(cfg)
		t.Log("executing defra command with args", args)
		defraCmd.RootCmd.SetArgs(args)
		_ = defraCmd.Execute(ctx)
	})
	return stdout, stderr
}

func contains(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

func readLoglines(t *testing.T, fpath string) ([]string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck
	scanner := bufio.NewScanner(f)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	assert.NoError(t, err)
	return lines, nil
}

func captureOutput(f func()) (stdout, stderr []string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rStdout, wStdout, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	rStderr, wStderr, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = wStdout
	os.Stderr = wStderr

	f()

	if err := wStdout.Close(); err != nil {
		panic(err)
	}
	if err := wStderr.Close(); err != nil {
		panic(err)
	}

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, rStdout); err != nil {
		panic(err)
	}
	if _, err := io.Copy(&stderrBuf, rStderr); err != nil {
		panic(err)
	}

	stdout = strings.Split(strings.TrimSuffix(stdoutBuf.String(), "\n"), "\n")
	stderr = strings.Split(strings.TrimSuffix(stderrBuf.String(), "\n"), "\n")

	return
}

var portsInUse = make(map[int]struct{})
var portMutex = sync.Mutex{}

// findFreePortInRange returns a free port in the range [minPort, maxPort].
// The range of ports that are unfrequently used is [49152, 65535].
func findFreePortInRange(t *testing.T, minPort, maxPort int) (int, error) {
	if minPort < 1 || maxPort > 65535 || minPort > maxPort {
		return 0, errors.New("invalid port range")
	}

	const maxAttempts = 100
	for i := 0; i < maxAttempts; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		if _, ok := portsInUse[port]; ok {
			continue
		}
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			portMutex.Lock()
			portsInUse[port] = struct{}{}
			portMutex.Unlock()
			t.Cleanup(func() {
				portMutex.Lock()
				delete(portsInUse, port)
				portMutex.Unlock()
			})
			_ = listener.Close()
			return port, nil
		}
	}

	return 0, errors.New("unable to find a free port")
}

func assertContainsSubstring(t *testing.T, haystack []string, substring string) {
	t.Helper()
	if !containsSubstring(haystack, substring) {
		t.Fatalf("expected %q to contain %q", haystack, substring)
	}
}

func assertNotContainsSubstring(t *testing.T, haystack []string, substring string) {
	t.Helper()
	if containsSubstring(haystack, substring) {
		t.Fatalf("expected %q to not contain %q", haystack, substring)
	}
}

func containsSubstring(haystack []string, substring string) bool {
	for _, s := range haystack {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

func schemaFileFixture(t *testing.T, fname string, schema string) string {
	absFname := filepath.Join(t.TempDir(), fname)
	err := os.WriteFile(absFname, []byte(schema), 0644)
	assert.NoError(t, err)
	return absFname
}
