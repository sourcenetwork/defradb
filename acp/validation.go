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
	"strings"

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
		permissionResponse, ok := resourceResponse.Permissions[requiredPermission]
		if !ok {
			return NewErrResourceIsMissingRequiredPermission(
				resourceName,
				requiredPermission,
				policyID,
			)
		}

		// Now we need to ensure that the "owner" relation has access to all the required resource
		// interface permissions. This is important because even if the policy has the required
		// permissions under the resource, it's possible that those permissions are not granted
		// to the "owner" relation, this will help users not shoot themseleves in the foot.
		// TODO-ACP: Better validation, once sourcehub implements meta-policies.
		// Issue: https://github.com/sourcenetwork/defradb/issues/2359
		if err := validateExpressionOfRequiredPermission(
			permissionResponse.Expression,
			requiredPermission,
		); err != nil {
			return err
		}
	}

	return nil
}

// validateExpressionOfRequiredPermission validates that the expression under the
// permission is valid. Moreover, resource interface requires that for all required
// permissions, the expression start with "owner" then a space or symbol, and then follow-up expression.
// This is important because even if the policy has the required permissions under the
// resource, it's still possible that those permissions are not granted to the "owner"
// relation. This validation will help users not shoot themseleves in the foot.
//
// Learn more about the DefraDB [ACP System](/acp/README.md)
func validateExpressionOfRequiredPermission(expression string, requiredPermission string) error {
	exprNoSpace := strings.ReplaceAll(expression, " ", "")

	if !strings.HasPrefix(exprNoSpace, acpTypes.RequiredRegistererRelationName) {
		return NewErrExprOfRequiredPermissionMustStartWithRelation(
			requiredPermission,
			acpTypes.RequiredRegistererRelationName,
		)
	}

	restOfTheExpr := exprNoSpace[len(acpTypes.RequiredRegistererRelationName):]
	if len(restOfTheExpr) != 0 {
		c := restOfTheExpr[0]
		// First non-space character after the required relation name MUST be a `+`.
		// The reason we are enforcing this here is because other set operations are
		// not applied to the registerer relation anyways.
		if c != '+' {
			return NewErrExprOfRequiredPermissionHasInvalidChar(
				requiredPermission,
				acpTypes.RequiredRegistererRelationName,
				c,
			)
		}
	}

	return nil
}
