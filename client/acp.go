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

// PolicyDescription describes a policy using it's ID and it's resource name, where:
// 1) the ID is the policyID of the registered policy on the document acp system and
// 2) the resource name is of a valid resource that adheres to the corresponding
// resource interface (RI) rules.
type PolicyDescription struct {
	// ID is the local policyID when using local acp, and global policyID when
	// using remote acp with sourcehub. This identifier is externally managed
	// by the acp system.
	ID string

	// ResourceName is the name of the corresponding resource within the policy.
	ResourceName string
}

// AddPolicyResult wraps the result of successfully adding/registering a Policy.
type AddPolicyResult struct {
	// PolicyID is the unique identifier returned by the acp system,
	// upon successful creation of a policy.
	PolicyID string
}

// AddActorRelationshipResult wraps the result of making an object-actor relationship.
type AddActorRelationshipResult struct {
	// ExistedAlready is true if the relationship existed already (no-op), and
	// it is false if a new relationship was created.
	ExistedAlready bool
}

// DeleteActorRelationshipResult wraps the result of making an object-actor relationship.
type DeleteActorRelationshipResult struct {
	// RecordFound is true if the relationship record was found, and
	// is false if the relationship record was not found (no-op).
	RecordFound bool
}

// NACStatus represents the current state/status of the Node ACP system.
type NACStatus int

const (
	// NACNotConfigured indicates that NAC system is in a clean state, meaning it has not been started/configured yet.
	//
	// Note: Upon purge or reset, NAC will be set back to this state as well.
	NACNotConfigured NACStatus = iota

	// NACEnabled indicates that NAC system was started and is currently enabled.
	NACEnabled

	// NACDisabledTemporarily indicates that NAC system was started but is temporarily disabled.
	NACDisabledTemporarily
)

func (status NACStatus) String() string {
	switch status {
	case NACNotConfigured:
		return "not configured"
	case NACEnabled:
		return "enabled"
	case NACDisabledTemporarily:
		return "disabled temporarily"
	default:
		return "invalid state"
	}
}

// NACStatusResult wraps the result of current node acp status.
type NACStatusResult struct {
	Status string
}
