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

// AddDocumentOptions contains options for AddDocument and AddManyDocuments operations.
type AddDocumentOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// EncryptDoc enables document encryption when adding a document.
	EncryptDoc bool
	// EncryptedFields specifies a list of fields to be encrypted.
	EncryptedFields []string
}

// GetIdentity returns the identity for the operation.
func (o *AddDocumentOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// AddDocumentOptionsBuilder is a builder for AddDocumentOptions.
type AddDocumentOptionsBuilder struct {
	enumerableBuilder[AddDocumentOptions]
}

// AddDocument creates a new AddDocumentOptionsBuilder instance.
func AddDocument() *AddDocumentOptionsBuilder {
	return &AddDocumentOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *AddDocumentOptionsBuilder) SetIdentity(id identity.Identity) *AddDocumentOptionsBuilder {
	b.append(func(opts *AddDocumentOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetEncryptDoc enables or disables document encryption.
func (b *AddDocumentOptionsBuilder) SetEncryptDoc(encrypt bool) *AddDocumentOptionsBuilder {
	b.append(func(opts *AddDocumentOptions) {
		opts.EncryptDoc = encrypt
	})
	return b
}

// SetEncryptedFields specifies fields to be encrypted.
func (b *AddDocumentOptionsBuilder) SetEncryptedFields(fields []string) *AddDocumentOptionsBuilder {
	b.append(func(opts *AddDocumentOptions) {
		if fields != nil {
			opts.EncryptedFields = make([]string, len(fields))
			copy(opts.EncryptedFields, fields)
		}
	})
	return b
}

// UpdateDocumentOptions contains options for UpdateDocument operation.
type UpdateDocumentOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *UpdateDocumentOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// UpdateDocumentOptionsBuilder is a builder for UpdateDocumentOptions.
type UpdateDocumentOptionsBuilder struct {
	enumerableBuilder[UpdateDocumentOptions]
}

// UpdateDocument creates a new UpdateDocumentOptionsBuilder instance.
func UpdateDocument() *UpdateDocumentOptionsBuilder {
	return &UpdateDocumentOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *UpdateDocumentOptionsBuilder) SetIdentity(id identity.Identity) *UpdateDocumentOptionsBuilder {
	b.append(func(opts *UpdateDocumentOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

type SaveDocumentOptions = AddDocumentOptions

type SaveDocumentOptionsBuilder = AddDocumentOptionsBuilder

// SaveDocument creates a new SaveDocumentOptionsBuilder instance.
func SaveDocument() *SaveDocumentOptionsBuilder {
	return &SaveDocumentOptionsBuilder{}
}

// DeleteDocumentOptions contains options for DeleteDocument operation.
type DeleteDocumentOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteDocumentOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteDocumentOptionsBuilder is a builder for DeleteDocumentOptions.
type DeleteDocumentOptionsBuilder struct {
	enumerableBuilder[DeleteDocumentOptions]
}

// DeleteDocument creates a new DeleteDocumentOptionsBuilder instance.
func DeleteDocument() *DeleteDocumentOptionsBuilder {
	return &DeleteDocumentOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteDocumentOptionsBuilder) SetIdentity(id identity.Identity) *DeleteDocumentOptionsBuilder {
	b.append(func(opts *DeleteDocumentOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// GetDocumentOptions contains options for GetDocument operation.
type GetDocumentOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
	// ShowDeleted specifies whether to return deleted documents.
	ShowDeleted bool
}

// GetIdentity returns the identity for the operation.
func (o *GetDocumentOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// GetDocumentOptionsBuilder is a builder for GetDocumentOptions.
type GetDocumentOptionsBuilder struct {
	enumerableBuilder[GetDocumentOptions]
}

// GetDocument creates a new GetDocumentOptionsBuilder instance.
func GetDocument() *GetDocumentOptionsBuilder {
	return &GetDocumentOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *GetDocumentOptionsBuilder) SetIdentity(id identity.Identity) *GetDocumentOptionsBuilder {
	b.append(func(opts *GetDocumentOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// SetShowDeleted sets whether to return deleted documents.
func (b *GetDocumentOptionsBuilder) SetShowDeleted(showDeleted bool) *GetDocumentOptionsBuilder {
	b.append(func(opts *GetDocumentOptions) {
		opts.ShowDeleted = showDeleted
	})
	return b
}

// UpdateDocumentsWithFilterOptions contains options for UpdateDocumentsWithFilter operation.
type UpdateDocumentsWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *UpdateDocumentsWithFilterOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// UpdateDocumentsWithFilterOptionsBuilder is a builder for UpdateDocumentsWithFilterOptions.
type UpdateDocumentsWithFilterOptionsBuilder struct {
	enumerableBuilder[UpdateDocumentsWithFilterOptions]
}

// UpdateDocumentsWithFilter creates a new UpdateDocumentsWithFilterOptionsBuilder instance.
func UpdateDocumentsWithFilter() *UpdateDocumentsWithFilterOptionsBuilder {
	return &UpdateDocumentsWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *UpdateDocumentsWithFilterOptionsBuilder) SetIdentity(
	id identity.Identity,
) *UpdateDocumentsWithFilterOptionsBuilder {
	b.append(func(opts *UpdateDocumentsWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// DeleteDocumentsWithFilterOptions contains options for DeleteDocumentsWithFilter operation.
type DeleteDocumentsWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteDocumentsWithFilterOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteDocumentsWithFilterOptionsBuilder is a builder for DeleteDocumentsWithFilterOptions.
type DeleteDocumentsWithFilterOptionsBuilder struct {
	enumerableBuilder[DeleteDocumentsWithFilterOptions]
}

// DeleteDocumentsWithFilter creates a new DeleteDocumentsWithFilterOptionsBuilder instance.
func DeleteDocumentsWithFilter() *DeleteDocumentsWithFilterOptionsBuilder {
	return &DeleteDocumentsWithFilterOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteDocumentsWithFilterOptionsBuilder) SetIdentity(
	id identity.Identity,
) *DeleteDocumentsWithFilterOptionsBuilder {
	b.append(func(opts *DeleteDocumentsWithFilterOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// CollectionAddIndexOptions contains options for AddIndex operation.
type CollectionAddIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionAddIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionAddIndexOptionsBuilder is a builder for CollectionAddIndexOptions.
type CollectionAddIndexOptionsBuilder struct {
	enumerableBuilder[CollectionAddIndexOptions]
}

// CollectionAddIndex creates a new CollectionAddIndexOptionsBuilder instance.
func CollectionAddIndex() *CollectionAddIndexOptionsBuilder {
	return &CollectionAddIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionAddIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionAddIndexOptionsBuilder {
	b.append(func(opts *CollectionAddIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// CollectionDeleteIndexOptions contains options for DeleteIndex operation.
type CollectionDeleteIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionDeleteIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionDeleteIndexOptionsBuilder is a builder for CollectionDeleteIndexOptions.
type CollectionDeleteIndexOptionsBuilder struct {
	enumerableBuilder[CollectionDeleteIndexOptions]
}

// CollectionDeleteIndex creates a new CollectionDeleteIndexOptionsBuilder instance.
func CollectionDeleteIndex() *CollectionDeleteIndexOptionsBuilder {
	return &CollectionDeleteIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionDeleteIndexOptionsBuilder) SetIdentity(id identity.Identity) *CollectionDeleteIndexOptionsBuilder {
	b.append(func(opts *CollectionDeleteIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// CollectionListIndexesOptions contains options for ListIndexes operation.
type CollectionListIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *CollectionListIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// CollectionListIndexesOptionsBuilder is a builder for CollectionListIndexesOptions.
type CollectionListIndexesOptionsBuilder struct {
	enumerableBuilder[CollectionListIndexesOptions]
}

// CollectionListIndexes creates a new CollectionListIndexesOptionsBuilder instance.
func CollectionListIndexes() *CollectionListIndexesOptionsBuilder {
	return &CollectionListIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *CollectionListIndexesOptionsBuilder) SetIdentity(id identity.Identity) *CollectionListIndexesOptionsBuilder {
	b.append(func(opts *CollectionListIndexesOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ExistsDocumentOptions contains options for ExistsDocument operation.
type ExistsDocumentOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ExistsDocumentOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ExistsDocumentOptionsBuilder is a builder for ExistsDocumentOptions.
type ExistsDocumentOptionsBuilder struct {
	enumerableBuilder[ExistsDocumentOptions]
}

// ExistsDocument creates a new ExistsDocumentOptionsBuilder instance.
func ExistsDocument() *ExistsDocumentOptionsBuilder {
	return &ExistsDocumentOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ExistsDocumentOptionsBuilder) SetIdentity(id identity.Identity) *ExistsDocumentOptionsBuilder {
	b.append(func(opts *ExistsDocumentOptions) {
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
