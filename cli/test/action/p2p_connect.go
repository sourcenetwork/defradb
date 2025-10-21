// Copyright 2025 Democratized Data Foundation
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

// P2PConnect executes the `client p2p connect` command.
type P2PConnect struct {
	stateful
	augmented

	// The addresses to connect to (required).
	Addresses []string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*P2PConnect)(nil)

func (a *P2PConnect) Execute() {
	args := []string{"client", "p2p", "connect"}

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
