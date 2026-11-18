// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// addDocuments adds the JSON documents in docJSON and gives docs the IDs the
// node assigned them.
func (c *Collection) addDocuments(
	ctx context.Context,
	docs []*client.Document,
	docJSON []byte,
	opt *options.AddDocumentOptions,
) error {
	encrypted := C.int(0)
	if opt.EncryptDoc {
		encrypted = 1
	}
	var args cstrings
	defer args.free()
	chosen := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}
	sel := chosen.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_add(c.target(ctx), args.str(string(docJSON)),
			C.DefraDocumentAddOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
				encrypt:                 encrypted,
				encrypted_fields:        args.str(strings.Join(opt.EncryptedFields, ",")),
				signing:                 chosen.signing(),
			}))
	})
	docIDs, err := unmarshalResult[[]string](res)
	if err != nil {
		return err
	}
	if len(docIDs) != len(docs) {
		return client.NewErrUnexpectedType[[]string]("docIDs", docIDs)
	}
	for i, id := range docIDs {
		docID, err := client.NewDocIDFromString(id)
		if err != nil {
			return err
		}
		client.ApplySavedDocumentID(docs[i], docID)
		docs[i].Clean()
	}
	return nil
}

func (c *Collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	docJSON, err := doc.MarshalJSON()
	if err != nil {
		return err
	}
	return c.addDocuments(ctx, []*client.Document{doc}, docJSON, utils.NewOptions(opts...))
}

func (c *Collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	jsonDocs := make([]json.RawMessage, 0, len(docs))
	for _, doc := range docs {
		docJSON, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to convert document to JSON: %w", err)
		}
		jsonDocs = append(jsonDocs, docJSON)
	}
	docJSON, err := json.Marshal(jsonDocs)
	if err != nil {
		return err
	}
	return c.addDocuments(ctx, docs, docJSON, utils.NewOptions(opts...))
}

func (c *Collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	chosen := collectionOptions{CollectionID: c.def.CollectionID, Signing: opt.EnableSigning}
	sel := chosen.c(&args)
	err = as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_update(c.target(ctx), args.str(string(patch)),
			C.DefraDocumentUpdateOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
				doc_id:                  args.str(doc.ID().String()),
				signing:                 chosen.signing(),
			}))
	}).err()
	if err != nil {
		return err
	}
	doc.Clean()
	return nil
}

func (c *Collection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	if !doc.ID().IsValid() {
		return c.AddDocument(ctx, doc, opts...)
	}

	ctx = c.in(ctx)
	if _, hadTxn := datastore.CtxTryGetClientTxn(ctx); hadTxn {
		return c.saveDocument(ctx, doc, opts...)
	}

	// The existence check and the write it decides share one transaction, so
	// no other writer can add or delete the document in between.
	txn, err := c.w.NewTxn(false)
	if err != nil {
		return err
	}
	defer txn.Discard()
	if err := c.saveDocument(datastore.CtxSetFromClientTxn(ctx, txn), doc, opts...); err != nil {
		return err
	}
	return txn.Commit()
}

// saveDocument adds or updates the document, by whether it exists. The
// caller gives ctx a transaction.
func (c *Collection) saveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	saveOpt := utils.NewOptions(opts...)
	getOpts := options.GetDocument().SetShowDeleted(true)
	if saveOpt.Identity.HasValue() {
		getOpts.SetIdentity(saveOpt.Identity.Value())
	}
	_, err := c.GetDocument(ctx, doc.ID(), getOpts)
	if err == nil {
		updateOpts := options.UpdateDocument()
		if saveOpt.Identity.HasValue() {
			updateOpts.SetIdentity(saveOpt.Identity.Value())
		}
		if saveOpt.EnableSigning.HasValue() {
			updateOpts.SetEnableSigning(saveOpt.EnableSigning.Value())
		}
		return c.UpdateDocument(ctx, doc, updateOpts)
	}
	if strings.Contains(err.Error(), client.ErrDocumentNotFoundOrNotAuthorized.Error()) {
		return c.AddDocument(ctx, doc, opts...)
	}
	return err
}

func (c *Collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	chosen := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}
	sel := chosen.c(&args)
	err := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_delete(c.target(ctx), C.DefraDocumentDeleteOptions{
			identity:                ident,
			collection_name:         sel.name,
			collection_version_id:   sel.version,
			collection_id:           sel.id,
			collection_get_inactive: sel.inactive,
			doc_id:                  args.str(docID.String()),
			signing:                 chosen.signing(),
		}))
	}).err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	var args cstrings
	defer args.free()
	sel := collectionOptions{Version: c.def.VersionID, CollectionID: c.def.CollectionID}.c(&args)
	err := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_get(c.target(ctx), args.str(docID.String()),
			C.DefraDocumentGetOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
			}))
	}).err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Collection) UpdateDocumentsWithFilter(
	ctx context.Context,
	filter any,
	updater string,
	opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
) (*client.UpdateResult, error) {
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	chosen := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}
	sel := chosen.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_update(c.target(ctx), args.str(updater),
			C.DefraDocumentUpdateOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
				filter_json:             args.str(string(filterJSON)),
				signing:                 chosen.signing(),
			}))
	})
	updated, err := unmarshalResult[client.UpdateResult](res)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *Collection) DeleteDocumentsWithFilter(
	ctx context.Context,
	filter any,
	opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
) (*client.DeleteResult, error) {
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	chosen := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}
	sel := chosen.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_delete(c.target(ctx), C.DefraDocumentDeleteOptions{
			identity:                ident,
			collection_name:         sel.name,
			collection_version_id:   sel.version,
			collection_id:           sel.id,
			collection_get_inactive: sel.inactive,
			filter_json:             args.str(string(filterJSON)),
			signing:                 chosen.signing(),
		}))
	})
	deleted, err := unmarshalResult[client.DeleteResult](res)
	if err != nil {
		return nil, err
	}
	return &deleted, nil
}

func (c *Collection) GetDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.GetDocumentOptions],
) (*client.Document, error) {
	opt := utils.NewOptions(opts...)
	showDeleted := C.int(0)
	if opt.ShowDeleted {
		showDeleted = 1
	}
	var args cstrings
	defer args.free()
	sel := c.byName().c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_get(c.target(ctx), args.str(docID.String()),
			C.DefraDocumentGetOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
				show_deleted:            showDeleted,
			}))
	})
	if err := res.err(); err != nil {
		return nil, err
	}

	doc, err := client.NewDocWithID(ctx, docID, c.def)
	if err != nil {
		return nil, err
	}
	if err := doc.SetWithJSON(ctx, []byte(res.Value)); err != nil {
		return nil, err
	}
	doc.Clean()
	return doc, nil
}
