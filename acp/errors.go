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
	errInitializationOfACPFailed             = "initialization of acp module failed"
	errStartingACPInEmptyPath                = "starting acp module in an empty path"
	errFailedToAddPolicyWithACP              = "failed to add policy with acp module"
	errFailedToRegisterDocWithACP            = "failed to register document with acp module"
	errFailedToCheckIfDocIsRegisteredWithACP = "failed to check if doc is registered with acp module"
	errFailedToVerifyDocAccessWithACP        = "failed to verify doc access with acp module"

	errObjectDidNotRegister = "no-op while registering object (already exists or error) with acp module"
	errNoPolicyArgs         = "missing policy arguments, must have both id and resource"

	errPolicyIDMustNotBeEmpty            = "policyID must not be empty"
	errPolicyDoesNotExistOnACPModule     = "policyID specified does not exist on acp module"
	errPolicyValidationFailedOnACPModule = "policyID validation through acp module failed"

	errResourceNameMustNotBeEmpty          = "resource name must not be empty"
	errResourceDoesNotExistOnTargetPolicy  = "resource does not exist on the specified policy"
	errResourceIsMissingRequiredPermission = "resource is missing required permission on policy"

	errExprOfRequiredPermMustStartWithRelation = "expr of required permission must start with required relation"
	errExprOfRequiredPermHasInvalidChar        = "expr of required permission has invalid character after relation"
)

var (
	ErrInitializationOfACPFailed             = errors.New(errInitializationOfACPFailed)
	ErrFailedToAddPolicyWithACP              = errors.New(errFailedToAddPolicyWithACP)
	ErrFailedToRegisterDocWithACP            = errors.New(errFailedToRegisterDocWithACP)
	ErrFailedToCheckIfDocIsRegisteredWithACP = errors.New(errFailedToCheckIfDocIsRegisteredWithACP)
	ErrFailedToVerifyDocAccessWithACP        = errors.New(errFailedToVerifyDocAccessWithACP)
	ErrPolicyDoesNotExistOnACPModule         = errors.New(errPolicyDoesNotExistOnACPModule)

	ErrResourceDoesNotExistOnTargetPolicy = errors.New(errResourceDoesNotExistOnTargetPolicy)

	ErrPolicyDataMustNotBeEmpty    = errors.New("policy data can not be empty")
	ErrPolicyCreatorMustNotBeEmpty = errors.New("policy creator can not be empty")
	ErrObjectDidNotRegister        = errors.New(errObjectDidNotRegister)
	ErrNoPolicyArgs                = errors.New(errNoPolicyArgs)
	ErrPolicyIDMustNotBeEmpty      = errors.New(errPolicyIDMustNotBeEmpty)
	ErrResourceNameMustNotBeEmpty  = errors.New(errResourceNameMustNotBeEmpty)
)

func NewErrInitializationOfACPFailed(
	inner error,
	moduleType string,
	path string,
) error {
	return errors.Wrap(
		errInitializationOfACPFailed,
		inner,
		errors.NewKV("ModuleType", moduleType),
		errors.NewKV("Path", path),
	)
}

func NewErrFailedToAddPolicyWithACP(
	inner error,
	moduleType string,
	creatorID string,
) error {
	return errors.Wrap(
		errFailedToAddPolicyWithACP,
		inner,
		errors.NewKV("ModuleType", moduleType),
		errors.NewKV("CreatorID", creatorID),
	)
}

func NewErrFailedToRegisterDocWithACP(
	inner error,
	moduleType string,
	policyID string,
	creatorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToRegisterDocWithACP,
		inner,
		errors.NewKV("ModuleType", moduleType),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("CreatorID", creatorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToCheckIfDocIsRegisteredWithACP(
	inner error,
	moduleType string,
	policyID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToCheckIfDocIsRegisteredWithACP,
		inner,
		errors.NewKV("ModuleType", moduleType),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func NewErrFailedToVerifyDocAccessWithACP(
	inner error,
	moduleType string,
	policyID string,
	actorID string,
	resourceName string,
	docID string,
) error {
	return errors.Wrap(
		errFailedToVerifyDocAccessWithACP,
		inner,
		errors.NewKV("ModuleType", moduleType),
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ActorID", actorID),
		errors.NewKV("ResourceName", resourceName),
		errors.NewKV("DocID", docID),
	)
}

func newErrPolicyDoesNotExistOnACPModule(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyDoesNotExistOnACPModule,
		inner,
		errors.NewKV("PolicyID", policyID),
	)
}

func newErrPolicyValidationFailedOnACPModule(
	inner error,
	policyID string,
) error {
	return errors.Wrap(
		errPolicyValidationFailedOnACPModule,
		inner,
		errors.NewKV("PolicyID", policyID),
	)
}

func newErrResourceDoesNotExistOnTargetPolicy(
	resourceName string,
	policyID string,
) error {
	return errors.New(
		errResourceDoesNotExistOnTargetPolicy,
		errors.NewKV("PolicyID", policyID),
		errors.NewKV("ResourceName", resourceName),
	)
}

func newErrResourceIsMissingRequiredPermission(
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

func newErrExprOfRequiredPermissionMustStartWithRelation(
	permission string,
	relation string,
) error {
	return errors.New(
		errExprOfRequiredPermMustStartWithRelation,
		errors.NewKV("Permission", permission),
		errors.NewKV("Relation", relation),
	)
}

func newErrExprOfRequiredPermissionHasInvalidChar(
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