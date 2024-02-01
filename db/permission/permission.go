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
	"github.com/sourcenetwork/defradb/client"
)

// isPermissioned returns true if the collection has a policy, otherwise returns false.
//
// This tells us if access control is enabled for this collection or not.
//
// When there is a policy, in addition to returning true in the last return value, the
// first returned value is policyID, second is the resource name.
func isPermissioned(collection client.Collection) (string, string, bool) {
	policy := collection.Definition().Description.Policy
	if policy.HasValue() &&
		policy.Value().ID != "" &&
		policy.Value().ResourceName != "" {
		return policy.Value().ID, policy.Value().ResourceName, true
	}

	return "", "", false
}
