// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package identity

import (
	"fmt"
	"testing"

	"github.com/cyware/ssi-sdk/crypto"
	"github.com/cyware/ssi-sdk/did/key"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
)

func Test_DIDFromPublicKey_ProducesDIDForPublicKey(t *testing.T) {
	pubKey := &secp256k1.PublicKey{}

	did, err := DIDFromPublicKey(pubKey)

	want := "did:key:z7r8ooUiNXK8TT8Xjg1EWStR2ZdfxbzVfvGWbA2FjmzcnmDxz71QkP1Er8PP3zyLZpBLVgaXbZPGJPS4ppXJDPRcqrx4F"
	require.Equal(t, want, did)
	require.NoError(t, err)
}

func Test_DIDFromPublicKey_ReturnsErrorWhenProducerFails(t *testing.T) {
	mockedProducer := func(crypto.KeyType, []byte) (*key.DIDKey, error) {
		return nil, fmt.Errorf("did generation err")
	}

	pubKey := &secp256k1.PublicKey{}

	did, err := didFromPublicKey(pubKey, mockedProducer)

	require.Empty(t, did)
	require.ErrorIs(t, err, ErrDIDCreation)
}

func Test_RawIdentityGeneration_ReturnsNewRawIdentity(t *testing.T) {
	newIdentity, err := Generate()
	require.NoError(t, err)

	// Check that both private and public key are not empty.
	require.NotEmpty(t, newIdentity.PrivateKey)
	require.NotEmpty(t, newIdentity.PublicKey)

	// Check leading `did:key` prefix.
	require.Equal(t, newIdentity.DID[:7], "did:key")
}

func Test_RawIdentityGenerationIsNotFixed_ReturnsUniqueRawIdentites(t *testing.T) {
	newIdentity1, err1 := Generate()
	newIdentity2, err2 := Generate()
	require.NoError(t, err1)
	require.NoError(t, err2)

	// Check that both private and public key are not empty.
	require.NotEmpty(t, newIdentity1.PrivateKey)
	require.NotEmpty(t, newIdentity1.PublicKey)
	require.NotEmpty(t, newIdentity2.PrivateKey)
	require.NotEmpty(t, newIdentity2.PublicKey)

	// Check leading `did:key` prefix.
	require.Equal(t, newIdentity1.DID[:7], "did:key")
	require.Equal(t, newIdentity2.DID[:7], "did:key")

	// Check both are different.
	require.NotEqual(t, newIdentity1.PrivateKey, newIdentity2.PrivateKey)
	require.NotEqual(t, newIdentity1.PublicKey, newIdentity2.PublicKey)
	require.NotEqual(t, newIdentity1.DID, newIdentity2.DID)
}
