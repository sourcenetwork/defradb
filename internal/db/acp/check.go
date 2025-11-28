// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
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
// (1) the request is permissioned (has an identity).
// (2) the collection is permissioned (has a policy).
// (3) the identity has "dac-bypass" Node ACP permission.
//
// Unrestricted Access to document if:
// - (2) is false.
// - Document is public (unregistered), whether signatured request or not doesn't matter.
// - (3) is true.
func CheckAccessOfDocOnCollectionWithACP(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	nodeACP NACInfo,
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
		nodeACP,
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
// (1) the request is permissioned (has an identity).
// (2) the collection is permissioned (has a policy).
// (3) the identity has "dac-bypass" Node ACP permission.
//
// Unrestricted Access to document if:
// - (2) is false.
// - Document is public (unregistered), whether signatured request or not doesn't matter.
// - (3) is true.
func CheckDocAccessWithIdentityFunc(
	ctx context.Context,
	identityFunc func() immutable.Option[acpIdentity.Identity],
	nodeACP NACInfo,
	documentACP dac.DocumentACP,
	collection client.Collection,
	permission acpTypes.ResourceInterfacePermission,
	docID string,
) (bool, error) {
	identity := identityFunc()
	var identityValue string
	// Note: The following must be done to handle the "*" edge case before:
	// - calling [hasDACBypass] which calls [CheckNodeOperationAccess].
	// - and before calling [documentACP.CheckDocAccess]
	if !identity.HasValue() {
		// We can't assume that there is no-access just because there is no identity even if the
		// resource is registered with the acp system, this is because it is possible that acp has
		// a registered relation targeting "*" (any) actor which would mean that even a request
		// without an identity might be able to access a document registered with acp, or might
		// have dac-bypass nac privilage. So we pass an empty `did` to accommodate that case.
		identityValue = ""
	} else {
		identityValue = identity.Value().DID()
	}

	// Check if can bypass DAC, if not then continue DAC.
	if canDACBypass(ctx, nodeACP, identityValue) {
		return true, nil
	}

	// Even if document acp exists, but there is no policy on the collection (unpermissioned collection)
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

	documentResourcePerm, ok := permission.(acpTypes.DocumentResourcePermission)
	if !ok {
		return false, client.ErrInvalidResourcePermissionType
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

// CheckNodeOperationAccess returns an [client.ErrNotAuthorizedToPerformOperation]
// error if the requesting user does not have the required permission to perform an operation.
// If something else goes wrong, it returns a different error, otherwise returns nil only if
// the check passes and the requesting user is authorized to perform the operation.
//
// Unrestricted access if:
// - node acp system is temporarily disabled (unless the operation is trying to turn on nac).
func CheckNodeOperationAccess(
	ctx context.Context,
	identity string,
	nacInfo NACInfo,
	permission acpTypes.ResourceInterfacePermission,
	objectID string,
) error {
	if nacInfo.NodeACPDesc.Status != client.NACEnabled &&
		permission != acpTypes.NodeNACReEnablePerm {
		// Unrestricted access if node acp is off, and not trying to turn it back on.
		return nil
	}

	// If node acp is enabled then it must have have node acp instance setup.
	if nacInfo.NodeACP == nil {
		return client.ErrNACIsEnabledButInstanceIsNotAvailable
	}

	// If node acp is enabled then it must have a valid policy information.
	if !nacInfo.NodeACPDesc.Policy.HasValue() {
		return client.ErrNACIsEnabledButIsMissingPolicyInfo
	}

	policyID := nacInfo.NodeACPDesc.Policy.Value().ID
	resourceName := nacInfo.NodeACPDesc.Policy.Value().ResourceName
	if policyID == "" || resourceName == "" {
		return client.ErrNACIsEnabledButIsMissingPolicyInfo
	}
	// Since public node will have unrestricted access, the object we are gating MUST be registered
	// if node access control is configured.
	isRegistered, err := nacInfo.NodeACP.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		objectID,
	)
	if err != nil {
		return err
	}

	if !isRegistered.HasValue() {
		return client.ErrNACNodeObjectToGateIsNotRegistered
	}

	nodeResourcePerm, ok := permission.(acpTypes.NodeResourcePermission)
	if !ok {
		return client.ErrInvalidResourcePermissionType
	}

	// Now actually check if this identity has access or not.
	hasAccess, err := nacInfo.NodeACP.VerifyAccessRequest(
		ctx,
		nodeResourcePerm,
		identity,
		policyID,
		resourceName,
		objectID,
	)

	if err != nil {
		return acp.NewErrFailedToVerifyNodeAccessWithACP(
			err,
			permission.String(),
			policyID,
			identity,
			resourceName,
			objectID,
		)
	}

	if hasAccess {
		return nil
	}

	return client.ErrNotAuthorizedToPerformOperation
}
