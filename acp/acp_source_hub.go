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
	"fmt"
	"strings"

	protoTypes "github.com/cosmos/gogoproto/types"
	acperrors "github.com/sourcenetwork/acp_core/pkg/errors"
	coretypes "github.com/sourcenetwork/acp_core/pkg/types"
	"github.com/sourcenetwork/immutable"
	sourcehub "github.com/sourcenetwork/sourcehub/sdk"
	acptypes "github.com/sourcenetwork/sourcehub/x/acp/types"

	"github.com/sourcenetwork/defradb/acp/identity"
)

type acpSourceHub struct {
	client    *sourcehub.Client
	txBuilder *sourcehub.TxBuilder
	signer    sourcehub.TxSigner
}

var _ sourceHubClient = (*acpSourceHub)(nil)

func NewACPSourceHub(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer sourcehub.TxSigner,
) (*acpSourceHub, error) {
	client, err := sourcehub.NewClient(
		sourcehub.WithGRPCAddr(grpcAddress),
		sourcehub.WithCometRPCAddr(cometRPCAddress),
	)
	if err != nil {
		return nil, err
	}

	txBuilder, err := sourcehub.NewTxBuilder(
		sourcehub.WithSDKClient(client),
		sourcehub.WithChainID(chainID),
	)
	if err != nil {
		return nil, err
	}

	return &acpSourceHub{
		client:    client,
		txBuilder: &txBuilder,
		signer:    signer,
	}, nil
}

func (a *acpSourceHub) Init(ctx context.Context, path string) {
	// no-op
}

func (a *acpSourceHub) Start(ctx context.Context) error {
	return nil
}

func (a *acpSourceHub) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	policyMarshalType policyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	msgSet := sourcehub.MsgSet{}
	policyMapper := msgSet.WithCreatePolicy(
		acptypes.NewMsgCreatePolicy(a.signer.GetAccAddress(), policy, coretypes.PolicyMarshalingType(policyMarshalType)),
	)
	tx, err := a.txBuilder.Build(ctx, a.signer, &msgSet)
	if err != nil {
		return "", err
	}

	resp, err := a.client.BroadcastTx(ctx, tx)
	if err != nil {
		return "", err
	}

	result, err := a.client.AwaitTx(ctx, resp.TxHash)
	if err != nil {
		return "", err
	}
	if result.Error() != nil {
		return "", result.Error()
	}

	policyResponse, err := policyMapper.Map(result.TxPayload())
	if err != nil {
		return "", err
	}

	return policyResponse.Record.Policy.Id, nil
}

func (a *acpSourceHub) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[policy], error) {
	response, err := a.client.ACPQueryClient().Policy(
		ctx,
		&acptypes.QueryPolicyRequest{Id: policyID},
	)
	if err != nil {
		// todo: https://github.com/sourcenetwork/defradb/issues/2826
		// Sourcehub errors do not currently work with errors.Is, errors.Is
		// should be used here instead of strings.Contains when that is fixed.
		if strings.Contains(err.Error(), acperrors.ErrorType_NOT_FOUND.Error()) {
			return immutable.None[policy](), nil
		}

		return immutable.None[policy](), err
	}

	return immutable.Some(
		fromSourceHubPolicy(response.Record.Policy),
	), nil
}

func fromSourceHubPolicy(pol *coretypes.Policy) policy {
	resources := make(map[string]*resource)
	for _, coreResource := range pol.Resources {
		resource := fromSourceHubResource(coreResource)
		resources[resource.Name] = resource
	}

	return policy{
		ID:        pol.Id,
		Resources: resources,
	}
}

func fromSourceHubResource(policy *coretypes.Resource) *resource {
	perms := make(map[string]*permission)
	for _, corePermission := range policy.Permissions {
		perm := fromSourceHubPermission(corePermission)
		perms[perm.Name] = perm
	}

	return &resource{
		Name:        policy.Name,
		Permissions: perms,
	}
}

func fromSourceHubPermission(perm *coretypes.Permission) *permission {
	return &permission{
		Name:       perm.Name,
		Expression: perm.Expression,
	}
}

func (a *acpSourceHub) RegisterObject(
	ctx context.Context,
	identity identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	msgSet := sourcehub.MsgSet{}
	cmdMapper := msgSet.WithBearerPolicyCmd(&acptypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: identity.BearerToken,
		PolicyId:    policyID,
		Cmd:         acptypes.NewRegisterObjectCmd(coretypes.NewObject(resourceName, objectID)),
	})
	tx, err := a.txBuilder.Build(ctx, a.signer, &msgSet)
	if err != nil {
		return err
	}
	resp, err := a.client.BroadcastTx(ctx, tx)
	if err != nil {
		return err
	}

	result, err := a.client.AwaitTx(ctx, resp.TxHash)
	if err != nil {
		return err
	}
	if result.Error() != nil {
		return result.Error()
	}

	_, err = cmdMapper.Map(result.TxPayload())

	return err
}

func (a *acpSourceHub) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	resp, err := a.client.ACPQueryClient().ObjectOwner(
		ctx,
		&acptypes.QueryObjectOwnerRequest{
			PolicyId: policyID,
			Object:   coretypes.NewObject(resourceName, objectID),
		},
	)
	if err != nil {
		return immutable.None[string](), err
	}

	if !resp.IsRegistered {
		return immutable.None[string](), nil
	}

	return immutable.Some(resp.Record.Metadata.OwnerDid), nil
}

func (a *acpSourceHub) VerifyAccessRequest(
	ctx context.Context,
	permission DPIPermission,
	actorID string,
	policyID string,
	resourceName string,
	docID string,
) (bool, error) {
	checkDocResponse, err := a.client.ACPQueryClient().VerifyAccessRequest(
		ctx,
		&acptypes.QueryVerifyAccessRequestRequest{
			PolicyId: policyID,
			AccessRequest: &coretypes.AccessRequest{
				Operations: []*coretypes.Operation{
					{
						Object:     coretypes.NewObject(resourceName, docID),
						Permission: permission.String(),
					},
				},
				Actor: &coretypes.Actor{
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

func (a *acpSourceHub) Close() error {
	return nil
}

func (a *acpSourceHub) ResetState(_ context.Context) error {
	return fmt.Errorf("sourcehub acp ResetState() unimplemented")
}

func (a *acpSourceHub) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	msgSet := sourcehub.MsgSet{}

	var newActorRelationship *coretypes.Relationship
	if targetActor == "*" {
		newActorRelationship = coretypes.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = coretypes.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	cmdMapper := msgSet.WithBearerPolicyCmd(&acptypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: requester.BearerToken,
		PolicyId:    policyID,
		Cmd:         acptypes.NewSetRelationshipCmd(newActorRelationship),
	})
	tx, err := a.txBuilder.Build(ctx, a.signer, &msgSet)
	if err != nil {
		return false, err
	}
	resp, err := a.client.BroadcastTx(ctx, tx)
	if err != nil {
		return false, err
	}

	result, err := a.client.AwaitTx(ctx, resp.TxHash)
	if err != nil {
		return false, err
	}
	if result.Error() != nil {
		return false, result.Error()
	}

	cmdResult, err := cmdMapper.Map(result.TxPayload())
	if err != nil {
		return false, err
	}

	return cmdResult.GetResult().GetSetRelationshipResult().RecordExisted, nil
}

func (a *acpSourceHub) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	msgSet := sourcehub.MsgSet{}

	var newActorRelationship *coretypes.Relationship
	if targetActor == "*" {
		newActorRelationship = coretypes.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = coretypes.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	cmdMapper := msgSet.WithBearerPolicyCmd(&acptypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: requester.BearerToken,
		PolicyId:    policyID,
		Cmd:         acptypes.NewDeleteRelationshipCmd(newActorRelationship),
	})

	tx, err := a.txBuilder.Build(ctx, a.signer, &msgSet)
	if err != nil {
		return false, err
	}

	resp, err := a.client.BroadcastTx(ctx, tx)
	if err != nil {
		return false, err
	}

	result, err := a.client.AwaitTx(ctx, resp.TxHash)
	if err != nil {
		return false, err
	}

	if result.Error() != nil {
		return false, result.Error()
	}

	cmdResult, err := cmdMapper.Map(result.TxPayload())
	if err != nil {
		return false, err
	}

	return cmdResult.GetResult().GetDeleteRelationshipResult().GetRecordFound(), nil
}
