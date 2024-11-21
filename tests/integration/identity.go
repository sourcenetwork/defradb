// Copyright 2024 Democratized Data Foundation
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
	"math/rand"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

type identityType int

const (
	clientIdentityType identityType = iota
	nodeIdentityType
)

// identity helps specify identity type info and selector/index of identity to use in a test case.
type identity struct {
	// type of identity
	kind identityType

	// selector can be a valid identity index or a selecting pattern like "*".
	// Note: "*" means to select all identities of the specified [kind] type.
	selector string
}

// NoIdentity returns an reference to an identity that represents no identity.
func NoIdentity() immutable.Option[identity] {
	return immutable.None[identity]()
}

// AllClientIdentities returns user identity selector specified with the "*".
func AllClientIdentities() immutable.Option[identity] {
	return immutable.Some(
		identity{
			kind:     clientIdentityType,
			selector: "*",
		},
	)
}

// ClientIdentity returns a user identity at the given index.
func ClientIdentity(indexSelector int) immutable.Option[identity] {
	return immutable.Some(
		identity{
			kind:     clientIdentityType,
			selector: strconv.Itoa(indexSelector),
		},
	)
}

// ClientIdentity returns a node identity at the given index.
func NodeIdentity(indexSelector int) immutable.Option[identity] {
	return immutable.Some(
		identity{
			kind:     nodeIdentityType,
			selector: strconv.Itoa(indexSelector),
		},
	)
}

// identityHolder holds an identity and the generated tokens for each target node.
// This is used to cache the generated tokens for each node.
type identityHolder struct {
	// Identity is the identity.
	Identity acpIdentity.Identity
	// NodeTokens is a map of node index to the generated token for that node.
	NodeTokens map[int]string
}

func newIdentityHolder(ident acpIdentity.Identity) *identityHolder {
	return &identityHolder{
		Identity:   ident,
		NodeTokens: make(map[int]string),
	}
}

// getIdentity returns the identity for the given reference.
// If the identity does not exist, it will be generated.
func getIdentity(s *state, identity immutable.Option[identity]) acpIdentity.Identity {
	if !identity.HasValue() {
		return acpIdentity.Identity{}
	}
	if identity.Value().selector == "*" {
		require.Fail(s.t, "Used the \"*\" selector for identity incorrectly.", s.testCase.Description)
	}
	return getIdentityHolder(s, identity.Value()).Identity
}

// getIdentityHolder returns the identity holder for the given reference.
// If the identity does not exist, it will be generated.
func getIdentityHolder(s *state, identity identity) *identityHolder {
	ident, ok := s.identities[identity]
	if ok {
		return ident
	}

	s.identities[identity] = newIdentityHolder(generateIdentity(s))
	return s.identities[identity]
}

// getIdentityForRequest returns the identity for the given reference and node index.
// It prepares the identity for a request by generating a token if needed, i.e. it will
// return an identity with [Identity.BearerToken] set.
func getIdentityForRequest(s *state, identity identity, nodeIndex int) acpIdentity.Identity {
	identHolder := getIdentityHolder(s, identity)
	ident := identHolder.Identity

	token, ok := identHolder.NodeTokens[nodeIndex]
	if ok {
		ident.BearerToken = token
	} else {
		audience := getNodeAudience(s, nodeIndex)
		if acpType == SourceHubACPType || audience.HasValue() {
			err := ident.UpdateToken(authTokenExpiration, audience, immutable.Some(s.sourcehubAddress))
			require.NoError(s.t, err)
			identHolder.NodeTokens[nodeIndex] = ident.BearerToken
		}
	}
	return ident
}

// Generate the keys using predefined seed so that multiple runs yield the same private key.
// This is important for stuff like the change detector.
func generateIdentity(s *state) acpIdentity.Identity {
	source := rand.NewSource(int64(s.nextIdentityGenSeed))
	r := rand.New(source)

	privateKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
	require.NoError(s.t, err)

	s.nextIdentityGenSeed++

	identity, err := acpIdentity.FromPrivateKey(privateKey)
	require.NoError(s.t, err)

	return identity
}

// getContextWithIdentity returns a context with the identity for the given reference and node index.
// If the identity does not exist, it will be generated.
// The identity added to the context is prepared for a request, i.e. its [Identity.BearerToken] is set.
func getContextWithIdentity(
	ctx context.Context,
	s *state,
	identity immutable.Option[identity],
	nodeIndex int,
) context.Context {
	if !identity.HasValue() {
		return ctx
	}
	return acpIdentity.WithContext(
		ctx,
		immutable.Some(
			getIdentityForRequest(
				s,
				identity.Value(),
				nodeIndex,
			),
		),
	)
}

func getIdentityDID(s *state, identity immutable.Option[identity]) string {
	if identity.HasValue() {
		if identity.Value().selector == "*" {
			return identity.Value().selector
		}
		return getIdentity(s, identity).DID
	}
	return ""
}
