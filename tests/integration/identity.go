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
	"crypto/ed25519"
	"math/rand"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

type identityType int

const (
	clientIdentityType identityType = iota
	nodeIdentityType
)

// Identity helps specify Identity type info and selector/index of Identity to use in a test case.
type Identity struct {
	// type of identity
	kind identityType

	// selector can be a valid identity index or a selecting pattern like "*".
	// Note: "*" means to select all identities of the specified [kind] type.
	selector string
}

// NoIdentity returns an reference to an identity that represents no identity.
func NoIdentity() immutable.Option[Identity] {
	return immutable.None[Identity]()
}

// AllClientIdentities returns user identity selector specified with the "*".
func AllClientIdentities() immutable.Option[Identity] {
	return immutable.Some(
		Identity{
			kind:     clientIdentityType,
			selector: "*",
		},
	)
}

// ClientIdentity returns a user identity at the given index.
func ClientIdentity(indexSelector int) immutable.Option[Identity] {
	return immutable.Some(
		Identity{
			kind:     clientIdentityType,
			selector: strconv.Itoa(indexSelector),
		},
	)
}

// ClientIdentity returns a node identity at the given index.
func NodeIdentity(indexSelector int) immutable.Option[Identity] {
	return immutable.Some(
		Identity{
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
func getIdentity(s *state, identity immutable.Option[Identity]) acpIdentity.Identity {
	if !identity.HasValue() {
		return nil
	}

	// The selector must never be "*" here because this function returns a specific identity from the
	// stored identities, if "*" string needs to be signaled to the acp module then it should be handled
	// a call before this function.
	if identity.Value().selector == "*" {
		require.Fail(s.t, "Used the \"*\" selector for identity incorrectly.", s.testCase.Description)
	}
	return getIdentityHolder(s, identity.Value()).Identity
}

// getIdentityHolder returns the identity holder for the given reference.
// If the identity does not exist, it will be generated.
func getIdentityHolder(s *state, identity Identity) *identityHolder {
	ident, ok := s.identities[identity]
	if ok {
		return ident
	}

	keyType := crypto.KeyTypeSecp256k1
	if k, ok := s.testCase.IdentityTypes[identity]; ok {
		keyType = k
	}

	s.identities[identity] = newIdentityHolder(generateIdentity(s, keyType))
	return s.identities[identity]
}

// getIdentityForRequest returns the identity for the given reference and node index.
// It prepares the identity for a request by generating a token if needed, i.e. it will
// return an identity with [Identity.BearerToken] set.
func getIdentityForRequest(s *state, identity Identity, nodeIndex int) acpIdentity.Identity {
	identHolder := getIdentityHolder(s, identity)
	ident := identHolder.Identity

	if fullIdent, ok := ident.(acpIdentity.FullIdentity); ok {
		token, ok := identHolder.NodeTokens[nodeIndex]
		if ok {
			fullIdent.SetBearerToken(token)
		} else {
			audience := getNodeAudience(s, nodeIndex)
			if documentACPType == SourceHubDocumentACPType || audience.HasValue() {
				err := fullIdent.UpdateToken(authTokenExpiration, audience, immutable.Some(s.sourcehubAddress))
				require.NoError(s.t, err)
				identHolder.NodeTokens[nodeIndex] = fullIdent.BearerToken()
			}
		}
	}
	return ident
}

// Generate the keys using predefined seed so that multiple runs yield the same private key.
// This is important for stuff like the change detector.
func generateIdentity(s *state, keyType crypto.KeyType) acpIdentity.Identity {
	source := rand.NewSource(int64(s.nextIdentityGenSeed))
	r := rand.New(source)

	var privateKey crypto.PrivateKey
	if keyType == crypto.KeyTypeSecp256k1 {
		privKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
		require.NoError(s.t, err)
		privateKey = crypto.NewPrivateKey(privKey)
	} else if keyType == crypto.KeyTypeEd25519 {
		_, privKey, err := ed25519.GenerateKey(r)
		require.NoError(s.t, err)
		privateKey = crypto.NewPrivateKey(privKey)
	} else {
		require.Fail(s.t, "Unsupported signing algorithm", s.testCase.Description)
	}

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
	identity immutable.Option[Identity],
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

func getIdentityDID(s *state, identity immutable.Option[Identity]) string {
	if identity.HasValue() {
		if identity.Value().selector == "*" {
			return identity.Value().selector
		}
		return getIdentity(s, identity).DID()
	}
	return ""
}
