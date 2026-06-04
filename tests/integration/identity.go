// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package tests

import (
	"strconv"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/tests/action"
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

// getIdentityOption returns the identity similar to [getIdentity] but in immutable.Option.
func getIdentityOption(
	s *state.State,
	identity immutable.Option[state.Identity],
) immutable.Option[acpIdentity.Identity] {
	ident := state.GetIdentity(s, identity)
	if ident == nil {
		return acpIdentity.None
	}
	return immutable.Some(ident)
}

// getIdentityForRequest returns the identity for the given reference and node index.
// It prepares the identity for a request by generating a token if needed, i.e. it will
// return an identity with [Identity.BearerToken] set.
func getIdentityForRequest(s *state.State, identity state.Identity, nodeIndex int) acpIdentity.Identity {
	identHolder := state.GetIdentityHolder(s, identity)
	ident := identHolder.Identity

	if fullIdent, ok := ident.(acpIdentity.FullIdentity); ok {
		audience := state.GetNodeAudience(s, nodeIndex)
		token, ok := identHolder.NodeTokens[nodeIndex]
		if ok {
			fullIdent.SetBearerToken(token)
		}

		// Generate/regenerate the token if:
		// - No token exists yet, OR
		// - An audience is now available but the token was generated without one
		//   (this can happen when the token is created during node setup before the
		//    HTTP wrapper is ready, causing the audience to be unavailable at that time).
		if !ok || (audience.HasValue() && !state.TokenHasAudience(token)) {
			if s.DocumentACPType == state.SourceHubDocumentACPType || audience.HasValue() {
				err := fullIdent.UpdateToken(
					action.AuthTokenExpiration,
					audience,
					immutable.Some(s.SourcehubAddress),
				)
				require.NoError(s.T, err)
				identHolder.NodeTokens[nodeIndex] = fullIdent.BearerToken()
			}
		}
	}
	return ident
}

// getIdentityForRequestSpecificToNode returns an identity for the request specific to the node.
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

func getIdentityDID(s *state.State, identity immutable.Option[state.Identity]) string {
	if identity.HasValue() {
		if identity.Value().Selector == "*" {
			return identity.Value().Selector
		}
		return state.GetIdentity(s, identity).DID()
	}
	return ""
}
