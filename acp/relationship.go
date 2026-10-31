// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import coreTypes "github.com/sourcenetwork/acp_core/pkg/types"

// ValidateActorRelationshipTarget checks that the target actor is not empty.
// It is also used by transports that send relationship arguments instead of ACP-core types.
func ValidateActorRelationshipTarget(targetActor string) error {
	if targetActor == "" {
		return ErrRelationshipTargetActorEmpty
	}
	return nil
}

// NewActorRelationship constructs a relationship with a non-empty target actor.
// A target of "*" creates a relationship with all actors instead of a single actor.
func NewActorRelationship(resourceName, objectID, relation, targetActor string) (*coreTypes.Relationship, error) {
	if err := ValidateActorRelationshipTarget(targetActor); err != nil {
		return nil, err
	}
	if targetActor == "*" {
		return coreTypes.NewAllActorsRelationship(resourceName, objectID, relation), nil
	}
	return coreTypes.NewActorRelationship(resourceName, objectID, relation, targetActor), nil
}
