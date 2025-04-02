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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStartReplicatorRetry_NoError(t *testing.T) {
	cmd := NewDefraCommand()
	args := []string{
		"start",
		"--no-keyring",
		"--url=127.0.0.1:",
		"--acp-type=local",
		"--replicator-retry-intervals=10,20,40",
	}
	cmd.SetArgs(args)

	// We do not expect the start command to return an error. So we will start
	// and wait 10 seconds. If it does not return, then we are good
	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	select {
	case <-time.After(10 * time.Second):
		// Pass
	case <-done:
		t.Fail() //Fail the test if the command returns an error
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
