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

// NewCollectionIndexOptions contains options for NewIndex operation.
type NewCollectionIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *NewCollectionIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// NewCollectionIndexOptionsBuilder is a builder for NewCollectionIndexOptions.
type NewCollectionIndexOptionsBuilder struct {
	enumerableBuilder[NewCollectionIndexOptions]
}

// NewCollectionIndex creates a new NewCollectionIndexOptionsBuilder instance.
func NewCollectionIndex() *NewCollectionIndexOptionsBuilder {
	return &NewCollectionIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *NewCollectionIndexOptionsBuilder) SetIdentity(id identity.Identity) *NewCollectionIndexOptionsBuilder {
	b.append(func(opts *NewCollectionIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// DeleteCollectionIndexOptions contains options for DeleteIndex operation.
type DeleteCollectionIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *DeleteCollectionIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// DeleteCollectionIndexOptionsBuilder is a builder for DeleteCollectionIndexOptions.
type DeleteCollectionIndexOptionsBuilder struct {
	enumerableBuilder[DeleteCollectionIndexOptions]
}

// DeleteCollectionIndex creates a new DeleteCollectionIndexOptionsBuilder instance.
func DeleteCollectionIndex() *DeleteCollectionIndexOptionsBuilder {
	return &DeleteCollectionIndexOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *DeleteCollectionIndexOptionsBuilder) SetIdentity(id identity.Identity) *DeleteCollectionIndexOptionsBuilder {
	b.append(func(opts *DeleteCollectionIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ListCollectionIndexesOptions contains options for ListIndexes operation.
type ListCollectionIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ListCollectionIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListCollectionIndexesOptionsBuilder is a builder for ListCollectionIndexesOptions.
type ListCollectionIndexesOptionsBuilder struct {
	enumerableBuilder[ListCollectionIndexesOptions]
}

// ListCollectionIndexes creates a new ListCollectionIndexesOptionsBuilder instance.
func ListCollectionIndexes() *ListCollectionIndexesOptionsBuilder {
	return &ListCollectionIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListCollectionIndexesOptionsBuilder) SetIdentity(id identity.Identity) *ListCollectionIndexesOptionsBuilder {
	b.append(func(opts *ListCollectionIndexesOptions) {
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

// TruncateCollectionOptions contains options for Truncate operation.
type TruncateCollectionOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *TruncateCollectionOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// TruncateCollectionOptionsBuilder is a builder for TruncateCollectionOptions.
type TruncateCollectionOptionsBuilder struct {
	enumerableBuilder[TruncateCollectionOptions]
}

// TruncateCollection creates a new TruncateCollectionOptionsBuilder instance.
func TruncateCollection() *TruncateCollectionOptionsBuilder {
	return &TruncateCollectionOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *TruncateCollectionOptionsBuilder) SetIdentity(id identity.Identity) *TruncateCollectionOptionsBuilder {
	b.append(func(opts *TruncateCollectionOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// NewEncryptedIndexOptions contains options for NewEncryptedIndex operation.
type NewEncryptedIndexOptions struct {
	Identity immutable.Option[identity.Identity]
}

func (o *NewEncryptedIndexOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

type NewEncryptedIndexOptionsBuilder struct {
	enumerableBuilder[NewEncryptedIndexOptions]
}

func NewEncryptedIndex() *NewEncryptedIndexOptionsBuilder {
	return &NewEncryptedIndexOptionsBuilder{}
}

func (b *NewEncryptedIndexOptionsBuilder) SetIdentity(id identity.Identity) *NewEncryptedIndexOptionsBuilder {
	b.append(func(opts *NewEncryptedIndexOptions) {
		opts.Identity = immutable.Some(id)
	})
	return b
}

// ListCollectionEncryptedIndexesOptions contains options for ListEncryptedIndexes operation.
type ListCollectionEncryptedIndexesOptions struct {
	// Identity is the identity of the actor performing the operation.
	Identity immutable.Option[identity.Identity]
}

// GetIdentity returns the identity for the operation.
func (o *ListCollectionEncryptedIndexesOptions) GetIdentity() immutable.Option[identity.Identity] {
	return o.Identity
}

// ListCollectionEncryptedIndexesOptionsBuilder is a builder for ListCollectionEncryptedIndexesOptions.
type ListCollectionEncryptedIndexesOptionsBuilder struct {
	enumerableBuilder[ListCollectionEncryptedIndexesOptions]
}

// ListCollectionEncryptedIndexes creates a new ListCollectionEncryptedIndexesOptionsBuilder instance.
func ListCollectionEncryptedIndexes() *ListCollectionEncryptedIndexesOptionsBuilder {
	return &ListCollectionEncryptedIndexesOptionsBuilder{}
}

// SetIdentity sets the identity for the operation.
func (b *ListCollectionEncryptedIndexesOptionsBuilder) SetIdentity(
	id identity.Identity,
) *ListCollectionEncryptedIndexesOptionsBuilder {
	b.append(func(opts *ListCollectionEncryptedIndexesOptions) {
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
