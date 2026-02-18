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

// ActivePeersOptions contains options for ActivePeers operation.
type ActivePeersOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ActivePeersOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ActivePeersOptionsBuilder is a builder for ActivePeersOptions.
type ActivePeersOptionsBuilder struct {
	enumerableBuilder[ActivePeersOptions]
}

// ActivePeers creates a new ActivePeersOptionsBuilder instance.
func ActivePeers() *ActivePeersOptionsBuilder {
	return &ActivePeersOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ActivePeersOptionsBuilder) SetIdentity(id identity.Identity) *ActivePeersOptionsBuilder {
	b.append(func(opts *ActivePeersOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ConnectOptions contains options for Connect operation.
type ConnectOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ConnectOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ConnectOptionsBuilder is a builder for ConnectOptions.
type ConnectOptionsBuilder struct {
	enumerableBuilder[ConnectOptions]
}

// Connect creates a new ConnectOptionsBuilder instance.
func Connect() *ConnectOptionsBuilder {
	return &ConnectOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ConnectOptionsBuilder) SetIdentity(id identity.Identity) *ConnectOptionsBuilder {
	b.append(func(opts *ConnectOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

type PeerInfoOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *PeerInfoOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// PeerInfoOptionsBuilder is a builder for PeerInfoOptions.
type PeerInfoOptionsBuilder struct {
	enumerableBuilder[PeerInfoOptions]
}

// AddReplicator creates a new AddReplicatorOptionsBuilder instance.
func PeerInfo() *PeerInfoOptionsBuilder {
	return &PeerInfoOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *PeerInfoOptionsBuilder) SetIdentity(id identity.Identity) *PeerInfoOptionsBuilder {
	b.append(func(opts *PeerInfoOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// AddReplicatorOptions contains options for AddReplicator operation.
type AddReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// CollectionNames is the list of collection names to replicate.
	CollectionNames []string
}

// GetIdentity returns the identity for the operation.
func (o *AddReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddReplicatorOptionsBuilder is a builder for AddReplicatorOptions.
type AddReplicatorOptionsBuilder struct {
	enumerableBuilder[AddReplicatorOptions]
}

// AddReplicator creates a new AddReplicatorOptionsBuilder instance.
func AddReplicator() *AddReplicatorOptionsBuilder {
	return &AddReplicatorOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddReplicatorOptionsBuilder) SetIdentity(id identity.Identity) *AddReplicatorOptionsBuilder {
	b.append(func(opts *AddReplicatorOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetCollectionNames sets the collection names to replicate.
func (b *AddReplicatorOptionsBuilder) SetCollectionNames(names []string) *AddReplicatorOptionsBuilder {
	b.append(func(opts *AddReplicatorOptions) {
		if names != nil {
			opts.CollectionNames = make([]string, len(names))
			copy(opts.CollectionNames, names)
		}
	})
	return b
}

// DeleteReplicatorOptions contains options for DeleteReplicator operation.
type DeleteReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// CollectionNames is the list of collection names to stop replicating.
	CollectionNames []string
}

// GetIdentity returns the identity for the operation.
func (o *DeleteReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteReplicatorOptionsBuilder is a builder for DeleteReplicatorOptions.
type DeleteReplicatorOptionsBuilder struct {
	enumerableBuilder[DeleteReplicatorOptions]
}

// DeleteReplicator creates a new DeleteReplicatorOptionsBuilder instance.
func DeleteReplicator() *DeleteReplicatorOptionsBuilder {
	return &DeleteReplicatorOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteReplicatorOptionsBuilder) SetIdentity(id identity.Identity) *DeleteReplicatorOptionsBuilder {
	b.append(func(opts *DeleteReplicatorOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetCollectionNames sets the collection names to stop replicating.
func (b *DeleteReplicatorOptionsBuilder) SetCollectionNames(names []string) *DeleteReplicatorOptionsBuilder {
	b.append(func(opts *DeleteReplicatorOptions) {
		if names != nil {
			opts.CollectionNames = make([]string, len(names))
			copy(opts.CollectionNames, names)
		}
	})
	return b
}

// ListReplicatorsOptions contains options for GetAllReplicators operation.
type ListReplicatorsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ListReplicatorsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListReplicatorsOptionsBuilder is a builder for GetAllReplicatorsOptions.
type ListReplicatorsOptionsBuilder struct {
	enumerableBuilder[ListReplicatorsOptions]
}

// ListReplicators creates a new GetAllReplicatorsOptionsBuilder instance.
func ListReplicators() *ListReplicatorsOptionsBuilder {
	return &ListReplicatorsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListReplicatorsOptionsBuilder) SetIdentity(id identity.Identity) *ListReplicatorsOptionsBuilder {
	b.append(func(opts *ListReplicatorsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// AddP2PCollectionsOptions contains options for AddP2PCollections operation.
type AddP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddP2PCollectionsOptionsBuilder is a builder for AddP2PCollectionsOptions.
type AddP2PCollectionsOptionsBuilder struct {
	enumerableBuilder[AddP2PCollectionsOptions]
}

// AddP2PCollections creates a new AddP2PCollectionsOptionsBuilder instance.
func AddP2PCollections() *AddP2PCollectionsOptionsBuilder {
	return &AddP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *AddP2PCollectionsOptionsBuilder {
	b.append(func(opts *AddP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// DeleteP2PCollectionsOptions contains options for RemoveP2PCollections operation.
type DeleteP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteP2PCollectionsOptionsBuilder is a builder for DeleteP2PCollectionsOptions.
type DeleteP2PCollectionsOptionsBuilder struct {
	enumerableBuilder[DeleteP2PCollectionsOptions]
}

// DeleteP2PCollections creates a new RemoveP2PCollectionsOptionsBuilder instance.
func DeleteP2PCollections() *DeleteP2PCollectionsOptionsBuilder {
	return &DeleteP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *DeleteP2PCollectionsOptionsBuilder {
	b.append(func(opts *DeleteP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ListP2PCollectionsOptions contains options for GetAllP2PCollections operation.
type ListP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ListP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListP2PCollectionsOptionsBuilder is a builder for ListP2PCollectionsOptions.
type ListP2PCollectionsOptionsBuilder struct {
	enumerableBuilder[ListP2PCollectionsOptions]
}

// ListP2PCollections creates a new GetAllP2PCollectionsOptionsBuilder instance.
func ListP2PCollections() *ListP2PCollectionsOptionsBuilder {
	return &ListP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *ListP2PCollectionsOptionsBuilder {
	b.append(func(opts *ListP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// AddP2PDocumentsOptions contains options for AddP2PDocuments operation.
type AddP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *AddP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddP2PDocumentsOptionsBuilder is a builder for AddP2PDocumentsOptions.
type AddP2PDocumentsOptionsBuilder struct {
	enumerableBuilder[AddP2PDocumentsOptions]
}

// AddP2PDocuments creates a new AddP2PDocumentsOptionsBuilder instance.
func AddP2PDocuments() *AddP2PDocumentsOptionsBuilder {
	return &AddP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *AddP2PDocumentsOptionsBuilder {
	b.append(func(opts *AddP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SyncCollectionVersionsOptions contains options for SyncCollectionVersions operation.
type SyncCollectionVersionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *SyncCollectionVersionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SyncCollectionVersionsOptionsBuilder is a builder for SyncCollectionVersionsOptions.
type SyncCollectionVersionsOptionsBuilder struct {
	enumerableBuilder[SyncCollectionVersionsOptions]
}

// SyncCollectionVersions creates a new SyncCollectionVersionsOptionsBuilder instance.
func SyncCollectionVersions() *SyncCollectionVersionsOptionsBuilder {
	return &SyncCollectionVersionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *SyncCollectionVersionsOptionsBuilder) SetIdentity(id identity.Identity) *SyncCollectionVersionsOptionsBuilder {
	b.append(func(opts *SyncCollectionVersionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SyncBranchableCollectionOptions contains options for SyncBranchableCollection operation.
type SyncBranchableCollectionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *SyncBranchableCollectionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SyncBranchableCollectionOptionsBuilder is a builder for SyncBranchableCollectionOptions.
type SyncBranchableCollectionOptionsBuilder struct {
	enumerableBuilder[SyncBranchableCollectionOptions]
}

// SyncBranchableCollection creates a new SyncBranchableCollectionOptionsBuilder instance.
func SyncBranchableCollection() *SyncBranchableCollectionOptionsBuilder {
	return &SyncBranchableCollectionOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *SyncBranchableCollectionOptionsBuilder) SetIdentity(
	id identity.Identity,
) *SyncBranchableCollectionOptionsBuilder {
	b.append(func(opts *SyncBranchableCollectionOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// DeleteP2PDocumentsOptions contains options for RemoveP2PDocuments operation.
type DeleteP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteP2PDocumentsOptionsBuilder is a builder for DeleteP2PDocumentsOptions.
type DeleteP2PDocumentsOptionsBuilder struct {
	enumerableBuilder[DeleteP2PDocumentsOptions]
}

// DeleteP2PDocuments creates a new RemoveP2PDocumentsOptionsBuilder instance.
func DeleteP2PDocuments() *DeleteP2PDocumentsOptionsBuilder {
	return &DeleteP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *DeleteP2PDocumentsOptionsBuilder {
	b.append(func(opts *DeleteP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ListP2PDocumentsOptions contains options for GetAllP2PDocuments operation.
type ListP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ListP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListP2PDocumentsOptionsBuilder is a builder for ListP2PDocumentsOptions.
type ListP2PDocumentsOptionsBuilder struct {
	enumerableBuilder[ListP2PDocumentsOptions]
}

// ListP2PDocuments creates a new GetAllP2PDocumentsOptionsBuilder instance.
func ListP2PDocuments() *ListP2PDocumentsOptionsBuilder {
	return &ListP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *ListP2PDocumentsOptionsBuilder {
	b.append(func(opts *ListP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}
