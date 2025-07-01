// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// SourceHub is not supported in JS environments.
//
//go:build !js

package dac

import (
	"context"
	"fmt"
	"strings"

	protoTypes "github.com/cosmos/gogoproto/types"
	acpErrors "github.com/sourcenetwork/acp_core/pkg/errors"
	coreTypes "github.com/sourcenetwork/acp_core/pkg/types"
	"github.com/sourcenetwork/immutable"
	sourcehub "github.com/sourcenetwork/sourcehub/sdk"
	sourcehubTypes "github.com/sourcenetwork/sourcehub/x/acp/types"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

func NewSourceHubACP(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer sourcehub.TxSigner,
) (DocumentACP, error) {
	acpSourceHub, err := NewACPSourceHub(chainID, grpcAddress, cometRPCAddress, signer)
	if err != nil {
		return nil, err
	}

	return &bridgeDocumentACP{
		clientACP:   acpSourceHub,
		supportsP2P: true,
	}, nil
}

type SourceHubDocumentACP struct {
	client    *sourcehub.Client
	txBuilder *sourcehub.TxBuilder
	signer    sourcehub.TxSigner
}

func NewACPSourceHub(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer sourcehub.TxSigner,
) (*SourceHubDocumentACP, error) {
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
		sourcehub.WithGasLimit(400000),
	)
	if err != nil {
		return nil, err
	}

	return &SourceHubDocumentACP{
		client:    client,
		txBuilder: &txBuilder,
		signer:    signer,
	}, nil
}

func (a *SourceHubDocumentACP) Init(ctx context.Context, path string) {
	// no-op
}

func (a *SourceHubDocumentACP) Start(ctx context.Context) error {
	return nil
}

func (a *SourceHubDocumentACP) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	policyMarshalType acpTypes.PolicyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	msgSet := sourcehub.MsgSet{}
	policyMapper := msgSet.WithCreatePolicy(
		sourcehubTypes.NewMsgCreatePolicy(
			a.signer.GetAccAddress(),
			policy,
			coreTypes.PolicyMarshalingType(policyMarshalType),
		),
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

func (a *SourceHubDocumentACP) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[acpTypes.Policy], error) {
	response, err := a.client.ACPQueryClient().Policy(
		ctx,
		&sourcehubTypes.QueryPolicyRequest{Id: policyID},
	)
	if err != nil {
		// todo: https://github.com/sourcenetwork/defradb/issues/2826
		// Sourcehub errors do not currently work with errors.Is, errors.Is
		// should be used here instead of strings.Contains when that is fixed.
		if strings.Contains(err.Error(), acpErrors.ErrorType_NOT_FOUND.Error()) {
			return immutable.None[acpTypes.Policy](), nil
		}

		return immutable.None[acpTypes.Policy](), err
	}

	return immutable.Some(
		fromSourceHubPolicy(response.Record.Policy),
	), nil
}

func fromSourceHubPolicy(pol *coreTypes.Policy) acpTypes.Policy {
	resources := make(map[string]*acpTypes.Resource)
	for _, coreResource := range pol.Resources {
		resource := fromSourceHubResource(coreResource)
		resources[resource.Name] = resource
	}

	return acpTypes.Policy{
		ID:        pol.Id,
		Resources: resources,
	}
}

func fromSourceHubResource(policy *coreTypes.Resource) *acpTypes.Resource {
	perms := make(map[string]*acpTypes.Permission)
	for _, corePermission := range policy.Permissions {
		perm := fromSourceHubPermission(corePermission)
		perms[perm.Name] = perm
	}

	return &acpTypes.Resource{
		Name:        policy.Name,
		Permissions: perms,
	}
}

func fromSourceHubPermission(perm *coreTypes.Permission) *acpTypes.Permission {
	return &acpTypes.Permission{
		Name:       perm.Name,
		Expression: perm.Expression,
	}
}

func (a *SourceHubDocumentACP) RegisterObject(
	ctx context.Context,
	ident identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	// Check if the identity is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := ident.(identity.TokenIdentity)
	if !ok {
		return fmt.Errorf("identity must be a TokenIdentity to register objects")
	}

	msgSet := sourcehub.MsgSet{}
	cmdMapper := msgSet.WithBearerPolicyCmd(&sourcehubTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         sourcehubTypes.NewRegisterObjectCmd(coreTypes.NewObject(resourceName, objectID)),
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

func (a *SourceHubDocumentACP) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	resp, err := a.client.ACPQueryClient().ObjectOwner(
		ctx,
		&sourcehubTypes.QueryObjectOwnerRequest{
			PolicyId: policyID,
			Object:   coreTypes.NewObject(resourceName, objectID),
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

func (a *SourceHubDocumentACP) VerifyAccessRequest(
	ctx context.Context,
	permission acpTypes.ResourceInterfacePermission,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
) (bool, error) {
	checkDocResponse, err := a.client.ACPQueryClient().VerifyAccessRequest(
		ctx,
		&sourcehubTypes.QueryVerifyAccessRequestRequest{
			PolicyId: policyID,
			AccessRequest: &coreTypes.AccessRequest{
				Operations: []*coreTypes.Operation{
					{
						Object:     coreTypes.NewObject(resourceName, objectID),
						Permission: permission.String(),
					},
				},
				Actor: &coreTypes.Actor{
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

func (a *SourceHubDocumentACP) Close() error {
	return nil
}

func (a *SourceHubDocumentACP) ResetState(_ context.Context) error {
	return fmt.Errorf("sourcehub acp ResetState() unimplemented")
}

func (a *SourceHubDocumentACP) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	// Check if the requester is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := requester.(identity.TokenIdentity)
	if !ok {
		return false, fmt.Errorf("requester must be a TokenIdentity to add actor relationships")
	}

	msgSet := sourcehub.MsgSet{}

	var newActorRelationship *coreTypes.Relationship
	if targetActor == "*" {
		newActorRelationship = coreTypes.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = coreTypes.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	cmdMapper := msgSet.WithBearerPolicyCmd(&sourcehubTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         sourcehubTypes.NewSetRelationshipCmd(newActorRelationship),
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

func (a *SourceHubDocumentACP) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	// Check if the requester is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := requester.(identity.TokenIdentity)
	if !ok {
		return false, fmt.Errorf("requester must be a TokenIdentity to delete actor relationships")
	}

	msgSet := sourcehub.MsgSet{}

	var newActorRelationship *coreTypes.Relationship
	if targetActor == "*" {
		newActorRelationship = coreTypes.NewAllActorsRelationship(
			resourceName,
			objectID,
			relation,
		)
	} else {
		newActorRelationship = coreTypes.NewActorRelationship(
			resourceName,
			objectID,
			relation,
			targetActor,
		)
	}

	cmdMapper := msgSet.WithBearerPolicyCmd(&sourcehubTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         sourcehubTypes.NewDeleteRelationshipCmd(newActorRelationship),
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
