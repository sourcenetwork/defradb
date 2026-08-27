// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Remote DAC implementation for non-JS environments.
// JS environments are handled by remote_js.go.
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
	vera "github.com/sourcenetwork/vera/sdk"
	veraTypes "github.com/sourcenetwork/vera/x/acp/types"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

// NewRemoteDocumentACP returns a Vera-backed Remote DAC instance.
func NewRemoteDocumentACP(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer vera.TxSigner,
) (DocumentACP, error) {
	remoteDAC, err := NewRemoteDocumentACPClient(chainID, grpcAddress, cometRPCAddress, signer)
	if err != nil {
		return nil, err
	}

	return &bridgeDocumentACP{
		clientACP:       remoteDAC,
		documentACPType: acpTypes.RemoteDocumentACP,
	}, nil
}

// RemoteDocumentACP is the Vera-backed client used by Remote DAC.
type RemoteDocumentACP struct {
	client    *vera.Client
	txBuilder *vera.TxBuilder
	signer    vera.TxSigner
}

// NewRemoteDocumentACPClient returns a Vera-backed Remote DAC client.
func NewRemoteDocumentACPClient(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer vera.TxSigner,
) (*RemoteDocumentACP, error) {
	client, err := vera.NewClient(
		vera.WithGRPCAddr(grpcAddress),
		vera.WithCometRPCAddr(cometRPCAddress),
	)
	if err != nil {
		return nil, err
	}

	txBuilder, err := vera.NewTxBuilder(
		vera.WithSDKClient(client),
		vera.WithChainID(chainID),
		vera.WithGasLimit(400000),
	)
	if err != nil {
		return nil, err
	}

	return &RemoteDocumentACP{
		client:    client,
		txBuilder: &txBuilder,
		signer:    signer,
	}, nil
}

func (a *RemoteDocumentACP) Start(ctx context.Context) error {
	return nil
}

func (a *RemoteDocumentACP) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	policyMarshalType acpTypes.PolicyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	msgSet := vera.MsgSet{}
	policyMapper := msgSet.WithCreatePolicy(
		veraTypes.NewMsgCreatePolicy(
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

func (a *RemoteDocumentACP) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[acpTypes.Policy], error) {
	response, err := a.client.ACPQueryClient().Policy(
		ctx,
		&veraTypes.QueryPolicyRequest{Id: policyID},
	)
	if err != nil {
		// todo: https://github.com/sourcenetwork/defradb/issues/2826
		// Vera errors do not currently work with errors.Is, so errors.Is
		// should be used here instead of strings.Contains when that is fixed.
		if strings.Contains(err.Error(), acpErrors.ErrorType_NOT_FOUND.Error()) {
			return immutable.None[acpTypes.Policy](), nil
		}

		return immutable.None[acpTypes.Policy](), err
	}

	return immutable.Some(
		fromRemoteDACPolicy(response.Record.Policy),
	), nil
}

func fromRemoteDACPolicy(pol *coreTypes.Policy) acpTypes.Policy {
	resources := make(map[string]*acpTypes.Resource)
	for _, coreResource := range pol.Resources {
		resource := fromRemoteDACResource(coreResource)
		resources[resource.Name] = resource
	}

	return acpTypes.Policy{
		ID:        pol.Id,
		Resources: resources,
	}
}

func fromRemoteDACResource(policy *coreTypes.Resource) *acpTypes.Resource {
	perms := make(map[string]*acpTypes.Permission)
	for _, corePermission := range policy.Permissions {
		perm := fromRemoteDACPermission(corePermission)
		perms[perm.Name] = perm
	}

	return &acpTypes.Resource{
		Name:        policy.Name,
		Permissions: perms,
	}
}

func fromRemoteDACPermission(perm *coreTypes.Permission) *acpTypes.Permission {
	return &acpTypes.Permission{
		Name:       perm.Name,
		Expression: perm.Expression,
	}
}

func (a *RemoteDocumentACP) RegisterObject(
	ctx context.Context,
	ident identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	objectID = remoteDACObjectID(objectID)

	// Check if the identity is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := ident.(identity.TokenIdentity)
	if !ok {
		return identity.ErrMustBeTokenIdentity
	}

	msgSet := vera.MsgSet{}
	cmdMapper := msgSet.WithBearerPolicyCmd(&veraTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         veraTypes.NewRegisterObjectCmd(coreTypes.NewObject(resourceName, objectID)),
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

func (a *RemoteDocumentACP) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	objectID = remoteDACObjectID(objectID)

	resp, err := a.client.ACPQueryClient().ObjectOwner(
		ctx,
		&veraTypes.QueryObjectOwnerRequest{
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

func (a *RemoteDocumentACP) VerifyAccessRequest(
	ctx context.Context,
	permission acpTypes.ResourceInterfacePermission,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
) (bool, error) {
	objectID = remoteDACObjectID(objectID)

	checkDocResponse, err := a.client.ACPQueryClient().VerifyAccessRequest(
		ctx,
		&veraTypes.QueryVerifyAccessRequestRequest{
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

func (a *RemoteDocumentACP) Close() error {
	return nil
}

func (a *RemoteDocumentACP) ResetState(_ context.Context) error {
	return fmt.Errorf("remote DAC ResetState() is not implemented")
}

func (a *RemoteDocumentACP) AddActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	objectID = remoteDACObjectID(objectID)

	// Check if the requester is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := requester.(identity.TokenIdentity)
	if !ok {
		return false, identity.ErrMustBeTokenIdentity
	}

	msgSet := vera.MsgSet{}

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

	cmdMapper := msgSet.WithBearerPolicyCmd(&veraTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         veraTypes.NewSetRelationshipCmd(newActorRelationship),
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

	return cmdResult.GetResult().GetSetRelationshipResult().GetRecordExisted(), nil
}

func (a *RemoteDocumentACP) DeleteActorRelationship(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
	relation string,
	requester identity.Identity,
	targetActor string,
	creationTime *protoTypes.Timestamp,
) (bool, error) {
	objectID = remoteDACObjectID(objectID)

	// Check if the requester is a TokenIdentity (has BearerToken)
	tokenIdentity, ok := requester.(identity.TokenIdentity)
	if !ok {
		return false, identity.ErrMustBeTokenIdentity
	}

	msgSet := vera.MsgSet{}

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

	cmdMapper := msgSet.WithBearerPolicyCmd(&veraTypes.MsgBearerPolicyCmd{
		Creator:     a.signer.GetAccAddress(),
		BearerToken: tokenIdentity.BearerToken(),
		PolicyId:    policyID,
		Cmd:         veraTypes.NewDeleteRelationshipCmd(newActorRelationship),
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
