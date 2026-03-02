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

// DeleteP2PReplicator executes the `client p2p replicator delete` command.
type DeleteP2PReplicator struct {
	stateful
	augmented

	// The peer to delete replication from  (required).
	PeerID string

	// The collections to delete from the given peer (optional).
	Collections []string

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*DeleteP2PReplicator)(nil)

func (a *DeleteP2PReplicator) Execute() {
	args := []string{"client", "p2p", "replicator", "delete"}

	if a.Collections != nil {
		args = append(args, "-c")
		args = append(args, strings.Join(a.Collections, ","))
	}

	if a.PeerID != "" {
		args = append(args, a.PeerID)
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
