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

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
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
	documentACP dac.DocumentACP,
	collection client.Collection,
	permission acpTypes.ResourceInterfacePermission,
	docID string,
) (bool, error) {
	identityFunc := func() immutable.Option[acpIdentity.Identity] {
		return identity
	}
	return CheckDocAccessWithIdentityFunc(
		ctx,
		identityFunc,
		documentACP,
		collection,
		permission,
		docID,
	)
}

// CheckDocAccessWithIdentityFunc handles the check, which tells us if access to the target
// document is valid, with respect to the permission type, and the specified collection.
//
// The identity is determined by an identity function.
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
func CheckDocAccessWithIdentityFunc(
	ctx context.Context,
	identityFunc func() immutable.Option[acpIdentity.Identity],
	documentACP dac.DocumentACP,
	collection client.Collection,
	permission acpTypes.ResourceInterfacePermission,
	docID string,
) (bool, error) {
	// Even if acp exists, but there is no policy on the collection (unpermissioned collection)
	// then we still have unrestricted access.
	policyID, resourceName, hasPolicy := IsPermissioned(collection)
	if !hasPolicy {
		return true, nil
	}

	// Now that we know acp is available and the collection is permissioned, before checking access with
	// acp directly we need to make sure that the document is not public, as public documents will not
	// be registered with acp. We give unrestricted access to public documents, so it does not matter
	// whether the request has a signature identity or not at this stage of the check.
	isRegistered, err := documentACP.IsDocRegistered(
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

	identity := identityFunc()
	var identityValue string
	if !identity.HasValue() {
		// We can't assume that there is no-access just because there is no identity even if the document
		// is registered with acp, this is because it is possible that acp has a registered relation targeting
		// "*" (any) actor which would mean that even a request without an identity might be able to access
		// a document registered with acp. So we pass an empty `did` to accommodate that case.
		identityValue = ""
	} else {
		identityValue = identity.Value().DID
	}

	documentResourcePerm, ok := permission.(acpTypes.DocumentResourcePermission)
	if !ok {
		return false, ErrInvalidResourcePermissionType
	}

	// Now actually check using the signature if this identity has access or not.
	hasAccess, err := documentACP.CheckDocAccess(
		ctx,
		documentResourcePerm,
		identityValue,
		policyID,
		resourceName,
		docID,
	)

	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
