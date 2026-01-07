// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package options

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// AddDACPolicyOptions contains options for AddDACPolicy operation.
type AddDACPolicyOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// AddDACPolicy creates a new AddDACPolicyOptions instance.
func AddDACPolicy() *AddDACPolicyOptions {
	return &AddDACPolicyOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *AddDACPolicyOptions) SetIdentity(id identity.Identity) *AddDACPolicyOptions {
	o.Identity = immutable.Some(id)
	return o
}

// AddDACActorRelationshipOptions contains options for AddDACActorRelationship operation.
type AddDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// AddDACActorRelationship creates a new AddDACActorRelationshipOptions instance.
func AddDACActorRelationship() *AddDACActorRelationshipOptions {
	return &AddDACActorRelationshipOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *AddDACActorRelationshipOptions) SetIdentity(id identity.Identity) *AddDACActorRelationshipOptions {
	o.Identity = immutable.Some(id)
	return o
}

// DeleteDACActorRelationshipOptions contains options for DeleteDACActorRelationship operation.
type DeleteDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// DeleteDACActorRelationship creates a new DeleteDACActorRelationshipOptions instance.
func DeleteDACActorRelationship() *DeleteDACActorRelationshipOptions {
	return &DeleteDACActorRelationshipOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DeleteDACActorRelationshipOptions) SetIdentity(id identity.Identity) *DeleteDACActorRelationshipOptions {
	o.Identity = immutable.Some(id)
	return o
}

// AddNACActorRelationshipOptions contains options for AddNACActorRelationship operation.
type AddNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// AddNACActorRelationship creates a new AddNACActorRelationshipOptions instance.
func AddNACActorRelationship() *AddNACActorRelationshipOptions {
	return &AddNACActorRelationshipOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *AddNACActorRelationshipOptions) SetIdentity(id identity.Identity) *AddNACActorRelationshipOptions {
	o.Identity = immutable.Some(id)
	return o
}

// DeleteNACActorRelationshipOptions contains options for DeleteNACActorRelationship operation.
type DeleteNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// DeleteNACActorRelationship creates a new DeleteNACActorRelationshipOptions instance.
func DeleteNACActorRelationship() *DeleteNACActorRelationshipOptions {
	return &DeleteNACActorRelationshipOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DeleteNACActorRelationshipOptions) SetIdentity(id identity.Identity) *DeleteNACActorRelationshipOptions {
	o.Identity = immutable.Some(id)
	return o
}

// NACOptions contains options for NAC operations (ReEnableNAC, DisableNAC, GetNACStatus).
type NACOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// NAC creates a new NACOptions instance.
func NAC() *NACOptions {
	return &NACOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *NACOptions) SetIdentity(id identity.Identity) *NACOptions {
	o.Identity = immutable.Some(id)
	return o
}

// VerifySignatureOptions contains options for VerifySignature operation.
type VerifySignatureOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// VerifySignature creates a new VerifySignatureOptions instance.
func VerifySignature() *VerifySignatureOptions {
	return &VerifySignatureOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *VerifySignatureOptions) SetIdentity(id identity.Identity) *VerifySignatureOptions {
	o.Identity = immutable.Some(id)
	return o
}

// AddViewOptions contains options for AddView operation.
type AddViewOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// AddView creates a new AddViewOptions instance.
func AddView() *AddViewOptions {
	return &AddViewOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *AddViewOptions) SetIdentity(id identity.Identity) *AddViewOptions {
	o.Identity = immutable.Some(id)
	return o
}

// RefreshViewsOptions contains options for RefreshViews operation.
type RefreshViewsOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// RefreshViews creates a new RefreshViewsOptions instance.
func RefreshViews() *RefreshViewsOptions {
	return &RefreshViewsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *RefreshViewsOptions) SetIdentity(id identity.Identity) *RefreshViewsOptions {
	o.Identity = immutable.Some(id)
	return o
}
