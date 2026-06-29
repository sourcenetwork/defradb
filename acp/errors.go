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
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errInvalidACPSystem                          = "invalid acp system"
	errInitializationOfACPFailed                 = "initialization of acp failed"
	errStartingACPInEmptyPath                    = "starting acp in an empty path"
	errFailedToAddPolicy                         = "failed to add policy with acp"
	errFailedToRegisterDoc                       = "failed to register document with acp"
	errFailedToCheckIfDocIsRegistered            = "failed to check if doc is registered with acp"
	errFailedToVerifyDocAccess                   = "failed to verify doc access with acp"
	errFailedToVerifyAdminAccess                 = "failed to verify admin access with acp"
	errFailedToAddDocActorRelationship           = "failed to add document actor relationship with acp"
	errFailedToDeleteDocActorRelationship        = "failed to delete document actor relationship with acp"
	errMissingReqArgToAddDocActorRelationship    = "missing a required argument needed to add doc actor relationship"
	errMissingReqArgToDeleteDocActorRelationship = "missing a required argument needed to delete doc actor relationship"

	errNoPolicyArgs = "missing policy arguments, must have both id and resource"

	errPolicyIDMustNotBeEmpty = "policyID must not be empty"
	errPolicyDoesNotExist     = "policyID specified does not exist with acp"
	errPolicyValidationFailed = "policyID validation through acp failed"

	errResourceNameMustNotBeEmpty          = "resource name must not be empty"
	errResourceDoesNotExistOnTargetPolicy  = "resource does not exist on the specified policy"
	errResourceIsMissingRequiredPermission = "resource is missing required permission on policy"

	errExprOfRequiredPermMustStartWithRelation = "expr of required permission must start with required relation"
	errExprOfRequiredPermHasInvalidChar        = "expr of required permission has invalid character after relation"

	errInvalidActorID = "invalid actor ID"
)

var (
	ErrInvalidACPSystem                   = errors.New(errInvalidACPSystem)
	ErrInitializationOfACPFailed          = errors.New(errInitializationOfACPFailed)
	ErrFailedToAddPolicy                  = errors.New(errFailedToAddPolicy)
	ErrFailedToRegisterDoc                = errors.New(errFailedToRegisterDoc)
	ErrFailedToCheckIfDocIsRegistered     = errors.New(errFailedToCheckIfDocIsRegistered)
	ErrFailedToVerifyDocAccess            = errors.New(errFailedToVerifyDocAccess)
	ErrFailedToAddDocActorRelationship    = errors.New(errFailedToAddDocActorRelationship)
	ErrFailedToDeleteDocActorRelationship = errors.New(errFailedToDeleteDocActorRelationship)
	ErrPolicyDoesNotExist                 = errors.New(errPolicyDoesNotExist)

	ErrResourceDoesNotExistOnTargetPolicy = errors.New(errResourceDoesNotExistOnTargetPolicy)

	ErrPolicyDataMustNotBeEmpty                    = errors.New("policy data can not be empty")
	ErrPolicyCreatorMustNotBeEmpty                 = errors.New("policy creator can not be empty")
	ErrACPResetState                               = errors.New("acp could not be reset")
	ErrLocalACPStoreNameEmptyWithPersistentStorage = errors.New("acp store name is empty with persistent storage")
	ErrLocalACPAlreadyStarted                      = errors.New("trying to start local acp, when it is already started")
	ErrNoPolicyArgs                                = errors.New(errNoPolicyArgs)
	ErrPolicyIDMustNotBeEmpty                      = errors.New(errPolicyIDMustNotBeEmpty)
	ErrResourceNameMustNotBeEmpty                  = errors.New(errResourceNameMustNotBeEmpty)
	ErrInvalidActorID                              = errors.New(errInvalidActorID)
	ErrResourceIsMissingRequiredPermission         = errors.New(errResourceIsMissingRequiredPermission)
)

func NewErrInitializationOfACPFailed(
	inner error,
	Type string,
	path string,
) error {
	return errors.Wrap(
		errInitializationOfACPFailed,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("Path", path),
	)
}

func NewErrFailedToAddPolicy(
	inner error,
	Type string,
	creatorID string,
) error {
	return errors.Wrap(
		errFailedToAddPolicy,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("CreatorID", creatorID),
	)
}

func NewErrFailedToRegisterDoc(
	inner error,
	Type string,
	policyID string,
	creatorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToRegisterDoc,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("CreatorID", creatorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToCheckIfDocIsRegistered(
	inner error,
	Type string,
	policyID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToCheckIfDocIsRegistered,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToVerifyNodeAccess(
	inner error,
	permission string,
	policyID string,
	actorID string,
	resourceName string,
	operation string,
) error {
	return errors.Wrap(
		errFailedToVerifyAdminAccess,
		inner,
		errors.NewKV("Permission", permission),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ActorID", actorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("Operation", operation),
	)
}

func NewErrFailedToVerifyDocAccess(
	inner error,
	Type string,
	permission string,
	policyID string,
	actorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToVerifyDocAccess,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("Permission", permission),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ActorID", actorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToAddDocActorRelationship(
	inner error,
	Type string,
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor string,
	targetActor string,
) error {
	return errors.Wrap(
		errFailedToAddDocActorRelationship,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
		errors.NewKV("Relation", relation),
		errors.NewKV("RequestActor", requestActor),
		errors.NewKV("TargetActor", targetActor),
	)
}

func NewErrFailedToDeleteDocActorRelationship(
	inner error,
	Type string,
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor string,
	targetActor string,
) error {
	return errors.Wrap(
		errFailedToDeleteDocActorRelationship,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
		errors.NewKV("Relation", relation),
		errors.NewKV("RequestActor", requestActor),
		errors.NewKV("TargetActor", targetActor),
	)
}

func NewErrPolicyDoesNotExist(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyDoesNotExist,
		inner,
		errors.NewKV("PolicyID", policyID),
	)
}

func NewErrPolicyValidationFailed(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyValidationFailed,
		inner,
		errors.NewKV("PolicyID", policyID),
	)
}

func NewErrResourceDoesNotExistOnTargetPolicy(
	resourceName string,
	policyID string,
) error {
	return errors.New(
		errResourceDoesNotExistOnTargetPolicy,
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
	)
}
func NewErrInvalidACPSystem(
	unknownACP string,
) error {
	return errors.New(
		errInvalidACPSystem,
		errors.NewKV("UnknownACP", unknownACP),
	)
}

func NewErrResourceIsMissingRequiredPermission(
	resourceName string,
	permission string,
	policyID string,
) error {
	return errors.New(
		errResourceIsMissingRequiredPermission,
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("Permission", permission),
	)
}

func NewErrExprOfRequiredPermissionMustStartWithRelation(
	permission string,
	relation string,
) error {
	return errors.New(
		errExprOfRequiredPermMustStartWithRelation,
		errors.NewKV("Permission", permission),
		errors.NewKV("Relation", relation),
	)
}

func NewErrExprOfRequiredPermissionHasInvalidChar(
	permission string,
	relation string,
	char byte,
) error {
	return errors.New(
		errExprOfRequiredPermHasInvalidChar,
		errors.NewKV("Permission", permission),
		errors.NewKV("Relation", relation),
		errors.NewKV("Character", string(char)),
	)
}

func NewErrMissingRequiredArgToAddDocActorRelationship(
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor string,
	targetActor string,
) error {
	return errors.New(
		errMissingReqArgToAddDocActorRelationship,
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
		errors.NewKV("Relation", relation),
		errors.NewKV("RequestActor", requestActor),
		errors.NewKV("TargetActor", targetActor),
	)
}

func NewErrMissingRequiredArgToDeleteDocActorRelationship(
	policyID string,
	resourceName string,
	docID string,
	relation string,
	requestActor string,
	targetActor string,
) error {
	return errors.New(
		errMissingReqArgToDeleteDocActorRelationship,
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
		errors.NewKV("Relation", relation),
		errors.NewKV("RequestActor", requestActor),
		errors.NewKV("TargetActor", targetActor),
	)
}

func NewErrInvalidActorID(
	inner error,
	id string,
) error {
	return errors.Wrap(
		errInvalidActorID,
		inner,
		errors.NewKV("ActorID", id),
	)
}
