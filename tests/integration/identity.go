// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"strconv"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/tests/state"
)

// NoIdentity returns an reference to an identity that represents no identity.
func NoIdentity() immutable.Option[state.Identity] {
	return immutable.None[state.Identity]()
}

// AllClientIdentities returns user identity selector specified with the "*".
func AllClientIdentities() immutable.Option[state.Identity] {
	return immutable.Some(
		state.Identity{
			Kind:     state.ClientIdentityType,
			Selector: "*",
		},
	)
}

// ClientIdentity returns a user identity at the given index.
func ClientIdentity(indexSelector int) immutable.Option[state.Identity] {
	return immutable.Some(
		state.Identity{
			Kind:     state.ClientIdentityType,
			Selector: strconv.Itoa(indexSelector),
		},
	)
}

// ClientIdentity returns a node identity at the given index.
func NodeIdentity(indexSelector int) immutable.Option[state.Identity] {
	return immutable.Some(
		state.Identity{
			Kind:     state.NodeIdentityType,
			Selector: strconv.Itoa(indexSelector),
		},
	)
}

// getIdentityForRequest returns the identity for the given reference and node index.
// It prepares the identity for a request by generating a token if needed, i.e. it will
// return an identity with [Identity.BearerToken] set.
func getIdentityForRequest(s *state.State, identity state.Identity, nodeIndex int) acpIdentity.Identity {
	identHolder := state.GetIdentityHolder(s, identity)
	ident := identHolder.Identity

	if fullIdent, ok := ident.(acpIdentity.FullIdentity); ok {
		token, ok := identHolder.NodeTokens[nodeIndex]
		if ok {
			fullIdent.SetBearerToken(token)
		} else {
			audience := getNodeAudience(s, nodeIndex)
			if documentACPType == SourceHubDocumentACPType || audience.HasValue() {
				err := fullIdent.UpdateToken(authTokenExpiration, audience, immutable.Some(s.SourcehubAddress))
				require.NoError(s.T, err)
				identHolder.NodeTokens[nodeIndex] = fullIdent.BearerToken()
			}
		}
	}
	return ident
}

// getIdentity returns an identity for the request specific to the node.
func getIdentityForRequestSpecificToNode(
	s *state.State,
	identity immutable.Option[state.Identity],
	nodeIndex int,
) immutable.Option[acpIdentity.Identity] {
	if !identity.HasValue() {
		return acpIdentity.None
	}
	return immutable.Some(getIdentityForRequest(s, identity.Value(), nodeIndex))
}

// getContextWithIdentity returns a context with the identity for the given reference and node index.
// If the identity does not exist, it will be generated.
// The identity added to the context is prepared for a request, i.e. its [Identity.BearerToken] is set.
func getContextWithIdentity(
	ctx context.Context,
	s *state.State,
	identity immutable.Option[state.Identity],
	nodeIndex int,
) context.Context {
	return acpIdentity.WithContext(ctx, getIdentityForRequestSpecificToNode(s, identity, nodeIndex))
}

func getIdentityDID(s *state.State, identity immutable.Option[state.Identity]) string {
	if identity.HasValue() {
		if identity.Value().Selector == "*" {
			return identity.Value().Selector
		}
		return state.GetIdentity(s, identity).DID()
	}
	return ""
}

// resetContextWithNoIdentity resets identity for the ctx to avoid, leaving it there and having the ctx
// reuse the same identity for other requests that don't specify an identity.
func resetStateContext(s *state.State) {
	s.Ctx = acpIdentity.WithContext(s.Ctx, acpIdentity.None)
}
