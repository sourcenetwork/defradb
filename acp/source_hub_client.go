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
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/sourcehub/sdk"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// sourceHubClient is a private abstraction to allow multiple ACP implementations
// based off of the SourceHub libraries to share the same Defra-specific logic via the
// sourceHubBridge.
type sourceHubClient interface {
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
		creator identity.Identity,
		policy string,
		marshalType policyMarshalType,
		creationTime *protoTypes.Timestamp,
	) (string, error)

	// Policy returns a policy of the given policyID if one is found.
	Policy(
		ctx context.Context,
		policyID string,
	) (immutable.Option[policy], error)

	// RegisterObject registers the object to have access control.
	// No error is returned upon successful registering of an object.
	RegisterObject(
		ctx context.Context,
		identity identity.Identity,
		policyID string,
		resourceName string,
		objectID string,
		creationTime *protoTypes.Timestamp,
	) (RegistrationResult, error)

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
		permission DPIPermission,
		actorID string,
		policyID string,
		resourceName string,
		docID string,
	) (bool, error)

	// Close closes any resources in use by acp.
	Close() error
}

// sourceHubBridge wraps a sourceHubClient, hosting the Defra-specific logic away from client-specific
// code.
type sourceHubBridge struct {
	client sourceHubClient
}

var _ ACP = (*sourceHubBridge)(nil)

func NewLocalACP() ACP {
	return &sourceHubBridge{
		client: &ACPLocal{},
	}
}

func NewSourceHubACP(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer sdk.TxSigner,
) (ACP, error) {
	acpSourceHub, err := NewACPSourceHub(chainID, grpcAddress, cometRPCAddress, signer)
	if err != nil {
		return nil, err
	}

	return &sourceHubBridge{
		client: acpSourceHub,
	}, nil
}

func (a *sourceHubBridge) Init(ctx context.Context, path string) {
	a.client.Init(ctx, path)
}

func (a *sourceHubBridge) Start(ctx context.Context) error {
	return a.client.Start(ctx)
}

func (a *sourceHubBridge) AddPolicy(ctx context.Context, creator identity.Identity, policy string) (string, error) {
	// Having a creator identity is a MUST requirement for adding a policy.
	if creator.DID == "" {
		return "", ErrPolicyCreatorMustNotBeEmpty
	}

	if policy == "" {
		return "", ErrPolicyDataMustNotBeEmpty
	}

	marshalType := policyMarshalType_YAML
	if isJSON := fastjson.Validate(policy) == nil; isJSON { // Detect JSON format.
		marshalType = policyMarshalType_JSON
	}

	policyID, err := a.client.AddPolicy(
		ctx,
		creator,
		policy,
		marshalType,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return "", NewErrFailedToAddPolicyWithACP(err, "Local", creator.DID)
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

	maybePolicy, err := a.client.Policy(ctx, policyID)

	if err != nil {
		return newErrPolicyValidationFailedWithACP(err, policyID)
	}
	if !maybePolicy.HasValue() {
		return newErrPolicyDoesNotExistWithACP(err, policyID)
	}

	policy := maybePolicy.Value()

	// So far we validated that the policy exists, now lets validate that resource exists.
	resourceResponse, ok := policy.Resources[resourceName]
	if !ok {
		return newErrResourceDoesNotExistOnTargetPolicy(resourceName, policyID)
	}

	// Now that we have validated that policyID exists and it contains a corresponding
	// resource with the matching name, validate that all required permissions
	// for DPI actually exist on the target resource.
	for _, requiredPermission := range dpiRequiredPermissions {
		permissionResponse, ok := resourceResponse.Permissions[requiredPermission]
		if !ok {
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
	identity identity.Identity,
	policyID string,
	resourceName string,
	docID string,
) error {
	registerDocResult, err := a.client.RegisterObject(
		ctx,
		identity,
		policyID,
		resourceName,
		docID,
		protoTypes.TimestampNow(),
	)

	if err != nil {
		return NewErrFailedToRegisterDocWithACP(err, "Local", policyID, identity.DID, resourceName, docID)
	}

	switch registerDocResult {
	case RegistrationResult_NoOp:
		return ErrObjectDidNotRegister

	case RegistrationResult_Registered:
		log.InfoContext(
			ctx,
			"Document registered with local acp",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", identity.DID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return nil

	case RegistrationResult_Unarchived:
		log.InfoContext(
			ctx,
			"Document re-registered (unarchived object) with local acp",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", identity.DID),
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
	maybeActor, err := a.client.ObjectOwner(
		ctx,
		policyID,
		resourceName,
		docID,
	)
	if err != nil {
		return false, NewErrFailedToCheckIfDocIsRegisteredWithACP(err, "Local", policyID, resourceName, docID)
	}

	return maybeActor.HasValue(), nil
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
