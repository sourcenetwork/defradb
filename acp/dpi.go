// Copyright 2024 Democratized Data Foundation
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
	"strings"
)

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
// relation. This validation will help users not shoot themseleves in the foot.
//
// Learn more about the DefraDB Policy Interface [ACP](/acp/README.md), can find more
// detailed valid and invalid `expr` (expression) examples there.
func validateDPIExpressionOfRequiredPermission(expression string, requiredPermission string) error {
	exprNoSpace := strings.ReplaceAll(expression, " ", "")

	if !strings.HasPrefix(exprNoSpace, requiredRegistererRelationName) {
		return newErrExprOfRequiredPermissionMustStartWithRelation(
			requiredPermission,
			requiredRegistererRelationName,
		)
	}

	restOfTheExpr := exprNoSpace[len(requiredRegistererRelationName):]
	if len(restOfTheExpr) != 0 {
		c := restOfTheExpr[0]
		// First non-space character after the required relation name MUST be a `+`.
		// The reason we are enforcing this here is because other set operations are
		// not applied to the registerer relation anyways.
		if c != '+' {
			return newErrExprOfRequiredPermissionHasInvalidChar(
				requiredPermission,
				requiredRegistererRelationName,
				c,
			)
		}
	}

	return nil
}
