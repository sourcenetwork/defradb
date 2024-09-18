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

func TestPurgeCommandWithoutForceFlagReturnsError(t *testing.T) {
	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"client", "purge"})

	err := cmd.Execute()
	require.ErrorIs(t, err, ErrPurgeForceFlagRequired)
}
