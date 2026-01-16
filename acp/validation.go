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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

func ValidateResourceInterface(
	ctx context.Context,
	policyID string,
	resourceName string,
	acpType acpTypes.ACPSystemType,
	acpClient ACPSystemClient,
) error {
	if policyID == "" && resourceName == "" {
		return ErrNoPolicyArgs
	}

	if policyID == "" {
		return ErrPolicyIDMustNotBeEmpty
	}

	if resourceName == "" {
		return ErrResourceNameMustNotBeEmpty
	}

	maybePolicy, err := acpClient.Policy(ctx, policyID)

	if err != nil {
		return NewErrPolicyValidationFailedWithACP(err, policyID)
	}
	if !maybePolicy.HasValue() {
		return NewErrPolicyDoesNotExistWithACP(err, policyID)
	}

	policy := maybePolicy.Value()

	// So far we validated that the policy exists, now lets validate that resource exists.
	resourceResponse, ok := policy.Resources[resourceName]
	if !ok {
		return NewErrResourceDoesNotExistOnTargetPolicy(resourceName, policyID)
	}

	var requiredResourcePermissions []string
	switch acpType {
	case acpTypes.LocalDocumentACP, acpTypes.SourceHubDocumentACP:
		requiredResourcePermissions = acpTypes.RequiredResourcePermissionsForDocument
	case acpTypes.NodeACP:
		requiredResourcePermissions = acpTypes.RequiredResourcePermissionsForNode
	default:
		return NewErrInvalidACPSystem(resourceName)
	}

	// Now that we have validated that policyID exists and it contains a corresponding
	// resource with the matching name, validate that all required resource interface
	// permissions actually exist on the target resource.
	for _, requiredPermission := range requiredResourcePermissions {
		_, ok := resourceResponse.Permissions[requiredPermission]
		if !ok {
			return NewErrResourceIsMissingRequiredPermission(
				resourceName,
				requiredPermission,
				policyID,
			)
		}
	}

	return nil
}
