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
	"github.com/sourcenetwork/sourcehub/x/acp/embedded"
	"github.com/sourcenetwork/sourcehub/x/acp/types"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

type ACPLocalEmbedded struct {
	pathToStore string
	localModule *embedded.LocalACP
}

var (
	_ ACPModule = (*ACPLocalEmbedded)(nil)
)

// SomeNewLocalACPModule returns an Option with the underlying acp module started (ACPModule.Start() called)
// if successfully initialized, otherwise the `Option.HasValue() == false`.
//
// The caller is responsible to free the resources by calling the `ACPModule.Close()` on the underlying
// acp module if `Option.HasValue() == true`, i.e. the acp module was initialized.
func SomeNewLocalACPModule(ctx context.Context, rootDir string) immutable.Option[ACPModule] {
	var acpModule ACPLocalEmbedded

	if err := acpModule.Start(ctx, rootDir); err != nil {
		return NoACPModule
	}

	return immutable.Some[ACPModule](&acpModule)
}

func (l *ACPLocalEmbedded) Start(ctx context.Context, rootDir string) error {
	acpStorePath := rootDir + "/" + embedded.DefaultDataDir

	var localACPModule embedded.LocalACP
	var err error

	if rootDir == "" { // Use a non-persistent, i.e. in memory store.
		localACPModule, err = embedded.NewLocalACP(
			embedded.WithInMemStore(),
		)
	} else {
		localACPModule, err = embedded.NewLocalACP(
			embedded.WithPersistentStorage(acpStorePath),
		)
	}

	if err != nil {
		log.ErrorE(ctx, "initialization of local acp module failed", err)
		return err
	}

	l.localModule = &localACPModule
	l.pathToStore = acpStorePath

	return nil
}

func (l *ACPLocalEmbedded) Close() error {
	return l.localModule.Close()
}

func (l *ACPLocalEmbedded) AddPolicy(
	ctx context.Context,
	creator string,
	policy string,
	isYAML bool,
) (string, error) {
	var createPolicy types.MsgCreatePolicy

	if isYAML {
		createPolicy = types.MsgCreatePolicy{
			Creator:      creator,
			Policy:       policy,
			MarshalType:  types.PolicyMarshalingType_SHORT_YAML,
			CreationTime: protoTypes.TimestampNow(),
		}
	} else { // use JSON marshal format
		createPolicy = types.MsgCreatePolicy{
			Creator:      creator,
			Policy:       policy,
			MarshalType:  types.PolicyMarshalingType_SHORT_JSON,
			CreationTime: protoTypes.TimestampNow(),
		}
	}

	createPolicyResponse, err := l.localModule.GetMsgService().CreatePolicy(
		l.localModule.GetCtx(),
		&createPolicy,
	)
	if err != nil {
		log.ErrorE(ctx, "failed to add/create policy with local acp module", err)
		return "", err
	}

	policyID := createPolicyResponse.Policy.Id
	log.Info(ctx, "Created Policy", logging.NewKV("PolicyID", policyID))

	return policyID, nil
}

// TODO-ACP: Change name to be ValidateDefraPoilicyInterface()
func (l *ACPLocalEmbedded) ValidatePolicyAndResourceExist(
	ctx context.Context,
	policyID string,
	resource string,
) error {
	if policyID == "" && resource == "" {
		return ErrNoPolicyArgs
	}

	if policyID == "" {
		return ErrPolicyIDMustNotBeEmpty
	}

	if resource == "" {
		return ErrResourceNameMustNotBeEmpty
	}

	queryPolicyRequest := types.QueryPolicyRequest{Id: policyID}
	queryPolicyResponse, err := l.localModule.GetQueryService().Policy(
		l.localModule.GetCtx(),
		&queryPolicyRequest,
	)

	if err != nil {
		if errors.Is(err, types.ErrPolicyNotFound) {
			return newErrPolicyIDValidation(
				err,
				policyID,
				errPolicyDoesNotExistOnACPModule,
			)
		} else {
			return newErrPolicyIDValidation(
				err,
				policyID,
				errPolicyValidationFailedOnACPModule,
			)
		}
	}

	// So far we validated that the policy exists, now lets validate that resource exists.
	resourceResponse := queryPolicyResponse.Policy.GetResourceByName(resource)
	if resourceResponse == nil || resourceResponse.Name != resource {
		return newErrResourceDoesNotExistOnTargetPolicy(err, resource, policyID)
	}

	// Now that we have validated that policyID exists and it contains a corresponding
	// resource with the matching name, validate that all required permissions
	// for DPI actually exist on the target resource.
	for _, requiredPermission := range dpiRequiredPermissions {
		permissionResponse := resourceResponse.GetPermissionByName(requiredPermission)
		if permissionResponse == nil || permissionResponse.Name != requiredPermission {
			return newErrResourceIsMissingRequiredPermission(
				err,
				resource,
				requiredPermission,
				policyID,
			)
		}

		// Now we need to ensure that the "owner" relation has access to all the required
		// permissions for DPI. This is important because even if the policy has the required
		// permissions under the resource, it's possible that those permissions are not granted
		// to the "owner" relation, this will help users not shoot themseleves in the foot.
		// TODO-ACP: Better validation, once sourcehub implements it.
		if err := validateDPIExpressionOfRequiredPermission(
			permissionResponse.Expression,
			requiredPermission,
		); err != nil {
			return err
		}
	}

	return nil
}

func (l *ACPLocalEmbedded) RegisterDocCreation(
	ctx context.Context,
	creator string,
	policyID string,
	resource string,
	docID string,
) error {
	registerDoc := types.MsgRegisterObject{
		Creator:      creator,
		PolicyId:     policyID,
		Object:       types.NewObject(resource, docID),
		CreationTime: protoTypes.TimestampNow(),
	}

	registerDocResponse, err := l.localModule.GetMsgService().RegisterObject(
		l.localModule.GetCtx(),
		&registerDoc,
	)

	if err != nil {
		log.ErrorE(
			ctx,
			"failed to register document with local acp module",
			err,
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("Creator", creator),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return err
	}

	switch registerDocResponse.Result {
	case types.RegistrationResult_NoOp:
		log.Error(
			ctx,
			errObjectDidNotRegister,
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("Creator", creator),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return ErrObjectDidNotRegister

	case types.RegistrationResult_Registered:
		log.Debug(
			ctx,
			"Document registered with local acp module",
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("Creator", creator),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return nil

	case types.RegistrationResult_Unarchived:
		log.Debug(
			ctx,
			"Document re-registered (unarchived object) with local acp module",
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("Creator", creator),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return nil
	}

	return ErrObjectDidNotRegister
}

func (l *ACPLocalEmbedded) IsDocRegistered(
	ctx context.Context,
	policyID string,
	resource string,
	docID string,
) (bool, error) {
	queryObjectOwner := types.QueryObjectOwnerRequest{
		PolicyId: policyID,
		Object:   types.NewObject(resource, docID),
	}

	queryObjectOwnerResponse, err := l.localModule.GetQueryService().ObjectOwner(
		l.localModule.GetCtx(),
		&queryObjectOwner,
	)
	if err != nil {
		log.ErrorE(
			ctx,
			"failed to check if doc is registered alread with local acp module",
			err,
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return false, err
	}

	return queryObjectOwnerResponse.IsRegistered, nil
}

func (l *ACPLocalEmbedded) CheckDocAccess(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resource string,
	docID string,
) (bool, error) {
	checkDoc := types.QueryVerifyAccessRequestRequest{
		PolicyId: policyID,
		AccessRequest: &types.AccessRequest{
			Operations: []*types.Operation{
				{
					Object:     types.NewObject(resource, docID),
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
		log.ErrorE(
			ctx,
			"failed to check/verify doc access with local acp module",
			err,
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("ActorID", actorID),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return false, err
	}

	if checkDocResponse.Valid {
		log.Info(
			ctx,
			"Document accessible",
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("ActorID", actorID),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return true, nil
	} else {
		log.Info(
			ctx,
			"Document inaccessible",
			logging.NewKV("PolicyID", policyID),
			logging.NewKV("ActorID", actorID),
			logging.NewKV("Resource", resource),
			logging.NewKV("DocID", docID),
		)
		return false, nil
	}
}
