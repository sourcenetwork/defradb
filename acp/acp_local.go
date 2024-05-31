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
)

// ACPLocal represents a local acp implementation that makes no remote calls.
type ACPLocal struct {
	pathToStore immutable.Option[string]
	localACP    *embedded.LocalACP
}

var _ sourceHubClient = (*ACPLocal)(nil)

func (l *ACPLocal) Init(ctx context.Context, path string) {
	if path == "" {
		l.pathToStore = immutable.None[string]()
	} else {
		l.pathToStore = immutable.Some(path)
	}
}

func (l *ACPLocal) Start(ctx context.Context) error {
	var localACP embedded.LocalACP
	var err error

	if !l.pathToStore.HasValue() { // Use a non-persistent, i.e. in memory store.
		localACP, err = embedded.NewLocalACP(
			embedded.WithInMemStore(),
		)

		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", "in-memory")
		}
	} else { // Use peristent storage.
		acpStorePath := l.pathToStore.Value() + "/" + embedded.DefaultDataDir
		localACP, err = embedded.NewLocalACP(
			embedded.WithPersistentStorage(acpStorePath),
		)
		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", l.pathToStore.Value())
		}
	}

	l.localACP = &localACP
	return nil
}

func (l *ACPLocal) Close() error {
	return l.localACP.Close()
}

func (l *ACPLocal) AddPolicy(
	ctx context.Context,
	creatorID string,
	policy string,
	policyMarshalType types.PolicyMarshalingType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	createPolicy := types.MsgCreatePolicy{
		Creator:      creatorID,
		Policy:       policy,
		MarshalType:  policyMarshalType,
		CreationTime: protoTypes.TimestampNow(),
	}

	createPolicyResponse, err := l.localACP.GetMsgService().CreatePolicy(
		l.localACP.GetCtx(),
		&createPolicy,
	)
	if err != nil {
		return "", err
	}

	return createPolicyResponse.Policy.Id, nil
}

func (l *ACPLocal) Policy(
	ctx context.Context,
	policyID string,
) (*types.Policy, error) {
	queryPolicyResponse, err := l.localACP.GetQueryService().Policy(
		l.localACP.GetCtx(),
		&types.QueryPolicyRequest{Id: policyID},
	)
	if err != nil {
		return nil, err
	}

	return queryPolicyResponse.Policy, nil
}

func (l *ACPLocal) RegisterObject(
	ctx context.Context,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) (types.RegistrationResult, error) {
	registerDocResponse, err := l.localACP.GetMsgService().RegisterObject(
		l.localACP.GetCtx(),
		&types.MsgRegisterObject{
			Creator:      actorID,
			PolicyId:     policyID,
			Object:       types.NewObject(resourceName, objectID),
			CreationTime: creationTime,
		},
	)
	if err != nil {
		return types.RegistrationResult(0), err
	}

	return registerDocResponse.Result, nil
}

func (l *ACPLocal) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (*types.QueryObjectOwnerResponse, error) {
	return l.localACP.GetQueryService().ObjectOwner(
		l.localACP.GetCtx(),
		&types.QueryObjectOwnerRequest{
			PolicyId: policyID,
			Object:   types.NewObject(resourceName, objectID),
		},
	)
}

func (l *ACPLocal) VerifyAccessRequest(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	checkDocResponse, err := l.localACP.GetQueryService().VerifyAccessRequest(
		l.localACP.GetCtx(),
		&types.QueryVerifyAccessRequestRequest{
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
		},
	)
	if err != nil {
		return false, err
	}

	return checkDocResponse.Valid, nil
}
