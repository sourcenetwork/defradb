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
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

type identityType int

const (
	noType identityType = iota
	clientType
	nodeType
)

// identity is a type that refers to a specific identity of a certain type.
type identity struct {
	index int
	iType identityType
}

// NoIdentity returns an reference to an identity that represents no identity.
func NodeIdentity(index int) identity {
	return identity{index, nodeType}
}

// ClientIdentity returns a reference to a user identity with a given index.
func ClientIdentity(index int) identity {
	return identity{index, clientType}
}

// NodeIdentity returns a reference to a node identity with a given index.
func NoIdentity() identity {
	return identity{0, noType}
}

func (i identity) get(s *state) acpIdentity.Identity {
	var identities map[int]*identityHolder
	switch i.iType {
	case clientType:
		identities = s.clientIdentities
	case nodeType:
		identities = s.nodeIdentities
	default:
		return acpIdentity.Identity{}
	}
	return getIdentityHolder(s, i.index, identities).Identity
}

func (i identity) withToken(s *state, nodeID int) acpIdentity.Identity {
	var identities map[int]*identityHolder
	switch i.iType {
	case clientType:
		identities = s.clientIdentities
	case nodeType:
		identities = s.nodeIdentities
	default:
		return acpIdentity.Identity{}
	}
	return getIdentityHolderWithToken(s, i.index, nodeID, identities).Identity
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

func getIdentityHolder(s *state, index int, identities map[int]*identityHolder) *identityHolder {
	_, ok := identities[index]
	if !ok {
		identities[index] = newIdentityHolder(generateIdentity(s))
	}
	return identities[index]
}

func getIdentityHolderWithToken(s *state, index, nodeID int, identities map[int]*identityHolder) *identityHolder {
	ident := getIdentityHolder(s, index, identities)
	token, ok := ident.NodeTokens[nodeID]
	if ok {
		ident.Identity.BearerToken = token
	} else {
		ident.NodeTokens[nodeID] = generateToken(s, &ident.Identity, nodeID)
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

func generateToken(s *state, ident *acpIdentity.Identity, nodeID int) string {
	audience := getNodeAudience(s, nodeID)
	if acpType == SourceHubACPType || audience.HasValue() {
		err := ident.UpdateToken(
			authTokenExpiration,
			audience,
			immutable.Some(s.sourcehubAddress),
		)
		require.NoError(s.t, err)
		return ident.BearerToken
	}
	return ""
}

// getContextWithIdentity returns a context with the identity for the given reference and node index.
// If the identity does not exist, it will be generated.
// The identity added to the context is prepared for a request, i.e. its [Identity.BearerToken] is set.
func getContextWithIdentity(ctx context.Context, s *state, ref identity, nodeIndex int) context.Context {
	if ref.iType == noType {
		return ctx
	}
	ident := ref.withToken(s, nodeIndex)
	return acpIdentity.WithContext(ctx, immutable.Some(ident))
}
