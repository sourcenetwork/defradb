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

// PolicyDescription describes a policy by its ID and resource name, where:
// 1) the ID identifies the policy registered with the ACP system and
// 2) the resource name identifies a valid resource that adheres to the corresponding
// resource interface (RI) rules.
type PolicyDescription struct {
	// ID is the policy identifier managed by the configured ACP system.
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
	// it is false if a new relationship was added.
	ExistedAlready bool
}

// DeleteActorRelationshipResult wraps the result of making an object-actor relationship.
type DeleteActorRelationshipResult struct {
	// RecordFound is true if the relationship record was found, and
	// is false if the relationship record was not found (no-op).
	RecordFound bool
}

// NACStatus represents the current state of the Local NAC system.
type NACStatus int

const (
	// NACNotConfigured indicates that Local NAC has not been configured yet.
	//
	// Note: Upon purge or reset, NAC will be set back to this state as well.
	NACNotConfigured NACStatus = iota

	// NACEnabled indicates that Local NAC was started and is currently enabled.
	NACEnabled

	// NACDisabledTemporarily indicates that Local NAC was started but is temporarily disabled.
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

// NACStatusResult wraps the current Local NAC status.
type NACStatusResult struct {
	Status string
}
