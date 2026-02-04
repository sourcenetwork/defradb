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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// AddDACPolicyOptions contains options for AddDACPolicy operation.
type AddDACPolicyOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *AddDACPolicyOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddDACActorRelationshipOptions contains options for AddDACActorRelationship operation.
type AddDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *AddDACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteDACActorRelationshipOptions contains options for DeleteDACActorRelationship operation.
type DeleteDACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *DeleteDACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddNACActorRelationshipOptions contains options for AddNACActorRelationship operation.
type AddNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *AddNACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteNACActorRelationshipOptions contains options for DeleteNACActorRelationship operation.
type DeleteNACActorRelationshipOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *DeleteNACActorRelationshipOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ReEnableNACOptions contains options for ReEnableNAC operation
type ReEnableNACOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// ReEnableNAC creates a new NACOptions instance.
func ReEnableNAC() *ReEnableNACOptions {
	return &ReEnableNACOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *ReEnableNACOptions) SetIdentity(id identity.Identity) *ReEnableNACOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *ReEnableNACOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DisableNACOptions contains options for DisableNAC operation.
type DisableNACOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// DisableNAC creates a new DisableNACOptions instance.
func DisableNAC() *DisableNACOptions {
	return &DisableNACOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DisableNACOptions) SetIdentity(id identity.Identity) *DisableNACOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *DisableNACOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetNACStatusOptions contains options for GetNACStatus operation.
type GetNACStatusOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetNACStatus creates a new GetNACStatusOptions instance.
func GetNACStatus() *GetNACStatusOptions {
	return &GetNACStatusOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *GetNACStatusOptions) SetIdentity(id identity.Identity) *GetNACStatusOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *GetNACStatusOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// VerifySignatureOptions contains options for VerifySignature operation.
type VerifySignatureOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *VerifySignatureOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddViewOptions contains options for AddView operation.
type AddViewOptions struct {
	// Identity is the identity of the actor performing the operation.
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

// GetIdentity returns the identity for the operation.
func (o *AddViewOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// RefreshViewsOptions contains options for RefreshViews operation.
type RefreshViewsOptions = GetCollectionsOptions

// RefreshViews creates a new RefreshViewsOptions instance.
func RefreshViews() *RefreshViewsOptions {
	return &RefreshViewsOptions{}
}

// GetCollectionByNameOptions contains options for GetCollectionByName operation.
type GetCollectionByNameOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetCollectionByName creates a new GetCollectionByNameOptions instance.
func GetCollectionByName() *GetCollectionByNameOptions {
	return &GetCollectionByNameOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *GetCollectionByNameOptions) SetIdentity(id identity.Identity) *GetCollectionByNameOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *GetCollectionByNameOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
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

// GetCollections creates a new GetCollectionsOptions instance.
func GetCollections() *GetCollectionsOptions {
	return &GetCollectionsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *GetCollectionsOptions) SetIdentity(id identity.Identity) *GetCollectionsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *GetCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetVersionID sets the version ID filter.
func (o *GetCollectionsOptions) SetVersionID(versionID string) *GetCollectionsOptions {
	o.VersionID = immutable.Some(versionID)
	return o
}

// SetCollectionID sets the collection ID filter.
func (o *GetCollectionsOptions) SetCollectionID(collectionID string) *GetCollectionsOptions {
	o.CollectionID = immutable.Some(collectionID)
	return o
}

// SetCollectionSetID sets the collection set ID filter.
func (o *GetCollectionsOptions) SetCollectionSetID(collectionSetID string) *GetCollectionsOptions {
	o.CollectionSetID = immutable.Some(collectionSetID)
	return o
}

// SetCollectionName sets the name filter.
func (o *GetCollectionsOptions) SetCollectionName(name string) *GetCollectionsOptions {
	o.CollectionName = immutable.Some(name)
	return o
}

// SetIncludeInactive sets whether to include inactive collections.
func (o *GetCollectionsOptions) SetIncludeInactive(includeInactive bool) *GetCollectionsOptions {
	o.IncludeInactive = immutable.Some(includeInactive)
	return o
}

// GetAllIndexesOptions contains options for GetAllIndexes operation.
type GetAllIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetAllIndexes creates a new GetAllIndexesOptions instance.
func GetAllIndexes() *GetAllIndexesOptions {
	return &GetAllIndexesOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *GetAllIndexesOptions) SetIdentity(id identity.Identity) *GetAllIndexesOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *GetAllIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddSchemaOptions contains options for AddSchema operation.
type AddSchemaOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// AddSchema creates a new AddSchemaOptions instance.
func AddSchema() *AddSchemaOptions {
	return &AddSchemaOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *AddSchemaOptions) SetIdentity(id identity.Identity) *AddSchemaOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *AddSchemaOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// PatchCollectionOptions contains options for PatchCollection operation.
type PatchCollectionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// PatchCollection creates a new PatchCollectionOptions instance.
func PatchCollection() *PatchCollectionOptions {
	return &PatchCollectionOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *PatchCollectionOptions) SetIdentity(id identity.Identity) *PatchCollectionOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *PatchCollectionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetActiveCollectionVersionOptions contains options for SetActiveCollectionVersion operation.
type SetActiveCollectionVersionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// SetActiveCollectionVersion creates a new SetActiveCollectionVersionOptions instance.
func SetActiveCollectionVersion() *SetActiveCollectionVersionOptions {
	return &SetActiveCollectionVersionOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *SetActiveCollectionVersionOptions) SetIdentity(id identity.Identity) *SetActiveCollectionVersionOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *SetActiveCollectionVersionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
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

// ExecRequest creates a new ExecRequestOptions instance.
func ExecRequest() *ExecRequestOptions {
	return &ExecRequestOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *ExecRequestOptions) SetIdentity(id identity.Identity) *ExecRequestOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *ExecRequestOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetOperationName sets the operation name for a GQL request.
func (o *ExecRequestOptions) SetOperationName(operationName string) *ExecRequestOptions {
	o.OperationName = immutable.Some(operationName)
	return o
}

// SetVariables sets the variables for a GQL request.
func (o *ExecRequestOptions) SetVariables(variables map[string]any) *ExecRequestOptions {
	o.Variables = variables
	return o
}
