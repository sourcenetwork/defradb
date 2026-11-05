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
	ctx = c.context(ctx)
	encrypted := C.int(0)
	if opt.EncryptDoc {
		encrypted = 1
	}
	var args cstrings
	defer args.free()
	copts := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_add(c.w.target(ctx), args.str(string(docJSON)), encrypted,
			args.str(strings.Join(opt.EncryptedFields, ",")), copts, ident))
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
	ctx = c.context(ctx)
	patch, err := doc.ToJSONPatch()
	if err != nil {
		return err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	copts := collectionOptions{CollectionID: c.def.CollectionID, Signing: opt.EnableSigning}.c(&args)
	err = as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_update(c.w.target(ctx), args.str(doc.ID().String()),
			args.str(""), args.str(string(patch)), copts, ident))
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

	ctx = c.context(ctx)
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
	ctx = c.context(ctx)
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	copts := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}.c(&args)
	err := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_delete(c.w.target(ctx), args.str(docID.String()),
			args.str(""), copts, ident))
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
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	copts := collectionOptions{Version: c.def.VersionID, CollectionID: c.def.CollectionID}.c(&args)
	err := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_get(c.w.target(ctx), args.str(docID.String()), 0, copts, ident))
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
	ctx = c.context(ctx)
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	copts := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_update(c.w.target(ctx), args.str(""),
			args.str(string(filterJSON)), args.str(updater), copts, ident))
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
	ctx = c.context(ctx)
	filterJSON, err := json.Marshal(utils.NormalizeFilterForJSON(filter))
	if err != nil {
		return nil, err
	}
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	copts := collectionOptions{Name: c.def.Name, Signing: opt.EnableSigning}.c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_delete(c.w.target(ctx), args.str(""),
			args.str(string(filterJSON)), copts, ident))
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
	ctx = c.context(ctx)
	opt := utils.NewOptions(opts...)
	showDeleted := C.int(0)
	if opt.ShowDeleted {
		showDeleted = 1
	}
	var args cstrings
	defer args.free()
	copts := c.byName().c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_document_get(c.w.target(ctx), args.str(docID.String()), showDeleted, copts, ident))
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
