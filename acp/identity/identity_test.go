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

func mockedDIDProducer(crypto.KeyType, []byte) (*key.DIDKey, error) {
	return nil, fmt.Errorf("some did generation error")
}

// newFailableIdentityProvider returns an identityProvider that always fails
func newFailableIdentityProvider() identityProvider {
	return identityProvider{
		producer: mockedDIDProducer,
	}
}

func Test_DIDFromPublicKey_ProducesDIDForPublicKey(t *testing.T) {
	pubKey := &secp256k1.PublicKey{}

	did, err := DIDFromPublicKey(pubKey)

	want := "did:key:z7r8ooUiNXK8TT8Xjg1EWStR2ZdfxbzVfvGWbA2FjmzcnmDxz71QkP1Er8PP3zyLZpBLVgaXbZPGJPS4ppXJDPRcqrx4F"
	require.Equal(t, want, did)
	require.NoError(t, err)
}

func Test_DIDFromPublicKey_ReturnsErrorWhenProducerFails(t *testing.T) {
	pubKey := &secp256k1.PublicKey{}
	did, err := generateDID(pubKey, mockedDIDProducer)

	require.Empty(t, did)
	require.ErrorIs(t, err, ErrDIDCreation)
}

func Test_FromPublicKey_ProducerFailureCausesError(t *testing.T) {
	pubKey := &secp256k1.PublicKey{}
	provider := newFailableIdentityProvider()
	identity, err := provider.FromPublicKey(pubKey)

	require.Equal(t, None, identity)
	require.ErrorIs(t, err, ErrDIDCreation)
}

func Test_FromPrivateKey_ProducerFailureCausesError(t *testing.T) {
	key := &secp256k1.PrivateKey{}
	provider := newFailableIdentityProvider()
	identity, err := provider.FromPrivateKey(key)

	require.Equal(t, None, identity)
	require.ErrorIs(t, err, ErrDIDCreation)
}
