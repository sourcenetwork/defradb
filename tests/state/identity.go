// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package state

import (
	"crypto/ed25519"
	"math/rand"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

type IdentityType int

const (
	ClientIdentityType IdentityType = iota
	NodeIdentityType
)

// Identity helps specify Identity type info and selector/index of Identity to use in a test case.
type Identity struct {
	// type of identity
	Kind IdentityType

	// Selector can be a valid identity index or a selecting pattern like "*".
	// Note: "*" means to select all identities of the specified [kind] type.
	Selector string
}

// IdentityHolder holds an identity and the generated tokens for each target node.
// This is used to cache the generated tokens for each node.
type IdentityHolder struct {
	// Identity is the identity.
	Identity acpIdentity.Identity
	// NodeTokens is a map of node index to the generated token for that node.
	NodeTokens map[int]string
}

func newIdentityHolder(ident acpIdentity.Identity) *IdentityHolder {
	return &IdentityHolder{
		Identity:   ident,
		NodeTokens: make(map[int]string),
	}
}

// GetIdentity returns the identity for the given reference.
// If the identity does not exist, it will be generated.
func GetIdentity(s *State, identity immutable.Option[Identity]) acpIdentity.Identity {
	if !identity.HasValue() {
		return nil
	}

	// The selector must never be "*" here because this function returns a specific identity from the
	// stored identities, if "*" string needs to be signaled to the acp module then it should be handled
	// a call before this function.
	if identity.Value().Selector == "*" {
		require.Fail(s.T, "Used the \"*\" selector for identity incorrectly.")
	}
	return GetIdentityHolder(s, identity.Value()).Identity
}

// GetIdentityHolder returns the identity holder for the given reference.
// If the identity does not exist, it will be generated.
func GetIdentityHolder(s *State, identity Identity) *IdentityHolder {
	ident, ok := s.Identities[identity]
	if ok {
		return ident
	}

	keyType := crypto.KeyTypeSecp256k1
	if k, ok := s.IdentityTypes[identity]; ok {
		keyType = k
	}

	s.Identities[identity] = newIdentityHolder(generateIdentity(s, keyType))
	return s.Identities[identity]
}

// Generate the keys using predefined seed so that multiple runs yield the same private key.
// This is important for stuff like the change detector.
func generateIdentity(s *State, keyType crypto.KeyType) acpIdentity.Identity {
	source := rand.NewSource(int64(s.NextIdentityGenSeed))
	r := rand.New(source)

	var privateKey crypto.PrivateKey
	if keyType == crypto.KeyTypeSecp256k1 {
		privKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
		require.NoError(s.T, err)
		privateKey = crypto.NewPrivateKey(privKey)
	} else if keyType == crypto.KeyTypeEd25519 {
		_, privKey, err := ed25519.GenerateKey(r)
		require.NoError(s.T, err)
		privateKey = crypto.NewPrivateKey(privKey)
	} else {
		require.Fail(s.T, "Unsupported signing algorithm")
	}

	s.NextIdentityGenSeed++

	identity, err := acpIdentity.FromPrivateKey(privateKey)
	require.NoError(s.T, err)

	return identity
}
