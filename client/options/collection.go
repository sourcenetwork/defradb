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

// CollectionCreateOptions contains options for Create and CreateMany operations.
type CollectionCreateOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
	// EncryptDoc enables document encryption when creating a document.
	EncryptDoc bool
	// EncryptedFields specifies a list of fields to be encrypted.
	EncryptedFields []string
}

// CollectionCreate creates a new CollectionCreateOptions instance.
func CollectionCreate() *CollectionCreateOptions {
	return &CollectionCreateOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionCreateOptions) SetIdentity(id identity.Identity) *CollectionCreateOptions {
	o.Identity = immutable.Some(id)
	return o
}

// SetEncryptDoc enables or disables document encryption.
func (o *CollectionCreateOptions) SetEncryptDoc(encrypt bool) *CollectionCreateOptions {
	o.EncryptDoc = encrypt
	return o
}

// SetEncryptedFields specifies fields to be encrypted.
func (o *CollectionCreateOptions) SetEncryptedFields(fields []string) *CollectionCreateOptions {
	o.EncryptedFields = fields
	return o
}

// CollectionUpdateOptions contains options for Update operation.
type CollectionUpdateOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionUpdate creates a new CollectionUpdateOptions instance.
func CollectionUpdate() *CollectionUpdateOptions {
	return &CollectionUpdateOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionUpdateOptions) SetIdentity(id identity.Identity) *CollectionUpdateOptions {
	o.Identity = immutable.Some(id)
	return o
}

// CollectionSaveOptions contains options for Save operation.
type CollectionSaveOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
	// EncryptDoc enables document encryption when creating a document.
	EncryptDoc bool
	// EncryptedFields specifies a list of fields to be encrypted.
	EncryptedFields []string
}

// CollectionSave creates a new CollectionSaveOptions instance.
func CollectionSave() *CollectionSaveOptions {
	return &CollectionSaveOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionSaveOptions) SetIdentity(id identity.Identity) *CollectionSaveOptions {
	o.Identity = immutable.Some(id)
	return o
}

// SetEncryptDoc enables or disables document encryption.
func (o *CollectionSaveOptions) SetEncryptDoc(encrypt bool) *CollectionSaveOptions {
	o.EncryptDoc = encrypt
	return o
}

// SetEncryptedFields specifies fields to be encrypted.
func (o *CollectionSaveOptions) SetEncryptedFields(fields []string) *CollectionSaveOptions {
	o.EncryptedFields = fields
	return o
}

// CollectionDeleteOptions contains options for Delete operation.
type CollectionDeleteOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionDelete creates a new CollectionDeleteOptions instance.
func CollectionDelete() *CollectionDeleteOptions {
	return &CollectionDeleteOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionDeleteOptions) SetIdentity(id identity.Identity) *CollectionDeleteOptions {
	o.Identity = immutable.Some(id)
	return o
}

// CollectionGetOptions contains options for Get operation.
type CollectionGetOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionGet creates a new CollectionGetOptions instance.
func CollectionGet() *CollectionGetOptions {
	return &CollectionGetOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionGetOptions) SetIdentity(id identity.Identity) *CollectionGetOptions {
	o.Identity = immutable.Some(id)
	return o
}

// CollectionUpdateWithFilterOptions contains options for UpdateWithFilter operation.
type CollectionUpdateWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionUpdateWithFilter creates a new CollectionUpdateWithFilterOptions instance.
func CollectionUpdateWithFilter() *CollectionUpdateWithFilterOptions {
	return &CollectionUpdateWithFilterOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionUpdateWithFilterOptions) SetIdentity(id identity.Identity) *CollectionUpdateWithFilterOptions {
	o.Identity = immutable.Some(id)
	return o
}

// CollectionDeleteWithFilterOptions contains options for DeleteWithFilter operation.
type CollectionDeleteWithFilterOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionDeleteWithFilter creates a new CollectionDeleteWithFilterOptions instance.
func CollectionDeleteWithFilter() *CollectionDeleteWithFilterOptions {
	return &CollectionDeleteWithFilterOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionDeleteWithFilterOptions) SetIdentity(id identity.Identity) *CollectionDeleteWithFilterOptions {
	o.Identity = immutable.Some(id)
	return o
}

// CollectionCreateIndexOptions contains options for CreateIndex operation.
type CollectionCreateIndexOptions struct {
	// Identity is the identity of the actor performing the operation.
	// If not set, identity will be retrieved from context.
	Identity immutable.Option[identity.Identity]
}

// CollectionCreateIndex creates a new CollectionCreateIndexOptions instance.
func CollectionCreateIndex() *CollectionCreateIndexOptions {
	return &CollectionCreateIndexOptions{}
}

// SetIdentity sets the identity for the operation.
func (o *CollectionCreateIndexOptions) SetIdentity(id identity.Identity) *CollectionCreateIndexOptions {
	o.Identity = immutable.Some(id)
	return o
}
