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

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	identity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

// identRef is a type that refers to a specific identity of a certain type.
type identRef struct {
	hasValue bool
	isUser   bool
	index    int
}

// NoIdentity returns an reference to an identity that represents no identity.
func NoIdentity() identRef {
	return identRef{
		hasValue: false,
	}
}

// UserIdentity returns a reference to a user identity with a given index.
func UserIdentity(index int) identRef {
	return identRef{
		hasValue: true,
		isUser:   true,
		index:    index,
	}
}

// NodeIdentity returns a reference to a node identity with a given index.
func NodeIdentity(index int) identRef {
	return identRef{
		hasValue: true,
		isUser:   false,
		index:    index,
	}
}

// identityHolder holds an identity and the generated tokens for each node.
// This is used to cache the generated tokens for each node.
type identityHolder struct {
	// Identity is the identity.
	Identity identity.Identity
	// NodeTokens is a map of node index to the generated token for that node.
	NodeTokens map[int]string
}

func newIdentityHolder(ident identity.Identity) *identityHolder {
	return &identityHolder{
		Identity:   ident,
		NodeTokens: make(map[int]string),
	}
}

// getIdentity returns the identity for the given reference.
// If the identity does not exist, it will be generated.
func getIdentity(s *state, ref identRef) acpIdentity.Identity {
	return getIdentityHolder(s, ref).Identity
}

// getIdentityHolder returns the identity holder for the given reference.
// If the identity does not exist, it will be generated.
func getIdentityHolder(s *state, ref identRef) *identityHolder {
	ident, ok := s.identities[ref]
	if ok {
		return ident
	}

	s.identities[ref] = newIdentityHolder(generateIdentity(s))
	return s.identities[ref]
}

// getIdentityForRequest returns the identity for the given reference and node index.
// It prepares the identity for a request by generating a token if needed, i.e. it will
// return an identity with [Identity.BearerToken] set.
func getIdentityForRequest(s *state, ref identRef, nodeIndex int) acpIdentity.Identity {
	identHolder := getIdentityHolder(s, ref)
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
func getContextWithIdentity(s *state, ref identRef, nodeIndex int) context.Context {
	if !ref.hasValue {
		return s.ctx
	}
	ident := getIdentityForRequest(s, ref, nodeIndex)
	return identity.WithContext(s.ctx, immutable.Some(ident))
}

func getIdentityDID(s *state, ident identRef) string {
	if ident.hasValue {
		return getIdentity(s, ident).DID
	}
	return ""
}
