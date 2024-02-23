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
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errObjectDidNotRegister = "no-op while registering object (already exists or error) with acp module"
	errNoPolicyArgs         = "missing policy arguments, must have both id and resource"

	errPolicyIDMustNotBeEmpty            = "policyID must not be empty"
	errPolicyDoesNotExistOnACPModule     = "policyID=%s specified does not exist on acp module"
	errPolicyValidationFailedOnACPModule = "policyID=%s validation through acp module failed"

	errResourceNameMustNotBeEmpty          = "resource name must not be empty"
	errResourceDoesNotExistOnTargetPolicy  = "resource=%s does not exist on the specified policy=%s"
	errResourceIsMissingRequiredPermission = "resource=%s is missing required permission=%s on policy=%s"

	errExprOfRequiredPermIsMissingRelation     = "expr of required permission=%s is missing required relation=%s"
	errExprOfRequiredPermMustStartWithRelation = "expr of required permission=%s must start with required relation=%s"
	errExprOfRequiredPermHasInvalidChar        = "expr of required permission=%s has invalid char after relation=%s"
)

var (
	ErrObjectDidNotRegister       = errors.New(errObjectDidNotRegister)
	ErrNoPolicyArgs               = errors.New(errNoPolicyArgs)
	ErrPolicyIDMustNotBeEmpty     = errors.New(errPolicyIDMustNotBeEmpty)
	ErrResourceNameMustNotBeEmpty = errors.New(errResourceNameMustNotBeEmpty)
)

// Use with `errPolicyDoesNotExistOnACPModule` or `errPolicyValidationFailedOnACPModule`.
func newErrPolicyIDValidation(
	inner error,
	policyID string,
	message string,
	kv ...errors.KV,
) error {
	return errors.Wrap(
		fmt.Sprintf(
			message,
			policyID,
		),
		inner,
		kv...,
	)
}

func newErrResourceDoesNotExistOnTargetPolicy(
	inner error,
	resourceName string,
	policyID string,
	kv ...errors.KV,
) error {
	return errors.Wrap(
		fmt.Sprintf(
			errResourceDoesNotExistOnTargetPolicy,
			resourceName,
			policyID,
		),
		inner,
		kv...,
	)
}

func newErrResourceIsMissingRequiredPermission(
	inner error,
	resourceName string,
	permission string,
	policyID string,
	kv ...errors.KV,
) error {
	return errors.Wrap(
		fmt.Sprintf(
			errResourceIsMissingRequiredPermission,
			resourceName,
			permission,
			policyID,
		),
		inner,
		kv...,
	)
}

func newErrExprOfRequiredPermissionIsMissingRelation(
	permission string,
	relation string,
	kv ...errors.KV,
) error {
	return errors.New(
		fmt.Sprintf(
			errExprOfRequiredPermIsMissingRelation,
			permission,
			relation,
		),
		kv...,
	)
}

func newErrExprOfRequiredPermissionMustStartWithRelation(
	permission string,
	relation string,
	kv ...errors.KV,
) error {
	return errors.New(
		fmt.Sprintf(
			errExprOfRequiredPermMustStartWithRelation,
			permission,
			relation,
		),
		kv...,
	)
}

func newErrExprOfRequiredPermissionHasInvalidChar(
	permission string,
	relation string,
	kv ...errors.KV,
) error {
	return errors.New(
		fmt.Sprintf(
			errExprOfRequiredPermHasInvalidChar,
			permission,
			relation,
		),
		kv...,
	)
}
