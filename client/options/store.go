// Copyright 2026 Democratized Data Foundation
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
	"maps"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// AddDACPolicyOptions contains options for AddDACPolicy operation.
type AddDACPolicyOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddDACPolicyOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddDACPolicyOptionsBuilder is a builder for AddDACPolicyOptions.
type AddDACPolicyOptionsBuilder struct {
	Opts []func(*AddDACPolicyOptions)
}

// AddDACPolicy creates a new AddDACPolicyOptionsBuilder instance.
func AddDACPolicy() *AddDACPolicyOptionsBuilder {
	return &AddDACPolicyOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddDACPolicyOptionsBuilder) SetIdentity(id identity.Identity) *AddDACPolicyOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *AddDACPolicyOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *AddDACPolicyOptionsBuilder) List() []func(*AddDACPolicyOptions) {
	return b.Opts
}

// AddDACActorRelationshipOptions contains options for AddDACActorRelationship operation.
type AddDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddDACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddDACActorRelationshipOptionsBuilder is a builder for AddDACActorRelationshipOptions.
type AddDACActorRelationshipOptionsBuilder struct {
	Opts []func(*AddDACActorRelationshipOptions)
}

// AddDACActorRelationship creates a new AddDACActorRelationshipOptionsBuilder instance.
func AddDACActorRelationship() *AddDACActorRelationshipOptionsBuilder {
	return &AddDACActorRelationshipOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddDACActorRelationshipOptionsBuilder) SetIdentity(
	id identity.Identity,
) *AddDACActorRelationshipOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *AddDACActorRelationshipOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *AddDACActorRelationshipOptionsBuilder) List() []func(*AddDACActorRelationshipOptions) {
	return b.Opts
}

// DeleteDACActorRelationshipOptions contains options for DeleteDACActorRelationship operation.
type DeleteDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteDACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteDACActorRelationshipOptionsBuilder is a builder for DeleteDACActorRelationshipOptions.
type DeleteDACActorRelationshipOptionsBuilder struct {
	Opts []func(*DeleteDACActorRelationshipOptions)
}

// DeleteDACActorRelationship creates a new DeleteDACActorRelationshipOptionsBuilder instance.
func DeleteDACActorRelationship() *DeleteDACActorRelationshipOptionsBuilder {
	return &DeleteDACActorRelationshipOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteDACActorRelationshipOptionsBuilder) SetIdentity(
	id identity.Identity,
) *DeleteDACActorRelationshipOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DeleteDACActorRelationshipOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DeleteDACActorRelationshipOptionsBuilder) List() []func(*DeleteDACActorRelationshipOptions) {
	return b.Opts
}

// AddNACActorRelationshipOptions contains options for AddNACActorRelationship operation.
type AddNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddNACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddNACActorRelationshipOptionsBuilder is a builder for AddNACActorRelationshipOptions.
type AddNACActorRelationshipOptionsBuilder struct {
	Opts []func(*AddNACActorRelationshipOptions)
}

// AddNACActorRelationship creates a new AddNACActorRelationshipOptionsBuilder instance.
func AddNACActorRelationship() *AddNACActorRelationshipOptionsBuilder {
	return &AddNACActorRelationshipOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddNACActorRelationshipOptionsBuilder) SetIdentity(
	id identity.Identity,
) *AddNACActorRelationshipOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *AddNACActorRelationshipOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *AddNACActorRelationshipOptionsBuilder) List() []func(*AddNACActorRelationshipOptions) {
	return b.Opts
}

// DeleteNACActorRelationshipOptions contains options for DeleteNACActorRelationship operation.
type DeleteNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteNACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteNACActorRelationshipOptionsBuilder is a builder for DeleteNACActorRelationshipOptions.
type DeleteNACActorRelationshipOptionsBuilder struct {
	Opts []func(*DeleteNACActorRelationshipOptions)
}

// DeleteNACActorRelationship creates a new DeleteNACActorRelationshipOptionsBuilder instance.
func DeleteNACActorRelationship() *DeleteNACActorRelationshipOptionsBuilder {
	return &DeleteNACActorRelationshipOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteNACActorRelationshipOptionsBuilder) SetIdentity(
	id identity.Identity,
) *DeleteNACActorRelationshipOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DeleteNACActorRelationshipOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DeleteNACActorRelationshipOptionsBuilder) List() []func(*DeleteNACActorRelationshipOptions) {
	return b.Opts
}

// ReEnableNACOptions contains options for ReEnableNAC operation
type ReEnableNACOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ReEnableNACOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ReEnableNACOptionsBuilder is a builder for ReEnableNACOptions.
type ReEnableNACOptionsBuilder struct {
	Opts []func(*ReEnableNACOptions)
}

// ReEnableNAC creates a new NACOptions instance.
func ReEnableNAC() *ReEnableNACOptionsBuilder {
	return &ReEnableNACOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ReEnableNACOptionsBuilder) SetIdentity(id identity.Identity) *ReEnableNACOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ReEnableNACOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *ReEnableNACOptionsBuilder) List() []func(*ReEnableNACOptions) {
	return b.Opts
}

// DisableNACOptions contains options for DisableNAC operation.
type DisableNACOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DisableNACOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DisableNACOptionsBuilder is a builder for DisableNACOptions.
type DisableNACOptionsBuilder struct {
	Opts []func(*DisableNACOptions)
}

// DisableNAC creates a new DisableNACOptionsBuilder instance.
func DisableNAC() *DisableNACOptionsBuilder {
	return &DisableNACOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DisableNACOptionsBuilder) SetIdentity(id identity.Identity) *DisableNACOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DisableNACOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DisableNACOptionsBuilder) List() []func(*DisableNACOptions) {
	return b.Opts
}

// GetNACStatusOptions contains options for GetNACStatus operation.
type GetNACStatusOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *GetNACStatusOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetNACStatusOptionsBuilder is a builder for GetNACStatusOptions.
type GetNACStatusOptionsBuilder struct {
	Opts []func(*GetNACStatusOptions)
}

// GetNACStatus creates a new GetNACStatusOptionsBuilder instance.
func GetNACStatus() *GetNACStatusOptionsBuilder {
	return &GetNACStatusOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetNACStatusOptionsBuilder) SetIdentity(id identity.Identity) *GetNACStatusOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetNACStatusOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *GetNACStatusOptionsBuilder) List() []func(*GetNACStatusOptions) {
	return b.Opts
}

// VerifySignatureOptions contains options for VerifySignature operation.
type VerifySignatureOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *VerifySignatureOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// VerifySignatureOptionsBuilder is a builder for VerifySignatureOptions.
type VerifySignatureOptionsBuilder struct {
	Opts []func(*VerifySignatureOptions)
}

// VerifySignature creates a new VerifySignatureOptionsBuilder instance.
func VerifySignature() *VerifySignatureOptionsBuilder {
	return &VerifySignatureOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *VerifySignatureOptionsBuilder) SetIdentity(id identity.Identity) *VerifySignatureOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *VerifySignatureOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *VerifySignatureOptionsBuilder) List() []func(*VerifySignatureOptions) {
	return b.Opts
}

// AddViewOptions contains options for AddView operation.
type AddViewOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddViewOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddViewOptionsBuilder is a builder for AddViewOptions.
type AddViewOptionsBuilder struct {
	Opts []func(*AddViewOptions)
}

// AddView creates a new AddViewOptionsBuilder instance.
func AddView() *AddViewOptionsBuilder {
	return &AddViewOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddViewOptionsBuilder) SetIdentity(id identity.Identity) *AddViewOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *AddViewOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *AddViewOptionsBuilder) List() []func(*AddViewOptions) {
	return b.Opts
}

// RefreshViewsOptions contains options for RefreshViews operation.
type RefreshViewsOptions = GetCollectionsOptions

// RefreshViewsOptionsBuilder is a builder for RefreshViewsOptions.
type RefreshViewsOptionsBuilder = GetCollectionsOptionsBuilder

// RefreshViews creates a new RefreshViewsOptionsBuilder instance.
func RefreshViews() *GetCollectionsOptionsBuilder {
	return &GetCollectionsOptionsBuilder{}
}

// GetCollectionByNameOptions contains options for GetCollectionByName operation.
type GetCollectionByNameOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *GetCollectionByNameOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetCollectionByNameOptionsBuilder is a builder for GetCollectionByNameOptions.
type GetCollectionByNameOptionsBuilder struct {
	Opts []func(*GetCollectionByNameOptions)
}

// GetCollectionByName creates a new GetCollectionByNameOptionsBuilder instance.
func GetCollectionByName() *GetCollectionByNameOptionsBuilder {
	return &GetCollectionByNameOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetCollectionByNameOptionsBuilder) SetIdentity(id identity.Identity) *GetCollectionByNameOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionByNameOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *GetCollectionByNameOptionsBuilder) List() []func(*GetCollectionByNameOptions) {
	return b.Opts
}

// GetCollectionsOptions contains options for GetCollections operation.
type GetCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// If provided, only collections with this version id will be returned.
	VersionID immutable.Option[string]
	// If provided, only collections with this CollectionID will be returned.
	CollectionID immutable.Option[string]
	// If provided, only collections with this CollectionSetID will be returned.
	CollectionSetID immutable.Option[string]
	// If provided, only collections with this name will be returned.
	CollectionName immutable.Option[string]
	// If IncludeInactive is true, then inactive collections will also be returned.
	IncludeInactive immutable.Option[bool]
}

// GetIdentity returns the identity for the operation.
func (o *GetCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetCollectionsOptionsBuilder is a builder for GetCollectionsOptions.
type GetCollectionsOptionsBuilder struct {
	Opts []func(*GetCollectionsOptions)
}

// GetCollections creates a new GetCollectionsOptionsBuilder instance.
func GetCollections() *GetCollectionsOptionsBuilder {
	return &GetCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetVersionID sets the version ID filter.
func (b *GetCollectionsOptionsBuilder) SetVersionID(versionID string) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.VersionID = immutable.Some(versionID)
	})
	return b
}

// SetCollectionID sets the collection ID filter.
func (b *GetCollectionsOptionsBuilder) SetCollectionID(collectionID string) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.CollectionID = immutable.Some(collectionID)
	})
	return b
}

// SetCollectionSetID sets the collection set ID filter.
func (b *GetCollectionsOptionsBuilder) SetCollectionSetID(collectionSetID string) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.CollectionSetID = immutable.Some(collectionSetID)
	})
	return b
}

// SetCollectionName sets the name filter.
func (b *GetCollectionsOptionsBuilder) SetCollectionName(name string) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.CollectionName = immutable.Some(name)
	})
	return b
}

// SetIncludeInactive sets whether to include inactive collections.
func (b *GetCollectionsOptionsBuilder) SetIncludeInactive(includeInactive bool) *GetCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetCollectionsOptions) {
		opts.IncludeInactive = immutable.Some(includeInactive)
	})
	return b
}

// List returns the list of functional options.
func (b *GetCollectionsOptionsBuilder) List() []func(*GetCollectionsOptions) {
	return b.Opts
}

// GetAllIndexesOptions contains options for GetAllIndexes operation.
type GetAllIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *GetAllIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetAllIndexesOptionsBuilder is a builder for GetAllIndexesOptions.
type GetAllIndexesOptionsBuilder struct {
	Opts []func(*GetAllIndexesOptions)
}

// GetAllIndexes creates a new GetAllIndexesOptionsBuilder instance.
func GetAllIndexes() *GetAllIndexesOptionsBuilder {
	return &GetAllIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetAllIndexesOptionsBuilder) SetIdentity(id identity.Identity) *GetAllIndexesOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetAllIndexesOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *GetAllIndexesOptionsBuilder) List() []func(*GetAllIndexesOptions) {
	return b.Opts
}

// AddSchemaOptions contains options for AddSchema operation.
type AddSchemaOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddSchemaOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddSchemaOptionsBuilder is a builder for AddSchemaOptions.
type AddSchemaOptionsBuilder struct {
	Opts []func(*AddSchemaOptions)
}

// AddSchema creates a new AddSchemaOptionsBuilder instance.
func AddSchema() *AddSchemaOptionsBuilder {
	return &AddSchemaOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddSchemaOptionsBuilder) SetIdentity(id identity.Identity) *AddSchemaOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *AddSchemaOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *AddSchemaOptionsBuilder) List() []func(*AddSchemaOptions) {
	return b.Opts
}

// PatchCollectionOptions contains options for PatchCollection operation.
type PatchCollectionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *PatchCollectionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// PatchCollectionOptionsBuilder is a builder for PatchCollectionOptions.
type PatchCollectionOptionsBuilder struct {
	Opts []func(*PatchCollectionOptions)
}

// PatchCollection creates a new PatchCollectionOptionsBuilder instance.
func PatchCollection() *PatchCollectionOptionsBuilder {
	return &PatchCollectionOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *PatchCollectionOptionsBuilder) SetIdentity(id identity.Identity) *PatchCollectionOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *PatchCollectionOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *PatchCollectionOptionsBuilder) List() []func(*PatchCollectionOptions) {
	return b.Opts
}

// SetActiveCollectionVersionOptions contains options for SetActiveCollectionVersion operation.
type SetActiveCollectionVersionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *SetActiveCollectionVersionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetActiveCollectionVersionOptionsBuilder is a builder for SetActiveCollectionVersionOptions.
type SetActiveCollectionVersionOptionsBuilder struct {
	Opts []func(*SetActiveCollectionVersionOptions)
}

// SetActiveCollectionVersion creates a new SetActiveCollectionVersionOptionsBuilder instance.
func SetActiveCollectionVersion() *SetActiveCollectionVersionOptionsBuilder {
	return &SetActiveCollectionVersionOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *SetActiveCollectionVersionOptionsBuilder) SetIdentity(
	id identity.Identity,
) *SetActiveCollectionVersionOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *SetActiveCollectionVersionOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *SetActiveCollectionVersionOptionsBuilder) List() []func(*SetActiveCollectionVersionOptions) {
	return b.Opts
}

// ExecRequestOptions contains options for ExecRequest operation.
type ExecRequestOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// OperationName is the name of the operation to exec.
	OperationName immutable.Option[string]
	// Variables is a map of names to variable values.
	Variables map[string]any
}

// GetIdentity returns the identity for the operation.
func (o *ExecRequestOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ExecRequestOptionsBuilder is a builder for ExecRequestOptions.
type ExecRequestOptionsBuilder struct {
	Opts []func(*ExecRequestOptions)
}

// ExecRequest creates a new ExecRequestOptionsBuilder instance.
func ExecRequest() *ExecRequestOptionsBuilder {
	return &ExecRequestOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ExecRequestOptionsBuilder) SetIdentity(id identity.Identity) *ExecRequestOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ExecRequestOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetOperationName sets the operation name for a GQL request.
func (b *ExecRequestOptionsBuilder) SetOperationName(operationName string) *ExecRequestOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ExecRequestOptions) {
		opts.OperationName = immutable.Some(operationName)
	})
	return b
}

// SetVariables sets the variables for a GQL request.
func (b *ExecRequestOptionsBuilder) SetVariables(variables map[string]any) *ExecRequestOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ExecRequestOptions) {
		if variables != nil {
			opts.Variables = make(map[string]any, len(variables))
			maps.Copy(opts.Variables, variables)
		}
	})
	return b
}

// List returns the list of functional options.
func (b *ExecRequestOptionsBuilder) List() []func(*ExecRequestOptions) {
	return b.Opts
}
