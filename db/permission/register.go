// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package acp utilizes the sourcehub acp module to bring the functionality to defradb, this package also helps
avoid the leakage of direct sourcehub references through out the code base, and eases in swapping
between local embedded use case and a more global on sourcehub use case.
*/

package permission

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
)

// The document is only registered with ACP if all (1) (2) and (3) are true.
// In all other cases, nothing is registered with ACP.

// RegisterDocCreationOnCollection handles the registration of the document with acp module.
// The registering is done at document creation on the collection.
//
// According to our access logic we have these components to worry about:
// (1) the request is permissioned (has an identity signature),
// (2) the collection is permissioned (has a policy),
// (3) acp module exists (acp is enabled).
//
// The document is only registered if all (1) (2) and (3) are true.
//
// Otherwise, nothing is registered on the acp module.
func RegisterDocCreationOnCollection(
	ctx context.Context,
	acpModule immutable.Option[acp.ACPModule],
	collection client.Collection,
	docID string,
) error {
	// If acp module is enabled / exists.
	if acpModule.HasValue() {
		// And collection has policy.
		if policyID, resourceName, hasPolicy := IsPermissioned(collection); hasPolicy {
			return acpModule.Value().RegisterDocCreation(
				ctx,
				"cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969", // TODO-ACP: Replace with signature identity
				policyID,
				resourceName,
				docID,
			)
		}
	}

	return nil
}
