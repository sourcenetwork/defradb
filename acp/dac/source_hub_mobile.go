// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// SourceHub ACP stub for mobile (Android/iOS) environments.
// SourceHub ACP requires the Cosmos SDK keyring which is incompatible with mobile platforms.
//
//go:build android || ios

package dac

import (
	"context"
	"errors"

	protoTypes "github.com/cosmos/gogoproto/types"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
)

// ErrSourceHubNotSupportedOnMobile is returned when SourceHub ACP is used on mobile.
var ErrSourceHubNotSupportedOnMobile = errors.New("SourceHub ACP is not supported on mobile platforms")

// SourceHubDocumentACP is a stub for mobile platforms.
// SourceHub ACP requires the Cosmos SDK keyring which is incompatible with mobile platforms.
type SourceHubDocumentACP struct{}

func (a *SourceHubDocumentACP) Start(_ context.Context) error {
	return ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) AddPolicy(
	_ context.Context,
	_ identity.Identity,
	_ string,
	_ acpTypes.PolicyMarshalType,
	_ *protoTypes.Timestamp,
) (string, error) {
	return "", ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) Policy(
	_ context.Context,
	_ string,
) (immutable.Option[acpTypes.Policy], error) {
	return immutable.None[acpTypes.Policy](), ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) RegisterObject(
	_ context.Context,
	_ identity.Identity,
	_ string,
	_ string,
	_ string,
	_ *protoTypes.Timestamp,
) error {
	return ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) ObjectOwner(
	_ context.Context,
	_ string,
	_ string,
	_ string,
) (immutable.Option[string], error) {
	return immutable.None[string](), ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) VerifyAccessRequest(
	_ context.Context,
	_ acpTypes.ResourceInterfacePermission,
	_ string,
	_ string,
	_ string,
	_ string,
) (bool, error) {
	return false, ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) AddActorRelationship(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
	_ identity.Identity,
	_ string,
	_ *protoTypes.Timestamp,
) (bool, error) {
	return false, ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) DeleteActorRelationship(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
	_ identity.Identity,
	_ string,
	_ *protoTypes.Timestamp,
) (bool, error) {
	return false, ErrSourceHubNotSupportedOnMobile
}

func (a *SourceHubDocumentACP) Close() error {
	return nil
}

func (a *SourceHubDocumentACP) ResetState(_ context.Context) error {
	return ErrSourceHubNotSupportedOnMobile
}

// NewSourceHubACP returns an error on mobile platforms as SourceHub ACP is not supported.
func NewSourceHubACP(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer any,
) (DocumentACP, error) {
	return nil, ErrSourceHubNotSupportedOnMobile
}
