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
	"github.com/sourcenetwork/defradb/crypto"
)

// Generate generates a new identity.
func Generate() (RawIdentity, error) {
	privateKey, err := crypto.GenerateSecp256k1()
	if err != nil {
		return RawIdentity{}, err
	}

	publicKey := crypto.NewPublicKey(privateKey.PubKey())

	did, err := publicKey.DID()
	if err != nil {
		return RawIdentity{}, err
	}

	return RawIdentity{
		PrivateKey: crypto.NewPrivateKey(privateKey).String(),
		PublicKey:  publicKey.String(),
		DID:        did,
	}, nil
}
