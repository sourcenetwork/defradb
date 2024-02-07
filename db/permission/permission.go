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
	"github.com/sourcenetwork/defradb/client"
)

// IsPermissioned returns true if the collection has a policy, otherwise returns false.
//
// This tells us if access control is enabled for this collection or not.
func IsPermissioned(collection client.Collection) (string, string, bool) {
	policy := collection.Definition().Description.Policy
	if policy.HasValue() &&
		policy.Value().ID != "" &&
		policy.Value().ResourceName != "" {
		return policy.Value().ID, policy.Value().ResourceName, true
	}

	return "", "", false
}
