// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"

	"github.com/sourcenetwork/defradb/datastore"
)

type EncryptedIndexType string

const (
	EncryptedIndexTypeEquality EncryptedIndexType = "equality"
)

// EncryptedIndexDescription represents metadata for an encrypted index.
type EncryptedIndexDescription struct {
	// FieldName is the name of the field that is being indexed.
	FieldName string

	// Type is the type of searchable encryption.
	// Currently only "equality" is supported.
	Type EncryptedIndexType
}

// EncryptedIndexCreateRequest describes an encrypted index creation request.
type EncryptedIndexCreateRequest struct {
	// FieldName contains the name of the field that is being indexed.
	FieldName string
	// Type is the type of searchable encryption.
	// Currently only "equality" is supported.
	Type EncryptedIndexType
}

// EncryptedCollectionIndex is an interface for indexing documents in a collection.
type EncryptedCollectionIndex interface {
	// Save indexes a document by storing indexed field values.
	// It doesn't retire previous values. For this [Update] should be used.
	Save(context.Context, datastore.Txn, *Document) error
	// Update updates an existing document in the index.
	// It removes the previous indexed field values and stores the new ones.
	Update(context.Context, datastore.Txn, *Document, *Document) error
	// Delete deletes an existing document from the index
	Delete(context.Context, datastore.Txn, *Document) error
	// Name returns the name of the index
	Name() string
	// Description returns the description of the index
	Description() EncryptedIndexDescription
}
