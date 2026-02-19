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
	enumerableBuilder[CollectionCreateOptions]
}

// CollectionCreate creates a new CollectionCreateOptionsBuilder instance.
func CollectionCreate() *CollectionCreateOptionsBuilder {
	return &CollectionCreateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionCreateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionCreateOptionsBuilder {
	b.append(func(opts *CollectionCreateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetEncryptDoc enables or disables document encryption.
func (b *CollectionCreateOptionsBuilder) SetEncryptDoc(encrypt bool) *CollectionCreateOptionsBuilder {
	b.append(func(opts *CollectionCreateOptions) {
		opts.EncryptDoc = encrypt
	})
	return b
}

// SetEncryptedFields specifies fields to be encrypted.
func (b *CollectionCreateOptionsBuilder) SetEncryptedFields(fields []string) *CollectionCreateOptionsBuilder {
	b.append(func(opts *CollectionCreateOptions) {
		if fields != nil {
			opts.EncryptedFields = make([]string, len(fields))
			copy(opts.EncryptedFields, fields)
		}
	})
	return b
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
	enumerableBuilder[CollectionUpdateOptions]
}

// CollectionUpdate creates a new CollectionUpdateOptionsBuilder instance.
func CollectionUpdate() *CollectionUpdateOptionsBuilder {
	return &CollectionUpdateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionUpdateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionUpdateOptionsBuilder {
	b.append(func(opts *CollectionUpdateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

type CollectionSaveOptions = CollectionCreateOptions

type CollectionSaveOptionsBuilder = CollectionCreateOptionsBuilder

// CollectionSave creates a new CollectionSaveOptionsBuilder instance.
func CollectionSave() *CollectionSaveOptionsBuilder {
	return &CollectionSaveOptionsBuilder{}
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
	enumerableBuilder[CollectionDeleteOptions]
}

// CollectionDelete creates a new CollectionDeleteOptionsBuilder instance.
func CollectionDelete() *CollectionDeleteOptionsBuilder {
	return &CollectionDeleteOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDeleteOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDeleteOptionsBuilder {
	b.append(func(opts *CollectionDeleteOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// CollectionGetOptions contains options for Get operation.
type CollectionGetOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// ShowDeleted specifies whether to return deleted documents.
	ShowDeleted bool
}

// GetIdentity returns the identity for the operation.
func (o *CollectionGetOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionGetOptionsBuilder is a builder for CollectionGetOptions.
type CollectionGetOptionsBuilder struct {
	enumerableBuilder[CollectionGetOptions]
}

// CollectionGet creates a new CollectionGetOptionsBuilder instance.
func CollectionGet() *CollectionGetOptionsBuilder {
	return &CollectionGetOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetOptionsBuilder {
	b.append(func(opts *CollectionGetOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetShowDeleted sets whether to return deleted documents.
func (b *CollectionGetOptionsBuilder) SetShowDeleted(showDeleted bool) *CollectionGetOptionsBuilder {
	b.append(func(opts *CollectionGetOptions) {
		opts.ShowDeleted = showDeleted
	})
	return b
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
	enumerableBuilder[CollectionUpdateWithFilterOptions]
}

// CollectionUpdateWithFilter creates a new CollectionUpdateWithFilterOptionsBuilder instance.
func CollectionUpdateWithFilter() *CollectionUpdateWithFilterOptionsBuilder {
	return &CollectionUpdateWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionUpdateWithFilterOptionsBuilder) SetIdentity(
	id identity.Identity,
) *CollectionUpdateWithFilterOptionsBuilder {
	b.append(func(opts *CollectionUpdateWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionDeleteWithFilterOptions]
}

// CollectionDeleteWithFilter creates a new CollectionDeleteWithFilterOptionsBuilder instance.
func CollectionDeleteWithFilter() *CollectionDeleteWithFilterOptionsBuilder {
	return &CollectionDeleteWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDeleteWithFilterOptionsBuilder) SetIdentity(
	id identity.Identity,
) *CollectionDeleteWithFilterOptionsBuilder {
	b.append(func(opts *CollectionDeleteWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionCreateIndexOptions]
}

// CollectionCreateIndex creates a new CollectionCreateIndexOptionsBuilder instance.
func CollectionCreateIndex() *CollectionCreateIndexOptionsBuilder {
	return &CollectionCreateIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionCreateIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionCreateIndexOptionsBuilder {
	b.append(func(opts *CollectionCreateIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionDropIndexOptions]
}

// CollectionDropIndex creates a new CollectionDropIndexOptionsBuilder instance.
func CollectionDropIndex() *CollectionDropIndexOptionsBuilder {
	return &CollectionDropIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDropIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDropIndexOptionsBuilder {
	b.append(func(opts *CollectionDropIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionGetIndexesOptions]
}

// CollectionGetIndexes creates a new CollectionGetIndexesOptionsBuilder instance.
func CollectionGetIndexes() *CollectionGetIndexesOptionsBuilder {
	return &CollectionGetIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetIndexesOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetIndexesOptionsBuilder {
	b.append(func(opts *CollectionGetIndexesOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionGetAllDocIDsOptions]
}

// CollectionGetAllDocIDs creates a new CollectionGetAllDocIDsOptionsBuilder instance.
func CollectionGetAllDocIDs() *CollectionGetAllDocIDsOptionsBuilder {
	return &CollectionGetAllDocIDsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionGetAllDocIDsOptionsBuilder) SetIdentity(id identity.Identity) *CollectionGetAllDocIDsOptionsBuilder {
	b.append(func(opts *CollectionGetAllDocIDsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionExistsOptions]
}

// CollectionExists creates a new CollectionExistsOptionsBuilder instance.
func CollectionExists() *CollectionExistsOptionsBuilder {
	return &CollectionExistsOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionExistsOptionsBuilder) SetIdentity(id identity.Identity) *CollectionExistsOptionsBuilder {
	b.append(func(opts *CollectionExistsOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
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
	enumerableBuilder[CollectionTruncateOptions]
}

// CollectionTruncate creates a new CollectionTruncateOptionsBuilder instance.
func CollectionTruncate() *CollectionTruncateOptionsBuilder {
	return &CollectionTruncateOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionTruncateOptionsBuilder) SetIdentity(id identity.Identity) *CollectionTruncateOptionsBuilder {
	b.append(func(opts *CollectionTruncateOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// AddEncryptedIndexOptions contains options for AddEncryptedIndex operation.
type AddEncryptedIndexOptions struct {
	Identity immutable.Option[identity.Identity]
}

func (o *AddEncryptedIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

type AddEncryptedIndexOptionsBuilder struct {
	enumerableBuilder[AddEncryptedIndexOptions]
}

func AddEncryptedIndex() *AddEncryptedIndexOptionsBuilder {
	return &AddEncryptedIndexOptionsBuilder{}
}

func (b *AddEncryptedIndexOptionsBuilder) SetIdentity(id identity.Identity) *AddEncryptedIndexOptionsBuilder {
	b.append(func(opts *AddEncryptedIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// CollectionListEncryptedIndexesOptions contains options for ListEncryptedIndexes operation.
type CollectionListEncryptedIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionListEncryptedIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionListEncryptedIndexesOptionsBuilder is a builder for CollectionListEncryptedIndexesOptions.
type CollectionListEncryptedIndexesOptionsBuilder struct {
	enumerableBuilder[CollectionListEncryptedIndexesOptions]
}

// CollectionListEncryptedIndexes creates a new CollectionListEncryptedIndexesOptionsBuilder instance.
func CollectionListEncryptedIndexes() *CollectionListEncryptedIndexesOptionsBuilder {
	return &CollectionListEncryptedIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionListEncryptedIndexesOptionsBuilder) SetIdentity(
	id identity.Identity,
) *CollectionListEncryptedIndexesOptionsBuilder {
	b.append(func(opts *CollectionListEncryptedIndexesOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// DeleteEncryptedIndexOptions contains options for DeleteEncryptedIndex operation.
type DeleteEncryptedIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteEncryptedIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteEncryptedIndexOptionsBuilder is a builder for DeleteEncryptedIndexOptions.
type DeleteEncryptedIndexOptionsBuilder struct {
	enumerableBuilder[DeleteEncryptedIndexOptions]
}

// DeleteEncryptedIndex creates a new DeleteEncryptedIndexOptionsBuilder instance.
func DeleteEncryptedIndex() *DeleteEncryptedIndexOptionsBuilder {
	return &DeleteEncryptedIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteEncryptedIndexOptionsBuilder) SetIdentity(id identity.Identity) *DeleteEncryptedIndexOptionsBuilder {
	b.append(func(opts *DeleteEncryptedIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}
