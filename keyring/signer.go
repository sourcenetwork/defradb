// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keyring

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/sourcenetwork/sourcehub/sdk"
)

type txnSigner struct {
	keyring Keyring
	keyName string

	// accAddress is public info and can be safely cached
	accAddress string
}

var _ sdk.TxSigner = (*txnSigner)(nil)

// NewTxSignerFromKeyringKey creates a new TxSigner backed by a keyring.
//
// The key used for signing is not cached and will be fetched from the keyring every time it
// is requested.  This minimizes the risk of it being leaked via stuff like memory paging.
func NewTxSignerFromKeyringKey(keyring Keyring, keyName string) (*txnSigner, error) {
	bytes, err := keyring.Get(keyName)
	if err != nil {
		return nil, err
	}

	key := &secp256k1.PrivKey{
		Key: bytes,
	}
	addr := key.PubKey().Address().Bytes()
	accAddress := cosmostypes.AccAddress(addr).String()

	return &txnSigner{
		keyring:    keyring,
		keyName:    keyName,
		accAddress: accAddress,
	}, nil
}

func (s *txnSigner) GetAccAddress() string {
	return s.accAddress
}

func (s *txnSigner) GetPrivateKey() cryptotypes.PrivKey {
	bytes, err := s.keyring.Get(s.keyName)
	if err != nil {
		panic(err)
	}

	return &secp256k1.PrivKey{
		Key: bytes,
	}
}
