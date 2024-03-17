// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// PolicyDescription describes a policy which is made up of a valid policyID that is
// registered with acp module and has a valid DPI compliant resource name that also
// exists on that policy, the description is already validated.
type PolicyDescription struct {
	// ID is the local policyID when using local acp, and global policyID when
	// using remote acp with sourcehub. This identifier is externally managed
	// by the acp module.
	ID string

	// ResourceName is the name of the corresponding resource within the policy.
	ResourceName string
}
