// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mobile

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// Collection wraps a client.Collection with context and identity information.
type Collection struct {
	col client.Collection
	db  *DB
	txn client.Txn // nil if not transactional
}

// context returns the appropriate context for this collection (with or without txn).
func (c *Collection) colContext() context.Context {
	if c.txn != nil {
		return c.db.contextWithTxn(c.txn)
	}
	return c.db.context()
}

// Name returns the name of this collection.
func (c *Collection) Name() string {
	return c.col.Name()
}

// VersionID returns the version ID of this collection.
func (c *Collection) VersionID() string {
	return c.col.VersionID()
}

// CollectionID returns the collection ID (root) of this collection.
func (c *Collection) CollectionID() string {
	return c.col.CollectionID()
}

// AddDocument adds a new document from a JSON map string.
// encryptDoc enables full document encryption.
// encryptedFieldsJSON is a JSON array of field names to encrypt (may be empty array or "").
func (c *Collection) AddDocument(docJSON string, encryptDoc bool, encryptedFieldsJSON string) error {
	ctx := c.colContext()
	var docMap map[string]any
	if err := json.Unmarshal([]byte(docJSON), &docMap); err != nil {
		return err
	}
	doc, err := client.NewDocFromMap(ctx, docMap, c.col.Version())
	if err != nil {
		return err
	}
	opt := options.AddDocument()
	setOptIdentity(opt, c.db.identity)
	if encryptDoc {
		opt.SetEncryptDoc(true)
	}
	if encryptedFieldsJSON != "" && encryptedFieldsJSON != "[]" {
		var fields []string
		if err := json.Unmarshal([]byte(encryptedFieldsJSON), &fields); err != nil {
			return err
		}
		if len(fields) > 0 {
			opt.SetEncryptedFields(fields)
		}
	}
	return c.col.AddDocument(ctx, doc, opt)
}

// AddManyDocuments adds multiple documents from a JSON array of map strings.
func (c *Collection) AddManyDocuments(docsJSON string, encryptDoc bool, encryptedFieldsJSON string) error {
	ctx := c.colContext()
	var docMaps []map[string]any
	if err := json.Unmarshal([]byte(docsJSON), &docMaps); err != nil {
		return err
	}
	docs := make([]*client.Document, 0, len(docMaps))
	for _, d := range docMaps {
		doc, err := client.NewDocFromMap(ctx, d, c.col.Version())
		if err != nil {
			return err
		}
		docs = append(docs, doc)
	}
	opt := options.AddDocument()
	setOptIdentity(opt, c.db.identity)
	if encryptDoc {
		opt.SetEncryptDoc(true)
	}
	if encryptedFieldsJSON != "" && encryptedFieldsJSON != "[]" {
		var fields []string
		if err := json.Unmarshal([]byte(encryptedFieldsJSON), &fields); err != nil {
			return err
		}
		if len(fields) > 0 {
			opt.SetEncryptedFields(fields)
		}
	}
	return c.col.AddManyDocuments(ctx, docs, opt)
}

// UpdateDocument updates the document with the given docID using the JSON patch string.
func (c *Collection) UpdateDocument(docID string, patch string) error {
	ctx := c.colContext()
	id, err := client.NewDocIDFromString(docID)
	if err != nil {
		return err
	}
	getOpt := options.GetDocument().SetShowDeleted(true)
	setOptIdentity(getOpt, c.db.identity)
	doc, err := c.col.GetDocument(ctx, id, getOpt)
	if err != nil {
		return err
	}
	if err := doc.SetWithJSON(ctx, []byte(patch)); err != nil {
		return err
	}
	opt := options.UpdateDocument()
	setOptIdentity(opt, c.db.identity)
	return c.col.UpdateDocument(ctx, doc, opt)
}

// DeleteDocument deletes the document with the given docID.
// Returns true if the document was deleted.
func (c *Collection) DeleteDocument(docID string) (bool, error) {
	ctx := c.colContext()
	id, err := client.NewDocIDFromString(docID)
	if err != nil {
		return false, err
	}
	opt := options.DeleteDocument()
	setOptIdentity(opt, c.db.identity)
	return c.col.DeleteDocument(ctx, id, opt)
}

// ExistsDocument checks if a document with the given docID exists.
func (c *Collection) ExistsDocument(docID string) (bool, error) {
	ctx := c.colContext()
	id, err := client.NewDocIDFromString(docID)
	if err != nil {
		return false, err
	}
	opt := options.ExistsDocument()
	setOptIdentity(opt, c.db.identity)
	return c.col.ExistsDocument(ctx, id, opt)
}

// UpdateDocumentsWithFilter updates documents matching the filter with the given updater.
// Returns a JSON-encoded UpdateResult.
func (c *Collection) UpdateDocumentsWithFilter(filter string, updater string) (string, error) {
	ctx := c.colContext()
	opt := options.UpdateDocumentsWithFilter()
	setOptIdentity(opt, c.db.identity)
	result, err := c.col.UpdateDocumentsWithFilter(ctx, filter, updater, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(result)
}

// DeleteDocumentsWithFilter soft-deletes documents matching the filter.
// Returns a JSON-encoded DeleteResult.
func (c *Collection) DeleteDocumentsWithFilter(filter string) (string, error) {
	ctx := c.colContext()
	opt := options.DeleteDocumentsWithFilter()
	setOptIdentity(opt, c.db.identity)
	result, err := c.col.DeleteDocumentsWithFilter(ctx, filter, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(result)
}

// GetDocument returns the document with the given docID as a JSON string.
func (c *Collection) GetDocument(docID string, showDeleted bool) (string, error) {
	ctx := c.colContext()
	id, err := client.NewDocIDFromString(docID)
	if err != nil {
		return "", err
	}
	opt := options.GetDocument().SetShowDeleted(showDeleted)
	setOptIdentity(opt, c.db.identity)
	doc, err := c.col.GetDocument(ctx, id, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(doc)
}

// NewIndex creates a new index on this collection.
// requestJSON is a JSON-encoded client.NewIndexRequest. Returns a JSON-encoded IndexDescription.
func (c *Collection) NewIndex(requestJSON string) (string, error) {
	ctx := c.colContext()
	var req client.NewIndexRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return "", err
	}
	opt := options.NewCollectionIndex()
	setOptIdentity(opt, c.db.identity)
	desc, err := c.col.NewIndex(ctx, req, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(desc)
}

// DeleteIndex deletes the named index from this collection.
func (c *Collection) DeleteIndex(name string) error {
	ctx := c.colContext()
	opt := options.DeleteCollectionIndex()
	setOptIdentity(opt, c.db.identity)
	return c.col.DeleteIndex(ctx, name, opt)
}

// ListIndexes returns all indexes on this collection as a JSON array.
func (c *Collection) ListIndexes() (string, error) {
	ctx := c.colContext()
	opt := options.ListCollectionIndexes()
	setOptIdentity(opt, c.db.identity)
	indexes, err := c.col.ListIndexes(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(indexes)
}

// NewEncryptedIndex creates a new encrypted index on this collection.
// requestJSON is a JSON-encoded client.EncryptedIndexDescription.
// Returns a JSON-encoded EncryptedIndexDescription.
func (c *Collection) NewEncryptedIndex(requestJSON string) (string, error) {
	ctx := c.colContext()
	var desc client.EncryptedIndexDescription
	if err := json.Unmarshal([]byte(requestJSON), &desc); err != nil {
		return "", err
	}
	opt := options.NewEncryptedIndex()
	setOptIdentity(opt, c.db.identity)
	result, err := c.col.NewEncryptedIndex(ctx, desc, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(result)
}

// DeleteEncryptedIndex deletes the encrypted index for the given field name.
func (c *Collection) DeleteEncryptedIndex(fieldName string) error {
	ctx := c.colContext()
	opt := options.DeleteEncryptedIndex()
	setOptIdentity(opt, c.db.identity)
	return c.col.DeleteEncryptedIndex(ctx, fieldName, opt)
}

// ListEncryptedIndexes returns all encrypted indexes as a JSON array.
func (c *Collection) ListEncryptedIndexes() (string, error) {
	ctx := c.colContext()
	opt := options.ListCollectionEncryptedIndexes()
	setOptIdentity(opt, c.db.identity)
	indexes, err := c.col.ListEncryptedIndexes(ctx, opt)
	if err != nil {
		return "", err
	}
	return marshalJSON(indexes)
}

// Truncate permanently deletes all document state in this collection on this node.
func (c *Collection) Truncate() error {
	ctx := c.colContext()
	opt := options.TruncateCollection()
	setOptIdentity(opt, c.db.identity)
	return c.col.Truncate(ctx, opt)
}
