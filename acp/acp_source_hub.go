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
	"strings"

	protoTypes "github.com/cosmos/gogoproto/types"
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
		acptypes.NewMsgCreatePolicyNow(a.signer.GetAccAddress(), policy, acptypes.PolicyMarshalingType(policyMarshalType)),
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

	return policyResponse.Policy.Id, nil
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
		if strings.Contains(err.Error(), acptypes.ErrPolicyNotFound.Error()) {
			return immutable.None[policy](), nil
		}

		return immutable.None[policy](), err
	}

	return immutable.Some(
		fromSourceHubPolicy(response.Policy),
	), nil
}

func fromSourceHubPolicy(pol *acptypes.Policy) policy {
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

func fromSourceHubResource(policy *acptypes.Resource) *resource {
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

func fromSourceHubPermission(perm *acptypes.Permission) *permission {
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
) (RegistrationResult, error) {
	msgSet := sourcehub.MsgSet{}
	cmdMapper := msgSet.WithBearerPolicyCmd(&acptypes.MsgBearerPolicyCmd{
		Creator:      a.signer.GetAccAddress(),
		BearerToken:  identity.BearerToken,
		PolicyId:     policyID,
		Cmd:          acptypes.NewRegisterObjectCmd(acptypes.NewObject(resourceName, objectID)),
		CreationTime: creationTime,
	})
	tx, err := a.txBuilder.Build(ctx, a.signer, &msgSet)
	if err != nil {
		return 0, err
	}
	resp, err := a.client.BroadcastTx(ctx, tx)
	if err != nil {
		return 0, err
	}

	result, err := a.client.AwaitTx(ctx, resp.TxHash)
	if err != nil {
		return 0, err
	}
	if result.Error() != nil {
		return 0, result.Error()
	}

	cmdResult, err := cmdMapper.Map(result.TxPayload())
	if err != nil {
		return 0, err
	}

	return RegistrationResult(cmdResult.GetResult().GetRegisterObjectResult().Result), nil
}

func (a *acpSourceHub) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	owner, err := a.client.ACPQueryClient().ObjectOwner(
		ctx,
		&acptypes.QueryObjectOwnerRequest{
			PolicyId: policyID,
			Object:   acptypes.NewObject(resourceName, objectID),
		},
	)
	if err != nil {
		return immutable.None[string](), err
	}

	if owner.OwnerId == "" {
		return immutable.None[string](), nil
	}

	return immutable.Some(owner.OwnerId), nil
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
			AccessRequest: &acptypes.AccessRequest{
				Operations: []*acptypes.Operation{
					{
						Object:     acptypes.NewObject(resourceName, docID),
						Permission: permission.String(),
					},
				},
				Actor: &acptypes.Actor{
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
