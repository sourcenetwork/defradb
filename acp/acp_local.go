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
	"github.com/sourcenetwork/sourcehub/x/acp/embedded"
	"github.com/sourcenetwork/sourcehub/x/acp/types"
	"github.com/valyala/fastjson"

	"github.com/sourcenetwork/defradb/errors"
)

var (
	_ ACPModule = (*ACPLocal)(nil)
)

// ACPLocal represents a local acp module implementation that makes no remote calls.
type ACPLocal struct {
	pathToStore immutable.Option[string]
	localModule *embedded.LocalACP
}

func (l *ACPLocal) Init(ctx context.Context, path string) {
	if path == "" {
		l.pathToStore = immutable.None[string]()
	} else {
		l.pathToStore = immutable.Some(path)
	}
}

func (l *ACPLocal) Start(ctx context.Context) error {
	var localACPModule embedded.LocalACP
	var err error

	if !l.pathToStore.HasValue() { // Use a non-persistent, i.e. in memory store.
		localACPModule, err = embedded.NewLocalACP(
			embedded.WithInMemStore(),
		)

		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", "in-memory")
		}
	} else { // Use peristent storage.
		acpStorePath := l.pathToStore.Value() + "/" + embedded.DefaultDataDir
		localACPModule, err = embedded.NewLocalACP(
			embedded.WithPersistentStorage(acpStorePath),
		)
		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", l.pathToStore.Value())
		}
	}

	l.localModule = &localACPModule
	return nil
}

func (l *ACPLocal) Close() error {
	return l.localModule.Close()
}

func (l *ACPLocal) AddPolicy(
	ctx context.Context,
	creatorID string,
	policy string,
) (string, error) {
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

	createPolicy := types.MsgCreatePolicy{
		Creator:      creatorID,
		Policy:       policy,
		MarshalType:  policyMarshalType,
		CreationTime: protoTypes.TimestampNow(),
	}

	createPolicyResponse, err := l.localModule.GetMsgService().CreatePolicy(
		l.localModule.GetCtx(),
		&createPolicy,
	)

	if err != nil {
		return "", NewErrFailedToAddPolicyWithACP(err, "Local", creatorID)
	}

	policyID := createPolicyResponse.Policy.Id
	log.InfoContext(ctx, "Created Policy", corelog.Any("PolicyID", policyID))

	return policyID, nil
}

func (l *ACPLocal) ValidateResourceExistsOnValidDPI(
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

	queryPolicyRequest := types.QueryPolicyRequest{Id: policyID}
	queryPolicyResponse, err := l.localModule.GetQueryService().Policy(
		l.localModule.GetCtx(),
		&queryPolicyRequest,
	)

	if err != nil {
		if errors.Is(err, types.ErrPolicyNotFound) {
			return newErrPolicyDoesNotExistOnACPModule(err, policyID)
		} else {
			return newErrPolicyValidationFailedOnACPModule(err, policyID)
		}
	}

	// So far we validated that the policy exists, now lets validate that resource exists.
	resourceResponse := queryPolicyResponse.Policy.GetResourceByName(resourceName)
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

func (l *ACPLocal) RegisterDocCreation(
	ctx context.Context,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) error {
	registerDoc := types.MsgRegisterObject{
		Creator:      actorID,
		PolicyId:     policyID,
		Object:       types.NewObject(resourceName, docID),
		CreationTime: protoTypes.TimestampNow(),
	}

	registerDocResponse, err := l.localModule.GetMsgService().RegisterObject(
		l.localModule.GetCtx(),
		&registerDoc,
	)

	if err != nil {
		return NewErrFailedToRegisterDocWithACP(err, "Local", policyID, actorID, resourceName, docID)
	}

	switch registerDocResponse.Result {
	case types.RegistrationResult_NoOp:
		return ErrObjectDidNotRegister

	case types.RegistrationResult_Registered:
		log.InfoContext(
			ctx,
			"Document registered with local acp module",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return nil

	case types.RegistrationResult_Unarchived:
		log.InfoContext(
			ctx,
			"Document re-registered (unarchived object) with local acp module",
			corelog.Any("PolicyID", policyID),
			corelog.Any("Creator", actorID),
			corelog.Any("Resource", resourceName),
			corelog.Any("DocID", docID),
		)
		return nil
	}

	return ErrObjectDidNotRegister
}

func (l *ACPLocal) IsDocRegistered(
	ctx context.Context,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	queryObjectOwner := types.QueryObjectOwnerRequest{
		PolicyId: policyID,
		Object:   types.NewObject(resourceName, docID),
	}

	queryObjectOwnerResponse, err := l.localModule.GetQueryService().ObjectOwner(
		l.localModule.GetCtx(),
		&queryObjectOwner,
	)
	if err != nil {
		return false, NewErrFailedToCheckIfDocIsRegisteredWithACP(err, "Local", policyID, resourceName, docID)
	}

	return queryObjectOwnerResponse.IsRegistered, nil
}

func (l *ACPLocal) CheckDocAccess(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	checkDoc := types.QueryVerifyAccessRequestRequest{
		PolicyId: policyID,
		AccessRequest: &types.AccessRequest{
			Operations: []*types.Operation{
				{
					Object:     types.NewObject(resourceName, docID),
					Permission: permission.String(),
				},
			},
			Actor: &types.Actor{
				Id: actorID,
			},
		},
	}

	checkDocResponse, err := l.localModule.GetQueryService().VerifyAccessRequest(
		l.localModule.GetCtx(),
		&checkDoc,
	)
	if err != nil {
		return false, NewErrFailedToVerifyDocAccessWithACP(err, "Local", policyID, actorID, resourceName, docID)
	}

	if checkDocResponse.Valid {
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
