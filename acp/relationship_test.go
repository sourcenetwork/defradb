// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp_test

import (
	"context"
	"testing"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	coreTypes "github.com/sourcenetwork/acp_core/pkg/types"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/acp/local"
	"github.com/sourcenetwork/defradb/crypto"
)

func TestNewActorRelationship(t *testing.T) {
	target, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	for _, test := range []struct {
		name     string
		target   string
		expected *coreTypes.Relationship
		err      error
	}{
		{
			name: "empty target",
			err:  acp.ErrRelationshipTargetActorEmpty,
		},
		{
			name:     "wildcard target",
			target:   "*",
			expected: coreTypes.NewAllActorsRelationship("resource", "object", "reader"),
		},
		{
			name:     "explicit target",
			target:   target.DID(),
			expected: coreTypes.NewActorRelationship("resource", "object", "reader", target.DID()),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, acp.ValidateActorRelationshipTarget(test.target), test.err)
			relationship, err := acp.NewActorRelationship("resource", "object", "reader", test.target)
			require.ErrorIs(t, err, test.err)
			require.Equal(t, test.expected, relationship)
		})
	}
}

func TestActorRelationshipEmptyTarget(t *testing.T) {
	requester, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	// Uninitialized adapters ensure invalid targets are rejected before any engine,
	// transaction, or JavaScript bridge call, without needing an external ACP service.
	for _, test := range []struct {
		name   string
		client acp.ACPSystemClient
	}{
		{name: "local", client: &local.LocalACP{}},
		{name: "remote", client: &dac.RemoteDocumentACP{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, operation := range []struct {
				name string
				call func(
					context.Context, string, string, string, string, identity.Identity, string, *protoTypes.Timestamp,
				) (bool, error)
			}{
				{name: "add", call: test.client.AddActorRelationship},
				{name: "delete", call: test.client.DeleteActorRelationship},
			} {
				t.Run(operation.name, func(t *testing.T) {
					result, err := operation.call(
						context.Background(), "policy", "resource", "object", "reader", requester, "", nil,
					)
					require.ErrorIs(t, err, acp.ErrRelationshipTargetActorEmpty)
					require.False(t, result)
				})
			}
		})
	}
}
