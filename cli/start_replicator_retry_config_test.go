// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStartReplicatorRetry_NoError(t *testing.T) {

	// Run one command to start the Defra service
	cmd := NewDefraCommand()
	args := []string{
		"start",
		"--no-keyring",
		"--acp-type=local",
		"--replicator-retry-intervals=10,20,40",
	}
	cmd.SetArgs(args)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	// Service runs inside a goroutine
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start command failed: %v", err)
		}
	default:
	}

	// Wait 5 seconds to give the service time to start
	time.Sleep(5.0 * time.Second)

	// Run another command to check the peer ID. If it exists, then the service started
	cmd2 := NewDefraCommand()
	cmd2.SetArgs([]string{"client", "p2p", "info"})
	var output bytes.Buffer
	cmd2.SetOut(&output)

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("Failed to get peer info: %v", err)
	}

	var result struct {
		ID    string   `json:"ID"`
		Addrs []string `json:"Addrs"`
	}

	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON containing peer ID: %v", err)
	}

	if result.ID == "" {
		t.Fatal("Peer ID is empty. Service did not start.")
	}
}

func TestStartReplicatorRetry_NegativeIntervalError(t *testing.T) {
	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"start", "--replicator-retry-intervals=10,-20,40"})
	err := cmd.Execute()
	require.ErrorIs(t, err, ErrInvalidReplicatorRetryIntervals)
}

func TestStartReplicatorRetry_InvalidIntervalError(t *testing.T) {
	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"start", "--replicator-retry-intervals=garbage"})
	err := cmd.Execute()
	expectedError := "invalid argument \"garbage\" for \"--replicator-retry-intervals\" flag: strconv.Atoi: parsing \"garbage\": invalid syntax"
	require.EqualError(t, err, expectedError)
}

func TestStartReplicatorRetry_InvalidValueOutOfRangeError(t *testing.T) {
	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"start", "--replicator-retry-intervals=10,40,55555555555555555555555555555555555555555555555555555555555555555555555"})
	err := cmd.Execute()
	expectedError := "invalid argument \"10,40,55555555555555555555555555555555555555555555555555555555555555555555555\" " +
		"for \"--replicator-retry-intervals\" flag: strconv.Atoi: parsing " +
		"\"55555555555555555555555555555555555555555555555555555555555555555555555\": value out of range"
	require.EqualError(t, err, expectedError)
}
