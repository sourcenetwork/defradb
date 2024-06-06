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
	"crypto/ed25519"
	"errors"
	"strings"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/acp_core/pkg/auth"
	"github.com/sourcenetwork/acp_core/pkg/did"
	"github.com/sourcenetwork/acp_core/pkg/engine"
	"github.com/sourcenetwork/acp_core/pkg/runtime"
	"github.com/sourcenetwork/acp_core/pkg/types"
	"github.com/sourcenetwork/immutable"
	"github.com/valyala/fastjson"
)

const localACPStoreName = "local_acp"

// ACPLocal represents a local acp implementation that makes no remote calls.
type ACPLocal struct {
	pathToStore immutable.Option[string]
	engine      types.ACPEngineServer
	manager     runtime.RuntimeManager
}

var _ sourceHubClient = (*ACPLocal)(nil)

var errGeneratingDIDFromNonAccAddr = errors.New("cannot generate did if address is not prefixed")

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

	if !l.pathToStore.HasValue() { // Use a non-persistent, i.e. in memory store.
		manager, err = runtime.NewRuntimeManager(
			runtime.WithMemKV(),
		)

		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", "in-memory")
		}
	} else { // Use peristent storage.
		acpStorePath := l.pathToStore.Value() + "/" + localACPStoreName
		manager, err = runtime.NewRuntimeManager(
			runtime.WithPersistentKV(acpStorePath),
		)

		if err != nil {
			return NewErrInitializationOfACPFailed(err, "Local", l.pathToStore.Value())
		}
	}

	engine := engine.NewACPEngine(manager)
	l.engine = engine
	l.manager = manager
	return nil
}

func (l *ACPLocal) Close() error {
	return l.manager.Terminate()
}

func (l *ACPLocal) AddPolicy(
	ctx context.Context,
	creatorID string,
	policy string,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	// FIXME remove once Identity is refactored
	did, err := genDIDFromSourceHubAddr(creatorID)
	if err != nil {
		return "", err
	}

	principal, err := auth.NewDIDPrincipal(did)
	if err != nil {
		return "", newErrInvalidActorID(err, creatorID)
	}
	ctx = auth.InjectPrincipal(ctx, principal)

	marshalType := types.PolicyMarshalingType_SHORT_YAML
	if isJSON := fastjson.Validate(policy) == nil; isJSON { // Detect JSON format.
		marshalType = types.PolicyMarshalingType_SHORT_JSON
	}

	createPolicy := types.CreatePolicyRequest{
		Policy:       policy,
		MarshalType:  marshalType,
		CreationTime: protoTypes.TimestampNow(),
	}

	response, err := l.engine.CreatePolicy(ctx, &createPolicy)

	if err != nil {
		return "", err
	}

	return response.Policy.Id, nil
}

func (l *ACPLocal) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[policy], error) {
	none := immutable.None[policy]()

	request := types.GetPolicyRequest{Id: policyID}
	response, err := l.engine.GetPolicy(ctx, &request)

	if err != nil {
		if errors.Is(err, types.ErrPolicyNotFound) {
			return none, nil
		}
		return none, err
	}

	policy := mapACPCorePolicy(response.Policy)
	return immutable.Some(policy), nil
}

func (l *ACPLocal) RegisterObject(
	ctx context.Context,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) (RegistrationResult, error) {
	// FIXME remove once Identity is refactored
	did, err := genDIDFromSourceHubAddr(actorID)
	if err != nil {
		return RegistrationResult_NoOp, err
	}

	principal, err := auth.NewDIDPrincipal(did)
	if err != nil {
		return RegistrationResult_NoOp, newErrInvalidActorID(err, actorID)
	}

	ctx = auth.InjectPrincipal(ctx, principal)
	req := types.RegisterObjectRequest{
		PolicyId:     policyID,
		Object:       types.NewObject(resourceName, objectID),
		CreationTime: creationTime,
	}

	registerDocResponse, err := l.engine.RegisterObject(ctx, &req)

	if err != nil {
		return RegistrationResult_NoOp, err
	}

	result := RegistrationResult(registerDocResponse.Result)
	return result, nil
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
	// FIXME remove once Identity is refactored
	did, err := genDIDFromSourceHubAddr(actorID)
	if err != nil {
		return false, err
	}
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
				Id: did,
			},
		},
	}
	resp, err := l.engine.VerifyAccessRequest(ctx, &req)

	if err != nil {
		return false, err
	}

	return resp.Valid, nil
}

// genDIDFromSourceHubAddr uses an account addr as a seed to produce a key pair
// and consequently generate a DID.
//
// NOTE: This is by no means a *safe* practice, however it's "okay" for two reasons:
//  1. It's a temporary workaround which will be invalidated once the new identity system
//     is in place (ie. Identity is a DID as opposed to a SourceHub Addr)
//  2. In Local ACP, the the temporary keys used to generate the DID aren't effectively
//     used for any cryptographic operations.
//
// This method will produce an error if `addr` does not begin with "source".
// The error will ensure that the tests break after the identity system is refactored,
// which will be a sign that this method can be deleted entirely
func genDIDFromSourceHubAddr(addr string) (string, error) {
	if !strings.HasPrefix(addr, "source") {
		return "", errGeneratingDIDFromNonAccAddr
	}

	seed := make([]byte, ed25519.SeedSize)
	copy(seed, []byte(addr))
	did, _, err := did.ProduceDIDFromSeed(seed)
	if err != nil {
		return "", err
	}
	return did, nil
}
