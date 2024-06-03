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
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
)

// CheckAccessOfDocOnCollectionWithACP handles the check, which tells us if access to the target
// document is valid, with respect to the permission type, and the specified collection.
//
// This function should only be called if acp is available. As we have unrestricted
// access when acp is not available (acp turned off).
//
// Since we know acp is enabled we have these components to check in this function:
// (1) the request is permissioned (has an identity),
// (2) the collection is permissioned (has a policy),
//
// Unrestricted Access to document if:
// - (2) is false.
// - Document is public (unregistered), whether signatured request or not doesn't matter.
func CheckAccessOfDocOnCollectionWithACP(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	acpSystem acp.ACP,
	collection client.Collection,
	permission acp.DPIPermission,
	docID string,
) (bool, error) {
	// Even if acp exists, but there is no policy on the collection (unpermissioned collection)
	// then we still have unrestricted access.
	policyID, resourceName, hasPolicy := isPermissioned(collection)
	if !hasPolicy {
		return true, nil
	}

	// Now that we know acp is available and the collection is permissioned, before checking access with
	// acp directly we need to make sure that the document is not public, as public documents will not
	// be regestered with acp. We give unrestricted access to public documents, so it does not matter
	// whether the request has a signature identity or not at this stage of the check.
	isRegistered, err := acpSystem.IsDocRegistered(
		ctx,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, err
	}

	if !isRegistered {
		// Unrestricted access as it is a public document.
		return true, nil
	}

	// At this point if the request is not signatured, then it has no access, because:
	// the collection has a policy on it, and the acp is enabled/available,
	// and the document is not public (is regestered with acp).
	if !identity.HasValue() {
		return false, nil
	}

	// Now actually check using the signature if this identity has access or not.
	hasAccess, err := acpSystem.CheckDocAccess(
		ctx,
		permission,
		identity.Value().Address,
		policyID,
		resourceName,
		docID,
	)

	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
