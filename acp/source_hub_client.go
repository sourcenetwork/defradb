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
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/sourcehub/x/acp/types"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/errors"
)

// sourcehubClient is a private abstraction to allow multiple ACP implementations
// based off of the SourceHub libraries to share the same Defra-specific logic via the
// sourceHubBridge.
type sourcehubClient interface {
	// Init initializes the acp, with an absolute path. The provided path indicates where the
	// persistent data will be stored for acp.
	//
	// If the path is empty then acp will run in memory.
	Init(ctx context.Context, path string)

	// Start starts the acp, using the initialized path. Will recover acp state
	// from a previous run if under the same path.
	//
	// If the path is empty then acp will run in memory.
	Start(ctx context.Context) error

	// AddPolicy attempts to add the given policy. Upon success a policyID is returned,
	// otherwise returns error.
	AddPolicy(
		ctx context.Context,
		creatorID string,
		policy string,
		policyMarshalingType types.PolicyMarshalingType,
		creationTime *protoTypes.Timestamp,
	) (string, error)

	// Policy returns a policy of the given policyID if one is found.
	Policy(
		ctx context.Context,
		policyID string,
	) (*types.Policy, error)

	// RegisterObject registers the object to have access control.
	// No error is returned upon successful registering of an object.
	RegisterObject(
		ctx context.Context,
		actorID string,
		policyID string,
		resourceName string,
		objectID string,
		creationTime *protoTypes.Timestamp,
	) (types.RegistrationResult, error)

	// ObjectOwner returns the owner of the object of the given objectID.
	ObjectOwner(
		ctx context.Context,
		policyID string,
		resourceName string,
		objectID string,
	) (*types.QueryObjectOwnerResponse, error)

	// VerifyAccessRequest returns true if the check was successfull and the request has access to the object. If
	// the check was successful but the request does not have access to the object, then returns false.
	// Otherwise if check failed then an error is returned (and the boolean result should not be used).
	VerifyAccessRequest(
		ctx context.Context,
		permission DPIPermission,
		actorID string,
		policyID string,
		resourceName string,
		docID string,
	) (bool, error)

	// Close closes any resources in use by acp.
	Close() error
}

// sourceHubBridge wraps a sourcehubClient, hosting the Defra-specific logic away from client-specific
// code.
type sourceHubBridge struct {
	client sourcehubClient
}

var _ ACP = (*sourceHubBridge)(nil)

func NewLocalACP() ACP {
	return &sourceHubBridge{
		client: &ACPLocal{},
	}
}

func (a *sourceHubBridge) Init(ctx context.Context, path string) {
	a.client.Init(ctx, path)
}

func (a *sourceHubBridge) Start(ctx context.Context) error {
	return a.client.Start(ctx)
}

func (a *sourceHubBridge) AddPolicy(ctx context.Context, creatorID string, policy string) (string, error) {
	// Having a creator identity is a MUST requirement for adding a policy.
	if creatorID == "" {
		return "", ErrPolicyCreatorMustNotBeEmpty
	}

	if policy == "" {
		return "", ErrPolicyDataMustNotBeEmpty
	}

	// Assume policy is in YAML format by default.
	policyMarshalType := types.PolicyMarshalingType_SHORT_YAML
	if isJSON := fastjson.Validate(policy) == nil; isJSON { // Detect JSON format.
		policyMarshalType = types.PolicyMarshalingType_SHORT_JSON
	}

	policyID, err := a.client.AddPolicy(
		ctx,
		creatorID,
		policy,
		policyMarshalType,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return "", NewErrFailedToAddPolicyWithACP(err, "Local", creatorID)
	}

	log.InfoContext(ctx, "Created Policy", corelog.Any("PolicyID", policyID))

	return policyID, nil
}

func (a *sourceHubBridge) ValidateResourceExistsOnValidDPI(
	ctx context.Context,
	policyID string,
	resourceName string,
) error {
	if policyID == "" && resourceName == "" {
		return ErrNoPolicyArgs
	}

	if policyID == "" {
		return ErrPolicyIDMustNotBeEmpty
	}

	if resourceName == "" {
		return ErrResourceNameMustNotBeEmpty
	}

	policy, err := a.client.Policy(ctx, policyID)

	if err != nil {
		if errors.Is(err, types.ErrPolicyNotFound) {
			return newErrPolicyDoesNotExistWithACP(err, policyID)
		} else {
			return newErrPolicyValidationFailedWithACP(err, policyID)
		}
	}

	// So far we validated that the policy exists, now lets validate that resource exists.
	resourceResponse := policy.GetResourceByName(resourceName)
	if resourceResponse == nil {
		return newErrResourceDoesNotExistOnTargetPolicy(resourceName, policyID)
	}

	// Now that we have validated that policyID exists and it contains a corresponding
	// resource with the matching name, validate that all required permissions
	// for DPI actually exist on the target resource.
	for _, requiredPermission := range dpiRequiredPermissions {
		permissionResponse := resourceResponse.GetPermissionByName(requiredPermission)
		if permissionResponse == nil {
			return newErrResourceIsMissingRequiredPermission(
				resourceName,
				requiredPermission,
				policyID,
			)
		}

		// Now we need to ensure that the "owner" relation has access to all the required
		// permissions for DPI. This is important because even if the policy has the required
		// permissions under the resource, it's possible that those permissions are not granted
		// to the "owner" relation, this will help users not shoot themseleves in the foot.
		// TODO-ACP: Better validation, once sourcehub implements meta-policies.
		// Issue: https://github.com/sourcenetwork/defradb/issues/2359
		if err := validateDPIExpressionOfRequiredPermission(
			permissionResponse.Expression,
			requiredPermission,
		); err != nil {
			return err
		}
	}

	return nil
}

func (a *sourceHubBridge) RegisterDocObject(
	ctx context.Context,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) error {
	registerDocResult, err := a.client.RegisterObject(
		ctx,
		actorID,
		policyID,
		resourceName,
		docID,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return NewErrFailedToRegisterDocWithACP(err, "Local", policyID, actorID, resourceName, docID)
	}

	switch registerDocResult {
	case types.RegistrationResult_NoOp:
		return ErrObjectDidNotRegister

	case types.RegistrationResult_Registered:
		log.InfoContext(
			ctx,
			"Document registered with local acp",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return nil

	case types.RegistrationResult_Unarchived:
		log.InfoContext(
			ctx,
			"Document re-registered (unarchived object) with local acp",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return nil
	}

	return ErrObjectDidNotRegister
}

func (a *sourceHubBridge) IsDocRegistered(
	ctx context.Context,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	queryObjectOwnerResponse, err := a.client.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, NewErrFailedToCheckIfDocIsRegisteredWithACP(err, "Local", policyID, resourceName, docID)
	}

	return queryObjectOwnerResponse.IsRegistered, nil
}

func (a *sourceHubBridge) CheckDocAccess(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	isValid, err := a.client.VerifyAccessRequest(
		ctx,
		permission,
		actorID,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, NewErrFailedToVerifyDocAccessWithACP(err, "Local", policyID, actorID, resourceName, docID)
	}

	if isValid {
		log.InfoContext(
			ctx,
			"Document accessible",
			corelog.Any("PolicyID", policyID),
			corelog.Any("ActorID", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return true, nil
	} else {
		log.InfoContext(
			ctx,
			"Document inaccessible",
			corelog.Any("PolicyID", policyID),
			corelog.Any("ActorID", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return false, nil
	}
}

func (a *sourceHubBridge) Close() error {
	return a.client.Close()
}
