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

package acp

import "strings"

type DPIPermission int

// Valid DefraDB Policy Interface Permission Type.
const (
	ReadPermission DPIPermission = iota
	WritePermission
)

// List of all valid DPI permissions, the order of permissions in this list must match
// the above defined ordering such that iota matches the index position within the list.
var dpiRequiredPermissions = []string{
	"read",
	"write",
}

func (dpiPermission DPIPermission) String() string {
	return dpiRequiredPermissions[dpiPermission]
}

const requiredRegistererRelationName string = "owner"

// validateDPIExpressionOfRequiredPermission validates that the expression under the
// permission is valid. Moreover, DPI requires that for all required permissions, the
// expression start with "owner" then a space or symbol, and then follow-up expression.
// This is important because even if the policy has the required permissions under the
// resource, it's still possible that those permissions are not granted to the "owner"
// relation, this validation will help users not shoot themseleves in the foot.
//
// Eventhough the following expressions are valid policy expressions, they are illegal
// DPI expressions for required permissions.
// Some Invalid expression examples (assuming expr is under a required permission):
// - expr: ownerMalicious + owner
// - expr: ownerMalicious
// - expr: owner_new
// - expr: reader+owner
// - expr: reader-owner
// - expr: reader - owner
//
// Some Valid expression examples (assuming expr is under a required permission):
// - expr: owner
// - expr: owner + reader
// - expr: owner +reader
// - expr: owner+reader
// - expr: owner-reader
// - expr: owner&reader
func validateDPIExpressionOfRequiredPermission(expression string, requiredPermission string) error {
	const requiredRelationLength = len(requiredRegistererRelationName)
	exprNoSpace := strings.ReplaceAll(strings.TrimSpace(expression), " ", "")

	if len(exprNoSpace) < requiredRelationLength {
		return newErrExprOfRequiredPermissionIsMissingRelation(
			requiredPermission,
			requiredRegistererRelationName,
		)
	}

	exprStart := exprNoSpace[0:requiredRelationLength]
	exprRest := exprNoSpace[requiredRelationLength:]

	if exprStart != requiredRegistererRelationName {
		return newErrExprOfRequiredPermissionMustStartWithRelation(
			requiredPermission,
			requiredRegistererRelationName,
		)
	}

	if len(exprRest) != 0 {
		c := exprRest[0]
		// First non-space character after the required relation name MUST be one of `+`, `-`, `&`.
		if c != '+' && c != '-' && c != '&' {
			return newErrExprOfRequiredPermissionHasInvalidChar(
				requiredPermission,
				requiredRegistererRelationName,
			)
		}
	}

	return nil
}
