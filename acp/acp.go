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
	"context"

	protoTypes "github.com/cosmos/gogoproto/types"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

// ACPSystemClient is an abstraction to allow multiple types of ACP systems to share DefraDB specific logic.
type ACPSystemClient interface {
	// Init initializes the acp, with an absolute path. The provided path indicates where the
	// persistent data will be stored for that ACP system.
	//
	// If the path is empty then acp will run in memory.
	Init(ctx context.Context, path string)

	// Start starts the acp, using the initialized path. Will recover acp state
	// from a previous run if under the same path.
	//
	// If the path is empty then acp will run in memory.
	Start(ctx context.Context) error

	// Close closes any resources in use by acp.
	Close() error

	// ResetState purges the entire ACP state.
	ResetState(context.Context) error

	// AddPolicy attempts to add the given policy. Upon success a policyID is returned,
	// otherwise returns error.
	AddPolicy(
		ctx context.Context,
		creator identity.Identity,
		policy string,
		marshalType acpTypes.PolicyMarshalType,
		creationTime *protoTypes.Timestamp,
	) (string, error)

	// Policy returns a policy of the given policyID if one is found.
	Policy(
		ctx context.Context,
		policyID string,
	) (immutable.Option[acpTypes.Policy], error)

	// RegisterObject registers the object to have access control.
	// No error is returned upon successful registering of an object.
	RegisterObject(
		ctx context.Context,
		identity identity.Identity,
		policyID string,
		resourceName string,
		objectID string,
		creationTime *protoTypes.Timestamp,
	) error

	// ObjectOwner returns the owner of the object of the given objectID.
	ObjectOwner(
		ctx context.Context,
		policyID string,
		resourceName string,
		objectID string,
	) (immutable.Option[string], error)

	// VerifyAccessRequest returns true if the check was successfull and the request has access to the object. If
	// the check was successful but the request does not have access to the object, then returns false.
	// Otherwise if check failed then an error is returned (and the boolean result should not be used).
	VerifyAccessRequest(
		ctx context.Context,
		permission acpTypes.ResourceInterfacePermission,
		actorID string,
		policyID string,
		resourceName string,
		objectID string,
	) (bool, error)

	// AddActorRelationship creates a relationship within a policy which ties the target actor
	// with the specified object, which means that the set of high level rules defined in the
	// policy will now apply to target actor as well.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship with actor already existed (no-op), and false if a new
	// relationship was made.
	//
	// Note: The requester identity must either be the owner of the object (being shared) or
	//       the manager (i.e. the relation has `manages` defined in the policy).
	AddActorRelationship(
		ctx context.Context,
		policyID string,
		resourceName string,
		objectID string,
		relation string,
		requester identity.Identity,
		targetActor string,
		creationTime *protoTypes.Timestamp,
	) (bool, error)

	// DeleteActorRelationship deletes a relationship within a policy which ties the target actor
	// with the specified object, which means that the set of high level rules defined in the
	// policy for that relation no-longer will apply to target actor anymore.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship record was found and deleted. Upon success the boolean value
	// will be false if the relationship record was not found (no-op).
	//
	// Note: The requester identity must either be the owner of the object (being shared) or
	//       the manager (i.e. the relation has `manages` defined in the policy).
	DeleteActorRelationship(
		ctx context.Context,
		policyID string,
		resourceName string,
		objectID string,
		relation string,
		requester identity.Identity,
		targetActor string,
		creationTime *protoTypes.Timestamp,
	) (bool, error)
}
