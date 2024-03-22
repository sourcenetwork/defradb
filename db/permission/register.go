// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package permission

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
)

// RegisterDocOnCollectionWithACP handles the registration of the document with acp module.
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
func RegisterDocOnCollectionWithACP(
	ctx context.Context,
	identity immutable.Option[string],
	acpModule immutable.Option[acp.ACPModule],
	collection client.Collection,
	docID string,
) error {
	// If acp module is enabled / exists.
	if acpModule.HasValue() && identity.HasValue() {
		// And collection has policy.
		if policyID, resourceName, hasPolicy := isPermissioned(collection); hasPolicy {
			return acpModule.Value().RegisterDocObject(
				ctx,
				identity.Value(),
				policyID,
				resourceName,
				docID,
			)
		}
	}

	return nil
}
