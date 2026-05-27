// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

//export NewIdentityFromPrivateKey
func NewIdentityFromPrivateKey(privateKeyHex *C.char) C.NewIdentityResult {
	data, err := hex.DecodeString(C.GoString(privateKeyHex))
	if err != nil {
		return returnNewIdentityResultC(1, err.Error(), nil)
	}

	if len(data) != secp256k1.PrivKeyBytesLen {
		return returnNewIdentityResultC(1, errInvalidPrivateKeyLength, nil)
	}
	privKey := secp256k1.PrivKeyFromBytes(data)
	ident, err := identity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	if err != nil {
		return returnNewIdentityResultC(1, err.Error(), nil)
	}

	return returnNewIdentityResultC(0, "", ident)
}
