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

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/sourcehub/sdk"
	"github.com/sourcenetwork/sourcehub/x/acp/did"
	acptypes "github.com/sourcenetwork/sourcehub/x/acp/types"

	"github.com/sourcenetwork/defradb/keyring"
)

type acpSourceHub struct {
	client     *sdk.Client
	txBuilder  *sdk.TxBuilder
	keyring    keyring.Keyring
	acpKeyName string
}

var _ sourceHubClient = (*acpSourceHub)(nil)

func NewACPSourceHub(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	keyring keyring.Keyring,
	acpKeyName string,
) (*acpSourceHub, error) {
	client, err := sdk.NewClient(sdk.WithGRPCAddr(grpcAddress), sdk.WithCometRPCAddr(cometRPCAddress))
	if err != nil {
		return nil, err
	}

	txBuilder, err := sdk.NewTxBuilder(
		sdk.WithSDKClient(client),
		sdk.WithChainID(chainID),
	)
	if err != nil {
		return nil, err
	}

	return &acpSourceHub{
		client:     client,
		txBuilder:  &txBuilder,
		keyring:    keyring,
		acpKeyName: acpKeyName,
	}, nil
}

func (a *acpSourceHub) Init(ctx context.Context, path string) {

}

func (a *acpSourceHub) Start(ctx context.Context) error {
	return nil
}

func (a *acpSourceHub) AddPolicy(
	ctx context.Context,
	creatorID string,
	policy string,
	policyMarshalType acptypes.PolicyMarshalingType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	adminKey, err := a.keyring.Get(a.acpKeyName)
	if err != nil {
		return "", err
	}

	signer := sdk.TxSignerFromCosmosKey(&secp256k1.PrivKey{
		Key: adminKey,
	})

	msgSet := sdk.MsgSet{}
	policyMapper := msgSet.WithCreatePolicy(
		acptypes.NewMsgCreatePolicyNow(creatorID, policy, policyMarshalType),
	)
	tx, err := a.txBuilder.Build(ctx, signer, &msgSet)
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
) (*acptypes.Policy, error) {
	response, err := a.client.ACPQueryClient().Policy(
		ctx,
		&acptypes.QueryPolicyRequest{Id: policyID},
	)
	if err != nil {
		return nil, err
	}

	return response.Policy, nil
}

func (a *acpSourceHub) RegisterObject(
	ctx context.Context,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) (acptypes.RegistrationResult, error) {
	cmdBuilder, err := sdk.NewCmdBuilder(ctx, a.client)
	if err != nil {
		return 0, err
	}

	aliceDid, aliceSigner, err := did.ProduceDID()
	if err != nil {
		return 0, err
	}
	cmdBuilder.RegisterObject(acptypes.NewObject(resourceName, objectID))
	cmdBuilder.PolicyID(policyID)
	cmdBuilder.Actor(aliceDid)
	cmdBuilder.SetSigner(aliceSigner)
	// todo - this jws must be provided by the user, see https://github.com/sourcenetwork/defradb/pull/2649
	jws, err := cmdBuilder.BuildJWS(ctx)
	if err != nil {
		return 0, err
	}

	adminKey, err := a.keyring.Get(a.acpKeyName)
	if err != nil {
		return 0, err
	}

	signer := sdk.TxSignerFromCosmosKey(&secp256k1.PrivKey{
		Key: []byte(adminKey),
	})

	msgSet := sdk.MsgSet{}
	msgSet.WithPolicyCmd(acptypes.NewMsgPolicyCmdFromJWS(signer.GetAccAddress(), jws))
	tx, err := a.txBuilder.Build(ctx, signer, &msgSet)
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

	return acptypes.RegistrationResult(resp.Code), nil // todo: unsure if this return is correct
}

func (a *acpSourceHub) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (*acptypes.QueryObjectOwnerResponse, error) {
	return a.client.ACPQueryClient().ObjectOwner(
		ctx,
		&acptypes.QueryObjectOwnerRequest{
			PolicyId: policyID,
			Object:   acptypes.NewObject(resourceName, objectID),
		},
	)
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
