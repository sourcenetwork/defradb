// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package dac

import (
	"context"

	"os"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/acp_core/pkg/auth"
	acpErrors "github.com/sourcenetwork/acp_core/pkg/errors"
	"github.com/sourcenetwork/acp_core/pkg/runtime"
	"github.com/sourcenetwork/acp_core/pkg/services"
	"github.com/sourcenetwork/acp_core/pkg/types"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/errors"
)

const localACPStoreName = "local_acp"

var _ DocumentACP = (*bridgeDocumentACP)(nil)

func NewLocalDocumentACP() DocumentACP {
	return &bridgeDocumentACP{
		clientACP:   &LocalDocumentACP{},
		supportsP2P: false,
	}
}

// LocalDocumentACP represents a local document acp implementation that makes no remote calls.
type LocalDocumentACP struct {
	pathToStore immutable.Option[string]
	engine      types.ACPEngineServer
	manager     runtime.RuntimeManager
	closed      bool
}

var _ acp.ACPSystemClient = (*LocalDocumentACP)(nil)

func (l *LocalDocumentACP) Init(ctx context.Context, path string) {
	if path == "" {
		l.pathToStore = immutable.None[string]()
	} else {
		l.pathToStore = immutable.Some(path)
	}
}

func (l *LocalDocumentACP) Start(ctx context.Context) error {
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
		return acp.NewErrInitializationOfACPFailed(err, "Local", storeLocation)
	}

	l.manager = manager
	l.engine = services.NewACPEngine(manager)
	return nil
}

func (l *LocalDocumentACP) Close() error {
	if !l.closed {
		err := l.manager.Terminate()
		if err != nil {
			return err
		}
		l.closed = true
	}
	return nil
}

func (l *LocalDocumentACP) ResetState(ctx context.Context) error {
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
			return errors.Join(acp.ErrACPResetState, err)
		} else if err != nil {
			return errors.Join(acp.ErrACPResetState, err)
		}

		if info.IsDir() {
			// remove dir
			if err := os.RemoveAll(path); err != nil {
				return errors.Join(acp.ErrACPResetState, err)
			}
		} else {
			// remove file
			if err := os.Remove(path); err != nil {
				return errors.Join(acp.ErrACPResetState, err)
			}
		}
	}

	// Start again
	return l.Start(ctx)
}

func (l *LocalDocumentACP) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	marshalType acpTypes.PolicyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	principal, err := types.NewDIDPrincipal(creator.DID())
	if err != nil {
		return "", acp.NewErrInvalidActorID(err, creator.DID())
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

func (l *LocalDocumentACP) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[acpTypes.Policy], error) {
	none := immutable.None[acpTypes.Policy]()

	request := types.GetPolicyRequest{Id: policyID}
	response, err := l.engine.GetPolicy(ctx, &request)

	if err != nil {
		if errors.Is(err, acpErrors.ErrorType_NOT_FOUND) {
			return none, nil
		}
		return none, err
	}

	policy := acpTypes.MapACPCorePolicy(response.Record.Policy)
	return immutable.Some(policy), nil
}

func (l *LocalDocumentACP) RegisterObject(
	ctx context.Context,
	identity identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	principal, err := types.NewDIDPrincipal(identity.DID())
	if err != nil {
		return acp.NewErrInvalidActorID(err, identity.DID())
	}

	ctx = auth.InjectPrincipal(ctx, principal)
	req := types.RegisterObjectRequest{
		PolicyId: policyID,
		Object:   types.NewObject(resourceName, objectID),
	}

	_, err = l.engine.RegisterObject(ctx, &req)
	return err
}

func (l *LocalDocumentACP) ObjectOwner(
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

func (l *LocalDocumentACP) VerifyAccessRequest(
	ctx context.Context,
	permission acpTypes.ResourceInterfacePermission,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
) (bool, error) {
	req := types.VerifyAccessRequestRequest{
		PolicyId: policyID,
		AccessRequest: &types.AccessRequest{
			Operations: []*types.Operation{
				{
					Object:     types.NewObject(resourceName, objectID),
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

func (l *LocalDocumentACP) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	principal, err := types.NewDIDPrincipal(requester.DID())
	if err != nil {
		return false, acp.NewErrInvalidActorID(err, requester.DID())
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

func (l *LocalDocumentACP) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	principal, err := types.NewDIDPrincipal(requester.DID())
	if err != nil {
		return false, acp.NewErrInvalidActorID(err, requester.DID())
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
