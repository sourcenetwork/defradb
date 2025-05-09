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
	errFailedToAddPolicyWithACP                  = "failed to add policy with acp"
	errFailedToRegisterDocWithACP                = "failed to register document with acp"
	errFailedToCheckIfDocIsRegisteredWithACP     = "failed to check if doc is registered with acp"
	errFailedToVerifyDocAccessWithACP            = "failed to verify doc access with acp"
	errFailedToAddDocActorRelationshipWithACP    = "failed to add document actor relationship with acp"
	errFailedToDeleteDocActorRelationshipWithACP = "failed to delete document actor relationship with acp"
	errMissingReqArgToAddDocActorRelationship    = "missing a required argument needed to add doc actor relationship"
	errMissingReqArgToDeleteDocActorRelationship = "missing a required argument needed to delete doc actor relationship"

	errNoPolicyArgs = "missing policy arguments, must have both id and resource"

	errPolicyIDMustNotBeEmpty        = "policyID must not be empty"
	errPolicyDoesNotExistWithACP     = "policyID specified does not exist with acp"
	errPolicyValidationFailedWithACP = "policyID validation through acp failed"

	errResourceNameMustNotBeEmpty          = "resource name must not be empty"
	errResourceDoesNotExistOnTargetPolicy  = "resource does not exist on the specified policy"
	errResourceIsMissingRequiredPermission = "resource is missing required permission on policy"

	errExprOfRequiredPermMustStartWithRelation = "expr of required permission must start with required relation"
	errExprOfRequiredPermHasInvalidChar        = "expr of required permission has invalid character after relation"

	errInvalidActorID = "invalid actor ID"
)

var (
	ErrInvalidACPSystem                          = errors.New(errInvalidACPSystem)
	ErrInitializationOfACPFailed                 = errors.New(errInitializationOfACPFailed)
	ErrFailedToAddPolicyWithACP                  = errors.New(errFailedToAddPolicyWithACP)
	ErrFailedToRegisterDocWithACP                = errors.New(errFailedToRegisterDocWithACP)
	ErrFailedToCheckIfDocIsRegisteredWithACP     = errors.New(errFailedToCheckIfDocIsRegisteredWithACP)
	ErrFailedToVerifyDocAccessWithACP            = errors.New(errFailedToVerifyDocAccessWithACP)
	ErrFailedToAddDocActorRelationshipWithACP    = errors.New(errFailedToAddDocActorRelationshipWithACP)
	ErrFailedToDeleteDocActorRelationshipWithACP = errors.New(errFailedToDeleteDocActorRelationshipWithACP)
	ErrPolicyDoesNotExistWithACP                 = errors.New(errPolicyDoesNotExistWithACP)

	ErrResourceDoesNotExistOnTargetPolicy = errors.New(errResourceDoesNotExistOnTargetPolicy)

	ErrPolicyDataMustNotBeEmpty    = errors.New("policy data can not be empty")
	ErrPolicyCreatorMustNotBeEmpty = errors.New("policy creator can not be empty")
	ErrACPResetState               = errors.New("acp could not be reset")
	ErrNoPolicyArgs                = errors.New(errNoPolicyArgs)
	ErrPolicyIDMustNotBeEmpty      = errors.New(errPolicyIDMustNotBeEmpty)
	ErrResourceNameMustNotBeEmpty  = errors.New(errResourceNameMustNotBeEmpty)
	ErrInvalidActorID              = errors.New(errInvalidActorID)
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

func NewErrFailedToAddPolicyWithACP(
	inner error,
	Type string,
	creatorID string,
) error {
	return errors.Wrap(
		errFailedToAddPolicyWithACP,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("CreatorID", creatorID),
	)
}

func NewErrFailedToRegisterDocWithACP(
	inner error,
	Type string,
	policyID string,
	creatorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToRegisterDocWithACP,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("CreatorID", creatorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToCheckIfDocIsRegisteredWithACP(
	inner error,
	Type string,
	policyID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToCheckIfDocIsRegisteredWithACP,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToVerifyDocAccessWithACP(
	inner error,
	Type string,
	permission string,
	policyID string,
	actorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToVerifyDocAccessWithACP,
		inner,
		errors.NewKV("Type", Type),
		errors.NewKV("Permission", permission),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ActorID", actorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToAddDocActorRelationshipWithACP(
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
		errFailedToAddDocActorRelationshipWithACP,
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

func NewErrFailedToDeleteDocActorRelationshipWithACP(
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
		errFailedToDeleteDocActorRelationshipWithACP,
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

func NewErrPolicyDoesNotExistWithACP(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyDoesNotExistWithACP,
		inner,
		errors.NewKV("PolicyID", policyID),
	)
}

func NewErrPolicyValidationFailedWithACP(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyValidationFailedWithACP,
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
