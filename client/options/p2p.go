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

// ConnectOptions contains options for Connect operation.
type ConnectOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// Connect creates a new ConnectOptions instance.
func Connect() *ConnectOptions {
	return &ConnectOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *ConnectOptions) SetIdentity(id identity.Identity) *ConnectOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *ConnectOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetReplicatorOptions contains options for SetReplicator operation.
type SetReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// SetReplicator creates a new SetReplicatorOptions instance.
func SetReplicator() *SetReplicatorOptions {
	return &SetReplicatorOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *SetReplicatorOptions) SetIdentity(id identity.Identity) *SetReplicatorOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *SetReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteReplicatorOptions contains options for DeleteReplicator operation.
type DeleteReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// DeleteReplicator creates a new DeleteReplicatorOptions instance.
func DeleteReplicator() *DeleteReplicatorOptions {
	return &DeleteReplicatorOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DeleteReplicatorOptions) SetIdentity(id identity.Identity) *DeleteReplicatorOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *DeleteReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetAllReplicatorsOptions contains options for GetAllReplicators operation.
type GetAllReplicatorsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetAllReplicators creates a new GetAllReplicatorsOptions instance.
func GetAllReplicators() *GetAllReplicatorsOptions {
	return &GetAllReplicatorsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *GetAllReplicatorsOptions) SetIdentity(id identity.Identity) *GetAllReplicatorsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *GetAllReplicatorsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CreateP2PCollectionsOptions contains options for AddP2PCollections operation.
type CreateP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// CreateP2PCollections creates a new AddP2PCollectionsOptions instance.
func CreateP2PCollections() *CreateP2PCollectionsOptions {
	return &CreateP2PCollectionsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CreateP2PCollectionsOptions) SetIdentity(id identity.Identity) *CreateP2PCollectionsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *CreateP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteP2PCollectionsOptions contains options for RemoveP2PCollections operation.
type DeleteP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// DeleteP2PCollections creates a new RemoveP2PCollectionsOptions instance.
func DeleteP2PCollections() *DeleteP2PCollectionsOptions {
	return &DeleteP2PCollectionsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DeleteP2PCollectionsOptions) SetIdentity(id identity.Identity) *DeleteP2PCollectionsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *DeleteP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListP2PCollectionsOptions contains options for GetAllP2PCollections operation.
type ListP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// ListP2PCollections creates a new GetAllP2PCollectionsOptions instance.
func ListP2PCollections() *ListP2PCollectionsOptions {
	return &ListP2PCollectionsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *ListP2PCollectionsOptions) SetIdentity(id identity.Identity) *ListP2PCollectionsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *ListP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CreateP2PDocumentsOptions contains options for AddP2PDocuments operation.
type CreateP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// CreateP2PDocuments creates a new AddP2PDocumentsOptions instance.
func CreateP2PDocuments() *CreateP2PDocumentsOptions {
	return &CreateP2PDocumentsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CreateP2PDocumentsOptions) SetIdentity(id identity.Identity) *CreateP2PDocumentsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *CreateP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteP2PDocumentsOptions contains options for RemoveP2PDocuments operation.
type DeleteP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// DeleteP2PDocuments creates a new RemoveP2PDocumentsOptions instance.
func DeleteP2PDocuments() *DeleteP2PDocumentsOptions {
	return &DeleteP2PDocumentsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *DeleteP2PDocumentsOptions) SetIdentity(id identity.Identity) *DeleteP2PDocumentsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *DeleteP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListP2PDocumentsOptions contains options for GetAllP2PDocuments operation.
type ListP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// ListP2PDocuments creates a new GetAllP2PDocumentsOptions instance.
func ListP2PDocuments() *ListP2PDocumentsOptions {
	return &ListP2PDocumentsOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *ListP2PDocumentsOptions) SetIdentity(id identity.Identity) *ListP2PDocumentsOptions {
	o.Identity = immutable.Some(id)
	return o
}

// GetIdentity returns the identity for the operation.
func (o *ListP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}
