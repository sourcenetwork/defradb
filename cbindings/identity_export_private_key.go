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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

//export ExportIdentityPrivateKey
func ExportIdentityPrivateKey(identityPtr C.uintptr_t) C.Result {
	ident, err := getIdentityFromPointer(identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if ident == nil {
		return returnC(returnGoC(1, errIdentiityIsNil, ""))
	}

	fullIdent, ok := ident.(acpIdentity.FullIdentity)
	if !ok {
		return returnC(returnGoC(1, errIdentityDoesNotContainPrivateKey, ""))
	}

	privKeyBytes := fullIdent.PrivateKey().Raw()

	return returnC(GoCResult{Status: 0, Value: hex.EncodeToString(privKeyBytes)})
}
