// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

var (
	Actor1Identity = MustParseIdentity("028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f")
	Actor2Identity = MustParseIdentity("4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5")
)

// MustParseIdentity returns an identity that uses the given private key or panics.
func MustParseIdentity(privateKeyHex string) immutable.Option[acpIdentity.Identity] {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		panic(err)
	}
	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	identity, err := acpIdentity.FromPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	return identity
}
