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
	"strings"

	"github.com/stretchr/testify/require"
)

// P2PReplicatorAdd executes the `client p2p replicator add` command.
type P2PReplicatorAdd struct {
	stateful
	augmented

	// The addresses to connect to (required).
	Addresses []string

	// The collections to replicate to the given addresses (optional).
	Collections []string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*P2PReplicatorAdd)(nil)

func (a *P2PReplicatorAdd) Execute() {
	args := []string{"client", "p2p", "replicator", "add"}

	if a.Collections != nil {
		args = append(args, "-c")
		args = append(args, strings.Join(a.Collections, ","))
	}

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
