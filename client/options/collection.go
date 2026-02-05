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

// CollectionCreateOptions contains options for Create and CreateMany operations.
type CollectionCreateOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// EncryptDoc enables document encryption when creating a document.
	EncryptDoc bool
	// EncryptedFields specifies a list of fields to be encrypted.
	EncryptedFields []string
}

// GetIdentity returns the identity for the operation.
func (o *CollectionCreateOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionCreateOptionsBuilder is a builder for CollectionCreateOptions.
type CollectionCreateOptionsBuilder struct {
	Opts []func(*CollectionCreateOptions)
}

// CollectionCreate creates a new CollectionCreateOptionsBuilder instance.
func CollectionCreate() *CollectionCreateOptionsBuilder {
	return &CollectionCreateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionCreateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionCreateOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionCreateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetEncryptDoc enables or disables document encryption.
func (b *CollectionCreateOptionsBuilder) SetEncryptDoc(encrypt bool) *CollectionCreateOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionCreateOptions) {
		opts.EncryptDoc = encrypt
	})
	return b
}

// SetEncryptedFields specifies fields to be encrypted.
func (b *CollectionCreateOptionsBuilder) SetEncryptedFields(fields []string) *CollectionCreateOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionCreateOptions) {
		if fields != nil {
			opts.EncryptedFields = make([]string, len(fields))
			copy(opts.EncryptedFields, fields)
		}
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionCreateOptionsBuilder) List() []func(*CollectionCreateOptions) {
	return b.Opts
}

// CollectionUpdateOptions contains options for Update operation.
type CollectionUpdateOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionUpdateOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionUpdateOptionsBuilder is a builder for CollectionUpdateOptions.
type CollectionUpdateOptionsBuilder struct {
	Opts []func(*CollectionUpdateOptions)
}

// CollectionUpdate creates a new CollectionUpdateOptionsBuilder instance.
func CollectionUpdate() *CollectionUpdateOptionsBuilder {
	return &CollectionUpdateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionUpdateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionUpdateOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionUpdateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionUpdateOptionsBuilder) List() []func(*CollectionUpdateOptions) {
	return b.Opts
}

// CollectionSaveOptions contains options for Save operation.
type CollectionSaveOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// EncryptDoc enables document encryption when creating a document.
	EncryptDoc bool
	// EncryptedFields specifies a list of fields to be encrypted.
	EncryptedFields []string
}

// GetIdentity returns the identity for the operation.
func (o *CollectionSaveOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionSaveOptionsBuilder is a builder for CollectionSaveOptions.
type CollectionSaveOptionsBuilder struct {
	Opts []func(*CollectionSaveOptions)
}

// CollectionSave creates a new CollectionSaveOptionsBuilder instance.
func CollectionSave() *CollectionSaveOptionsBuilder {
	return &CollectionSaveOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionSaveOptionsBuilder) SetIdentity(id identity.Identity) *CollectionSaveOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionSaveOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetEncryptDoc enables or disables document encryption.
func (b *CollectionSaveOptionsBuilder) SetEncryptDoc(encrypt bool) *CollectionSaveOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionSaveOptions) {
		opts.EncryptDoc = encrypt
	})
	return b
}

// SetEncryptedFields specifies fields to be encrypted.
func (b *CollectionSaveOptionsBuilder) SetEncryptedFields(fields []string) *CollectionSaveOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionSaveOptions) {
		if fields != nil {
			opts.EncryptedFields = make([]string, len(fields))
			copy(opts.EncryptedFields, fields)
		}
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionSaveOptionsBuilder) List() []func(*CollectionSaveOptions) {
	return b.Opts
}

// CollectionDeleteOptions contains options for Delete operation.
type CollectionDeleteOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionDeleteOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionDeleteOptionsBuilder is a builder for CollectionDeleteOptions.
type CollectionDeleteOptionsBuilder struct {
	Opts []func(*CollectionDeleteOptions)
}

// CollectionDelete creates a new CollectionDeleteOptionsBuilder instance.
func CollectionDelete() *CollectionDeleteOptionsBuilder {
	return &CollectionDeleteOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDeleteOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDeleteOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionDeleteOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionDeleteOptionsBuilder) List() []func(*CollectionDeleteOptions) {
	return b.Opts
}

// CollectionGetOptions contains options for Get operation.
type CollectionGetOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionGetOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionGetOptionsBuilder is a builder for CollectionGetOptions.
type CollectionGetOptionsBuilder struct {
	Opts []func(*CollectionGetOptions)
}

// CollectionGet creates a new CollectionGetOptionsBuilder instance.
func CollectionGet() *CollectionGetOptionsBuilder {
	return &CollectionGetOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionGetOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionGetOptionsBuilder) List() []func(*CollectionGetOptions) {
	return b.Opts
}

// CollectionUpdateWithFilterOptions contains options for UpdateWithFilter operation.
type CollectionUpdateWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionUpdateWithFilterOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionUpdateWithFilterOptionsBuilder is a builder for CollectionUpdateWithFilterOptions.
type CollectionUpdateWithFilterOptionsBuilder struct {
	Opts []func(*CollectionUpdateWithFilterOptions)
}

// CollectionUpdateWithFilter creates a new CollectionUpdateWithFilterOptionsBuilder instance.
func CollectionUpdateWithFilter() *CollectionUpdateWithFilterOptionsBuilder {
	return &CollectionUpdateWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionUpdateWithFilterOptionsBuilder) SetIdentity(id identity.Identity) *CollectionUpdateWithFilterOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionUpdateWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionUpdateWithFilterOptionsBuilder) List() []func(*CollectionUpdateWithFilterOptions) {
	return b.Opts
}

// CollectionDeleteWithFilterOptions contains options for DeleteWithFilter operation.
type CollectionDeleteWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionDeleteWithFilterOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionDeleteWithFilterOptionsBuilder is a builder for CollectionDeleteWithFilterOptions.
type CollectionDeleteWithFilterOptionsBuilder struct {
	Opts []func(*CollectionDeleteWithFilterOptions)
}

// CollectionDeleteWithFilter creates a new CollectionDeleteWithFilterOptionsBuilder instance.
func CollectionDeleteWithFilter() *CollectionDeleteWithFilterOptionsBuilder {
	return &CollectionDeleteWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDeleteWithFilterOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDeleteWithFilterOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionDeleteWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionDeleteWithFilterOptionsBuilder) List() []func(*CollectionDeleteWithFilterOptions) {
	return b.Opts
}

// CollectionCreateIndexOptions contains options for CreateIndex operation.
type CollectionCreateIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionCreateIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionCreateIndexOptionsBuilder is a builder for CollectionCreateIndexOptions.
type CollectionCreateIndexOptionsBuilder struct {
	Opts []func(*CollectionCreateIndexOptions)
}

// CollectionCreateIndex creates a new CollectionCreateIndexOptionsBuilder instance.
func CollectionCreateIndex() *CollectionCreateIndexOptionsBuilder {
	return &CollectionCreateIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionCreateIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionCreateIndexOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionCreateIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionCreateIndexOptionsBuilder) List() []func(*CollectionCreateIndexOptions) {
	return b.Opts
}

// CollectionDropIndexOptions contains options for DropIndex operation.
type CollectionDropIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionDropIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionDropIndexOptionsBuilder is a builder for CollectionDropIndexOptions.
type CollectionDropIndexOptionsBuilder struct {
	Opts []func(*CollectionDropIndexOptions)
}

// CollectionDropIndex creates a new CollectionDropIndexOptionsBuilder instance.
func CollectionDropIndex() *CollectionDropIndexOptionsBuilder {
	return &CollectionDropIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDropIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDropIndexOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionDropIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionDropIndexOptionsBuilder) List() []func(*CollectionDropIndexOptions) {
	return b.Opts
}

// CollectionGetIndexesOptions contains options for GetIndexes operation.
type CollectionGetIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionGetIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionGetIndexesOptionsBuilder is a builder for CollectionGetIndexesOptions.
type CollectionGetIndexesOptionsBuilder struct {
	Opts []func(*CollectionGetIndexesOptions)
}

// CollectionGetIndexes creates a new CollectionGetIndexesOptionsBuilder instance.
func CollectionGetIndexes() *CollectionGetIndexesOptionsBuilder {
	return &CollectionGetIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetIndexesOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetIndexesOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionGetIndexesOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionGetIndexesOptionsBuilder) List() []func(*CollectionGetIndexesOptions) {
	return b.Opts
}

// CollectionGetAllDocIDsOptions contains options for GetAllDocIDs operation.
type CollectionGetAllDocIDsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionGetAllDocIDsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionGetAllDocIDsOptionsBuilder is a builder for CollectionGetAllDocIDsOptions.
type CollectionGetAllDocIDsOptionsBuilder struct {
	Opts []func(*CollectionGetAllDocIDsOptions)
}

// CollectionGetAllDocIDs creates a new CollectionGetAllDocIDsOptionsBuilder instance.
func CollectionGetAllDocIDs() *CollectionGetAllDocIDsOptionsBuilder {
	return &CollectionGetAllDocIDsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetAllDocIDsOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetAllDocIDsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionGetAllDocIDsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionGetAllDocIDsOptionsBuilder) List() []func(*CollectionGetAllDocIDsOptions) {
	return b.Opts
}

// CollectionExistsOptions contains options for Exists operation.
type CollectionExistsOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionExistsOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionExistsOptionsBuilder is a builder for CollectionExistsOptions.
type CollectionExistsOptionsBuilder struct {
	Opts []func(*CollectionExistsOptions)
}

// CollectionExists creates a new CollectionExistsOptionsBuilder instance.
func CollectionExists() *CollectionExistsOptionsBuilder {
	return &CollectionExistsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionExistsOptionsBuilder) SetIdentity(id identity.Identity) *CollectionExistsOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionExistsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionExistsOptionsBuilder) List() []func(*CollectionExistsOptions) {
	return b.Opts
}

// CollectionTruncateOptions contains options for Truncate operation.
type CollectionTruncateOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionTruncateOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionTruncateOptionsBuilder is a builder for CollectionTruncateOptions.
type CollectionTruncateOptionsBuilder struct {
	Opts []func(*CollectionTruncateOptions)
}

// CollectionTruncate creates a new CollectionTruncateOptionsBuilder instance.
func CollectionTruncate() *CollectionTruncateOptionsBuilder {
	return &CollectionTruncateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionTruncateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionTruncateOptionsBuilder {
	b.Opts = append(b.Opts, func(opts *CollectionTruncateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// List returns the list of functional options.
func (b *CollectionTruncateOptionsBuilder) List() []func(*CollectionTruncateOptions) {
	return b.Opts
}
