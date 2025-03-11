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

	"os"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/acp_core/pkg/auth"
	acperrors "github.com/sourcenetwork/acp_core/pkg/errors"
	"github.com/sourcenetwork/acp_core/pkg/runtime"
	"github.com/sourcenetwork/acp_core/pkg/services"
	"github.com/sourcenetwork/acp_core/pkg/types"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/errors"
)

const localACPStoreName = "local_acp"

// ACPLocal represents a local acp implementation that makes no remote calls.
type ACPLocal struct {
	pathToStore immutable.Option[string]
	engine      types.ACPEngineServer
	manager     runtime.RuntimeManager
	closed      bool
}

var _ sourceHubClient = (*ACPLocal)(nil)

func mapACPCorePolicy(pol *types.Policy) policy {
	resources := make(map[string]*resource)
	for _, coreResource := range pol.Resources {
		resource := mapACPCoreResource(coreResource)
		resources[resource.Name] = resource
	}

	return policy{
		ID:        pol.Id,
		Resources: resources,
	}
}

func mapACPCoreResource(policy *types.Resource) *resource {
	perms := make(map[string]*permission)
	for _, corePermission := range policy.Permissions {
		perm := mapACPCorePermission(corePermission)
		perms[perm.Name] = perm
	}

	return &resource{
		Name:        policy.Name,
		Permissions: perms,
	}
}

func mapACPCorePermission(perm *types.Permission) *permission {
	return &permission{
		Name:       perm.Name,
		Expression: perm.Expression,
	}
}

func (l *ACPLocal) Init(ctx context.Context, path string) {
	if path == "" {
		l.pathToStore = immutable.None[string]()
	} else {
		l.pathToStore = immutable.Some(path)
	}
}

func (l *ACPLocal) Start(ctx context.Context) error {
	var manager runtime.RuntimeManager
	var err error
	var opts []runtime.Opt
	var storeLocation string

	l.closed = false
	if !l.pathToStore.HasValue() { // Use a non-persistent, i.e. in memory store.
		storeLocation = "in-memory"
		opts = append(opts, runtime.WithMemKV())
	} else { // Use peristent storage.
		storeLocation = l.pathToStore.Value()
		acpStorePath := storeLocation + "/" + localACPStoreName
		opts = append(opts, runtime.WithPersistentKV(acpStorePath))
	}

	manager, err = runtime.NewRuntimeManager(opts...)
	if err != nil {
		return NewErrInitializationOfACPFailed(err, "Local", storeLocation)
	}

	l.manager = manager
	l.engine = services.NewACPEngine(manager)
	return nil
}

func (l *ACPLocal) Close() error {
	if !l.closed {
		err := l.manager.Terminate()
		if err != nil {
			return err
		}
		l.closed = true
	}
	return nil
}

func (l *ACPLocal) ResetState(ctx context.Context) error {
	err := l.Close()
	if err != nil {
		return err
	}

	// delete state (applicable to persistent store)
	if l.pathToStore.HasValue() {
		storeLocation := l.pathToStore.Value()
		path := storeLocation + "/" + localACPStoreName
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return errors.Join(ErrACPResetState, err)
		} else if err != nil {
			return errors.Join(ErrACPResetState, err)
		}

		if info.IsDir() {
			// remove dir
			if err := os.RemoveAll(path); err != nil {
				return errors.Join(ErrACPResetState, err)
			}
		} else {
			// remove file
			if err := os.Remove(path); err != nil {
				return errors.Join(ErrACPResetState, err)
			}
		}
	}

	// Start again
	return l.Start(ctx)
}

func (l *ACPLocal) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	marshalType policyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	principal, err := types.NewDIDPrincipal(creator.DID)
	if err != nil {
		return "", newErrInvalidActorID(err, creator.DID)
	}
	ctx = auth.InjectPrincipal(ctx, principal)

	createPolicy := types.CreatePolicyRequest{
		Policy:      policy,
		MarshalType: types.PolicyMarshalingType(marshalType),
	}

	response, err := l.engine.CreatePolicy(ctx, &createPolicy)
	if err != nil {
		return "", err
	}

	return response.Record.Policy.Id, nil
}

func (l *ACPLocal) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[policy], error) {
	none := immutable.None[policy]()

	request := types.GetPolicyRequest{Id: policyID}
	response, err := l.engine.GetPolicy(ctx, &request)

	if err != nil {
		if errors.Is(err, acperrors.ErrorType_NOT_FOUND) {
			return none, nil
		}
		return none, err
	}

	policy := mapACPCorePolicy(response.Record.Policy)
	return immutable.Some(policy), nil
}

func (l *ACPLocal) RegisterObject(
	ctx context.Context,
	identity identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	principal, err := types.NewDIDPrincipal(identity.DID)
	if err != nil {
		return newErrInvalidActorID(err, identity.DID)
	}

	ctx = auth.InjectPrincipal(ctx, principal)
	req := types.RegisterObjectRequest{
		PolicyId: policyID,
		Object:   types.NewObject(resourceName, objectID),
	}

	_, err = l.engine.RegisterObject(ctx, &req)
	return err
}

func (l *ACPLocal) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	none := immutable.None[string]()

	req := types.GetObjectRegistrationRequest{
		PolicyId: policyID,
		Object:   types.NewObject(resourceName, objectID),
	}
	result, err := l.engine.GetObjectRegistration(ctx, &req)
	if err != nil {
		return none, err
	}

	if result.IsRegistered {
		return immutable.Some(result.OwnerId), nil
	}

	return none, nil
}

func (l *ACPLocal) VerifyAccessRequest(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	req := types.VerifyAccessRequestRequest{
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
	resp, err := l.engine.VerifyAccessRequest(ctx, &req)

	if err != nil {
		return false, err
	}

	return resp.Valid, nil
}

func (l *ACPLocal) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	principal, err := types.NewDIDPrincipal(requester.DID)
	if err != nil {
		return false, newErrInvalidActorID(err, requester.DID)
	}

	ctx = auth.InjectPrincipal(ctx, principal)

	var newActorRelationship *types.Relationship
	if targetActor == "*" {
		newActorRelationship = types.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = types.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	setRelationshipRequest := types.SetRelationshipRequest{
		PolicyId:     policyID,
		Relationship: newActorRelationship,
	}

	setRelationshipResponse, err := l.engine.SetRelationship(ctx, &setRelationshipRequest)
	if err != nil {
		return false, err
	}

	return setRelationshipResponse.RecordExisted, nil
}

func (l *ACPLocal) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	principal, err := types.NewDIDPrincipal(requester.DID)
	if err != nil {
		return false, newErrInvalidActorID(err, requester.DID)
	}

	ctx = auth.InjectPrincipal(ctx, principal)

	var newActorRelationship *types.Relationship
	if targetActor == "*" {
		newActorRelationship = types.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = types.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	deleteRelationshipRequest := types.DeleteRelationshipRequest{
		PolicyId:     policyID,
		Relationship: newActorRelationship,
	}

	deleteRelationshipResponse, err := l.engine.DeleteRelationship(ctx, &deleteRelationshipRequest)
	if err != nil {
		return false, err
	}

	return deleteRelationshipResponse.RecordFound, nil
}
