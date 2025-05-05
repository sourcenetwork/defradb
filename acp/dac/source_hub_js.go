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
//go:build js

package dac

import (
	"context"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/immutable"
)

// SourceHub is not supported in JS environments so these mocks are needed
// to avoid undeclared errors while building when source_hub types are
// missing due to go build ignoring the source_hub implementation.

type SourceHubDocumentACP struct{}

func (a *SourceHubDocumentACP) Init(ctx context.Context, path string) {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) Start(ctx context.Context) error {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) AddPolicy(
	ctx context.Context,
	creator identity.Identity,
	policy string,
	policyMarshalType acpTypes.PolicyMarshalType,
	creationTime *protoTypes.Timestamp,
) (string, error) {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) Policy(
	ctx context.Context,
	policyID string,
) (immutable.Option[acpTypes.Policy], error) {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) RegisterObject(
	ctx context.Context,
	identity identity.Identity,
	policyID string,
	resourceName string,
	objectID string,
	creationTime *protoTypes.Timestamp,
) error {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) ObjectOwner(
	ctx context.Context,
	policyID string,
	resourceName string,
	objectID string,
) (immutable.Option[string], error) {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) VerifyAccessRequest(
	ctx context.Context,
	permission acpTypes.ResourceInterfacePermission,
	actorID string,
	policyID string,
	resourceName string,
	objectID string,
) (bool, error) {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) Close() error {
	panic("warning js build does not support sourcehub")
}

func (a *SourceHubDocumentACP) ResetState(_ context.Context) error {
	panic("warning js build does not support sourcehub")
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
	panic("warning js build does not support sourcehub")
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
	panic("warning js build does not support sourcehub")
}
