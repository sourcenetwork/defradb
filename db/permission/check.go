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

// CheckDocPermissionedAccessOnCollection handles the check, which tells us if access to the target
// document is valid, with respect to the permission type, and the specified collection.
//
// According to our access logic we have these components to worry about:
// (1) the request is permissioned (has an identity signature),
// (2) the collection is permissioned (has a policy),
// (3) acp module exists (acp is enabled).
//
// Unrestricted Access (Read + Write) to target document if:
// - Either of (2) or (3) is false.
// - Document is public (unregistered), whether signatured request or not, doesn't matter.
//
// Otherwise, check with acp module to verify signature has the appropriate access.
func CheckDocPermissionedAccessOnCollection(
	ctx context.Context,
	acpModuleOptional immutable.Option[acp.ACPModule],
	collection client.Collection,
	permission acp.DPIPermission,
	docID string,
) (bool, error) {
	// If no acp module, then we have unrestricted access.
	if !acpModuleOptional.HasValue() {
		return true, nil
	}

	// Even if acp module exists, but there is no policy on the collection (unpermissioned collection)
	// then we still have unrestricted access.
	policyID, resourceName, hasPolicy := IsPermissioned(collection)
	if !hasPolicy {
		return true, nil
	}

	acpModule := acpModuleOptional.Value()

	// Now that we know acp module exists and the collection is permissioned, before checking access with
	// acp module directly we need to make sure that the document is not public, as public documents will
	// not be regestered with acp. We give unrestricted access to public documents, so it does not matter
	// whether the request has a signature identity or not at this stage of the check.
	isNotPublic, err := acpModule.IsDocRegistered(
		ctx,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, err
	}

	if !isNotPublic {
		// Unrestricted access as it is a public document.
		return true, nil
	}

	// TODO-ACP: Implement signatures
	// hasSignature := false
	hasSignature := true

	// At this point if the request is not signatured, then it has no access, because:
	// the collection has a policy on it, the acp module is enabled/available,
	// and the document is not public (is regestered with the acp module).
	if !hasSignature {
		return false, nil
	}

	// Now actually check using the signature if this identity has access or not.
	hasAccess, err := acpModule.CheckDocAccess(
		ctx,
		permission,
		"cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969", // TODO-ACP: Replace with signature identity
		policyID,
		resourceName,
		docID,
	)

	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
