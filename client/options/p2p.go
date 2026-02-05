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
	Opts []func(*ConnectOptions)
}

// Connect creates a new ConnectOptionsBuilder instance.
func Connect() *ConnectOptionsBuilder {
	return &ConnectOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ConnectOptionsBuilder) SetIdentity(id identity.Identity) *ConnectOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ConnectOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *ConnectOptionsBuilder) List() []func(*ConnectOptions) {
	return b.Opts
}

// SetReplicatorOptions contains options for SetReplicator operation.
type SetReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *SetReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// SetReplicatorOptionsBuilder is a builder for SetReplicatorOptions.
type SetReplicatorOptionsBuilder struct {
	Opts []func(*SetReplicatorOptions)
}

// SetReplicator creates a new SetReplicatorOptionsBuilder instance.
func SetReplicator() *SetReplicatorOptionsBuilder {
	return &SetReplicatorOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *SetReplicatorOptionsBuilder) SetIdentity(id identity.Identity) *SetReplicatorOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *SetReplicatorOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *SetReplicatorOptionsBuilder) List() []func(*SetReplicatorOptions) {
	return b.Opts
}

// DeleteReplicatorOptions contains options for DeleteReplicator operation.
type DeleteReplicatorOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteReplicatorOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteReplicatorOptionsBuilder is a builder for DeleteReplicatorOptions.
type DeleteReplicatorOptionsBuilder struct {
	Opts []func(*DeleteReplicatorOptions)
}

// DeleteReplicator creates a new DeleteReplicatorOptionsBuilder instance.
func DeleteReplicator() *DeleteReplicatorOptionsBuilder {
	return &DeleteReplicatorOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteReplicatorOptionsBuilder) SetIdentity(id identity.Identity) *DeleteReplicatorOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DeleteReplicatorOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DeleteReplicatorOptionsBuilder) List() []func(*DeleteReplicatorOptions) {
	return b.Opts
}

// GetAllReplicatorsOptions contains options for GetAllReplicators operation.
type GetAllReplicatorsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *GetAllReplicatorsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetAllReplicatorsOptionsBuilder is a builder for GetAllReplicatorsOptions.
type GetAllReplicatorsOptionsBuilder struct {
	Opts []func(*GetAllReplicatorsOptions)
}

// GetAllReplicators creates a new GetAllReplicatorsOptionsBuilder instance.
func GetAllReplicators() *GetAllReplicatorsOptionsBuilder {
	return &GetAllReplicatorsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetAllReplicatorsOptionsBuilder) SetIdentity(id identity.Identity) *GetAllReplicatorsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *GetAllReplicatorsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *GetAllReplicatorsOptionsBuilder) List() []func(*GetAllReplicatorsOptions) {
	return b.Opts
}

// CreateP2PCollectionsOptions contains options for AddP2PCollections operation.
type CreateP2PCollectionsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CreateP2PCollectionsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CreateP2PCollectionsOptionsBuilder is a builder for CreateP2PCollectionsOptions.
type CreateP2PCollectionsOptionsBuilder struct {
	Opts []func(*CreateP2PCollectionsOptions)
}

// CreateP2PCollections creates a new AddP2PCollectionsOptionsBuilder instance.
func CreateP2PCollections() *CreateP2PCollectionsOptionsBuilder {
	return &CreateP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CreateP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *CreateP2PCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CreateP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CreateP2PCollectionsOptionsBuilder) List() []func(*CreateP2PCollectionsOptions) {
	return b.Opts
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
	Opts []func(*DeleteP2PCollectionsOptions)
}

// DeleteP2PCollections creates a new RemoveP2PCollectionsOptionsBuilder instance.
func DeleteP2PCollections() *DeleteP2PCollectionsOptionsBuilder {
	return &DeleteP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *DeleteP2PCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DeleteP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DeleteP2PCollectionsOptionsBuilder) List() []func(*DeleteP2PCollectionsOptions) {
	return b.Opts
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
	Opts []func(*ListP2PCollectionsOptions)
}

// ListP2PCollections creates a new GetAllP2PCollectionsOptionsBuilder instance.
func ListP2PCollections() *ListP2PCollectionsOptionsBuilder {
	return &ListP2PCollectionsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListP2PCollectionsOptionsBuilder) SetIdentity(id identity.Identity) *ListP2PCollectionsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ListP2PCollectionsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *ListP2PCollectionsOptionsBuilder) List() []func(*ListP2PCollectionsOptions) {
	return b.Opts
}

// CreateP2PDocumentsOptions contains options for AddP2PDocuments operation.
type CreateP2PDocumentsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CreateP2PDocumentsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CreateP2PDocumentsOptionsBuilder is a builder for CreateP2PDocumentsOptions.
type CreateP2PDocumentsOptionsBuilder struct {
	Opts []func(*CreateP2PDocumentsOptions)
}

// CreateP2PDocuments creates a new AddP2PDocumentsOptionsBuilder instance.
func CreateP2PDocuments() *CreateP2PDocumentsOptionsBuilder {
	return &CreateP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CreateP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *CreateP2PDocumentsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CreateP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CreateP2PDocumentsOptionsBuilder) List() []func(*CreateP2PDocumentsOptions) {
	return b.Opts
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
	Opts []func(*DeleteP2PDocumentsOptions)
}

// DeleteP2PDocuments creates a new RemoveP2PDocumentsOptionsBuilder instance.
func DeleteP2PDocuments() *DeleteP2PDocumentsOptionsBuilder {
	return &DeleteP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *DeleteP2PDocumentsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *DeleteP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *DeleteP2PDocumentsOptionsBuilder) List() []func(*DeleteP2PDocumentsOptions) {
	return b.Opts
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
	Opts []func(*ListP2PDocumentsOptions)
}

// ListP2PDocuments creates a new GetAllP2PDocumentsOptionsBuilder instance.
func ListP2PDocuments() *ListP2PDocumentsOptionsBuilder {
	return &ListP2PDocumentsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListP2PDocumentsOptionsBuilder) SetIdentity(id identity.Identity) *ListP2PDocumentsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *ListP2PDocumentsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *ListP2PDocumentsOptionsBuilder) List() []func(*ListP2PDocumentsOptions) {
	return b.Opts
}
