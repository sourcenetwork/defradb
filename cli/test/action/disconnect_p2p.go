// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/stretchr/testify/require"
)

// DisconnectP2P executes the `client p2p disconnect` command.
type DisconnectP2P struct {
	stateful
	augmented

	// The addresses to disconnect from (required).
	Addresses []string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*DisconnectP2P)(nil)

func (a *DisconnectP2P) Execute() {
	args := []string{"client", "p2p", "disconnect"}

	if a.Addresses != nil {
		args = append(args, a.Addresses...)
	}

	args = a.AppendDirections(args)

	err := execute(a.s.Ctx, args)

	if a.ExpectError != "" {
		require.Error(a.s.T, err)
		require.Contains(a.s.T, err.Error(), a.ExpectError)
		return
	}

	require.NoError(a.s.T, err)
}
