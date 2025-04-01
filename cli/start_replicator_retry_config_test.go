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

	"github.com/stretchr/testify/require"
)

func TestStartReplicatorRetry_NoError(t *testing.T) {
	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"start", "--no-keyring", "--replicator-retry-intervals=10,20,40"})
	err := cmd.Execute()
	require.NoError(t, err)
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
