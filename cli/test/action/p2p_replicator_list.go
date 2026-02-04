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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// P2PReplicatorList executes the `client p2p replicator getall` command.
type P2PReplicatorList struct {
	stateful
	augmented

	// The expected list of replicators.
	// If provided, it will be compared with the actual result.
	Expected immutable.Option[[]client.Replicator]

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*P2PReplicatorList)(nil)

func (a *P2PReplicatorList) Execute() {
	args := []string{"client", "p2p", "replicator", "list"}

	args = a.AppendDirections(args)

	result, err := executeJson[[]client.Replicator](a.s.Ctx, args)

	if a.ExpectError != "" {
		require.Error(a.s.T, err)
		require.Contains(a.s.T, err.Error(), a.ExpectError)
		return
	}

	require.NoError(a.s.T, err)

	if a.Expected.HasValue() {
		expected := a.Expected.Value()
		require.Equal(a.s.T, len(expected), len(result))
		for i, exp := range expected {
			act := result[i]
			require.Equal(a.s.T, exp.ID, act.ID)
			require.Equal(a.s.T, exp.Addresses, act.Addresses)
			require.Equal(a.s.T, exp.CollectionIDs, act.CollectionIDs)
		}
	}
}
