// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encryption"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// docIDResult wraps the result of an attempt at a DocID retrieval operation.
type docIDResult struct {
	ID  client.DocID
	Err error
}

func (c *collection) getAllDocIDsChan(
	ctx context.Context,
) (<-chan docIDResult, error) {
	shortID, err := id.GetUncachedShortCollectionID(ctx, c.Version().CollectionID, c.db.Multistore().Systemstore())
	if err != nil {
		return nil, err
	}
	prefix := keys.PrimaryDataStoreKey{ // empty path for all keys prefix
		CollectionShortID: shortID,
	}
	iter, err := c.db.Multistore().Datastore().Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	resCh := make(chan docIDResult)
	go func() {
		closeIterator := func() {
			if err := iter.Close(); err != nil {
				log.ErrorContextE(ctx, errFailedtoCloseQueryReqAllIDs, err)
			}
		}
		defer func() {
			closeIterator()
			close(resCh)
		}()
		for {
			// check for Done on context first
			select {
			case <-ctx.Done():
				// we've been cancelled! ;)
				return
			default:
				// noop, just continue on the with the for loop
			}

			hasNext, err := iter.Next()
			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}
			if !hasNext {
				break
			}

			splitString := strings.Split(string(iter.Key()), "/")
			rawDocID := splitString[len(splitString)-1]

			docID, err := client.NewDocIDFromString(rawDocID)
			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}

			canRead, err := c.checkAccessOfDocWithACP(
				ctx,
				acpTypes.DocumentReadPerm,
				docID.String(),
			)

			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}

			if canRead {
				resCh <- docIDResult{
					ID: docID,
				}
			}
		}
	}()

	return resCh, nil
}

// AddDocument adds a new document.
// Will verify the DocID/CID to ensure that the new document is correctly formatted.
func (c *collection) AddDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	err = c.add(ctx, doc, opt)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// AddManyDocuments adds a collection of documents at once.
// Will verify the DocID/CID to ensure that the new documents are correctly formatted.
func (c *collection) AddManyDocuments(
	ctx context.Context,
	docs []*client.Document,
	opts ...options.Enumerable[options.AddDocumentOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	for _, doc := range docs {
		err = c.add(ctx, doc, opt)
		if err != nil {
			return err
		}
	}
	return txn.Commit()
}

func (c *collection) getDocIDAndPrimaryKeyFromDoc(
	ctx context.Context,
	doc *client.Document,
) (client.DocID, keys.PrimaryDataStoreKey, error) {
	docID, err := doc.GenerateDocID()
	if err != nil {
		return client.DocID{}, keys.PrimaryDataStoreKey{}, err
	}

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return client.DocID{}, keys.PrimaryDataStoreKey{}, err
	}

	if primaryKey.DocID != doc.ID().String() {
		return client.DocID{}, keys.PrimaryDataStoreKey{},
			NewErrDocVerification(doc.ID().String(), primaryKey.DocID)
	}
	return docID, primaryKey, nil
}

func (c *collection) add(
	ctx context.Context,
	doc *client.Document,
	opt *options.AddDocumentOptions,
) error {
	err := c.setEmbedding(ctx, doc, true)
	if err != nil {
		return err
	}

	docID, primaryKey, err := c.getDocIDAndPrimaryKeyFromDoc(ctx, doc)
	if err != nil {
		return err
	}

	// check if doc already exists
	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil {
		return err
	}
	if exists {
		return NewErrDocumentAlreadyExists(primaryKey.DocID)
	}
	if isDeleted {
		return NewErrDocumentDeleted(primaryKey.DocID)
	}

	// write value object marker if we have an empty doc
	if len(doc.Values()) == 0 {
		txn := datastore.CtxMustGetTxn(ctx)

		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}

		valueKey := keys.DataStoreKey{
			CollectionShortID: shortID,
			DocID:             docID.String(),
			InstanceType:      keys.ValueKey,
		}

		err = txn.Datastore().Set(ctx, valueKey, []byte{base.ObjectMarker})
		if err != nil {
			return err
		}
	}

	ctx = setContextDocEncryption(ctx, opt)

	// write data to DB via MerkleClock/CRDT
	err = c.save(ctx, doc, true)
	if err != nil {
		return err
	}

	err = c.addDocToIndex(ctx, doc)
	if err != nil {
		return err
	}

	return c.registerDocWithACP(ctx, doc.ID().String())
}

func setContextDocEncryption(
	ctx context.Context,
	opt *options.AddDocumentOptions,
) context.Context {
	if !opt.EncryptDoc && len(opt.EncryptedFields) == 0 {
		return ctx
	}
	ctx = encryption.SetContextConfigFromParams(ctx, opt.EncryptDoc, opt.EncryptedFields)
	return ctx
}

// UpdateDocument updates an existing document with the new values.
// Any field that needs to be removed or cleared should call doc.Clear(field) before.
// Any field that is nil/empty that hasn't called Clear will be ignored.
func (c *collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, doc.ID())
	if err != nil {
		return err
	}

	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil {
		return err
	}
	if !exists {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}
	if isDeleted {
		return NewErrDocumentDeleted(primaryKey.DocID)
	}

	err = c.update(ctx, doc)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// Contract: DB Exists check is already performed, and a doc with the given ID exists.
// Note: Should we CompareAndSet the update, IE: Query(read-only) the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (c *collection) update(
	ctx context.Context,
	doc *client.Document,
) error {
	// Stop the update if the correct permissions aren't there.
	canUpdate, err := c.checkAccessOfDocWithACP(
		ctx,
		acpTypes.DocumentUpdatePerm,
		doc.ID().String(),
	)
	if err != nil {
		return err
	}
	if !canUpdate {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}

	err = c.setEmbedding(ctx, doc, false)
	if err != nil {
		return err
	}

	err = c.save(ctx, doc, false)
	if err != nil {
		return err
	}
	return nil
}

// SaveDocument saves a document into the db.
// Either by creating a new document or by updating an existing one.
func (c *collection) SaveDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.SaveDocumentOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	// Check if document already exists with primary DS key.
	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, doc.ID())
	if err != nil {
		return err
	}

	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil {
		return err
	}

	if isDeleted {
		return NewErrDocumentDeleted(doc.ID().String())
	}

	if exists {
		err = c.update(ctx, doc)
	} else {
		err = c.add(ctx, doc, opt)
	}
	if err != nil {
		return err
	}

	return txn.Commit()
}

// hasPrivateKey checks if the identity is a FullIdentity and has a non-nil private key.
func hasPrivateKey(ident identity.Identity) bool {
	if fullIdent, ok := ident.(identity.FullIdentity); ok {
		return fullIdent.PrivateKey() != nil
	}
	return false
}

func (c *collection) validateEncryptedFields(ctx context.Context) error {
	encConf := encryption.GetContextConfig(ctx)
	if !encConf.HasValue() {
		return nil
	}
	fields := encConf.Value().EncryptedFields
	if len(fields) == 0 {
		return nil
	}

	for _, field := range fields {
		if _, exists := c.Version().GetFieldByName(field); !exists {
			return client.NewErrFieldNotExist(field)
		}
		if strings.HasPrefix(field, "_") {
			return NewErrCanNotEncryptBuiltinField(field)
		}
	}
	return nil
}

// save saves the document state. save MUST not be called outside the `c.add`
// and `c.update` methods as we wrap the acp logic within those methods. Calling
// save elsewhere could cause the omission of acp checks.
func (c *collection) save(
	ctx context.Context,
	doc *client.Document,
	isAdd bool,
) error {
	if err := c.validateEncryptedFields(ctx); err != nil {
		return err
	}

	if !isAdd {
		err := c.updateIndexedDoc(ctx, doc)
		if err != nil {
			return err
		}
	}
	txn := datastore.CtxMustGetTxn(ctx)

	ident := iIdentity.FromContext(ctx)
	if (!ident.HasValue() || !hasPrivateKey(ident.Value())) && c.db.nodeIdentity.HasValue() {
		ctx = iIdentity.WithContext(ctx, c.db.nodeIdentity)
	}

	if !c.db.signingDisabled {
		ctx = coreblock.ContextWithEnabledSigning(ctx)
	}

	// NOTE: We delay the final Clean() call until we know
	// the commit on the transaction is successful. If we didn't
	// wait, and just did it here, then *if* the commit fails down
	// the line, then we have no way to roll back the state
	// side-effect on the document func called here.
	txn.OnSuccess(func() {
		doc.Clean()
	})

	shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
	if err != nil {
		return err
	}

	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instantiate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	primaryKey := keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocID:             doc.ID().String(),
	}

	links := make([]coreblock.DAGLink, 0)
	for k, v := range doc.Fields() {
		val, err := doc.GetValueWithField(v)
		if err != nil {
			return err
		}

		if val.IsDirty() {
			fieldDescription, valid := c.Version().GetFieldByName(k)
			if !valid {
				return client.NewErrFieldNotExist(k)
			}

			fieldID, err := id.GetShortFieldID(ctx, shortID, fieldDescription.FieldID)
			if err != nil {
				return err
			}
			fieldKey := keys.DataStoreKey{
				CollectionShortID: shortID,
				DocID:             primaryKey.DocID,
				FieldID:           strconv.FormatUint(uint64(fieldID), 10),
			}

			// by default the type will have been set to LWW_REGISTER. We need to ensure
			// that it's set to the same as the field description CRDT type.
			val.SetType(fieldDescription.Typ)

			merkleCRDT, err := crdt.FieldLevelCRDTWithStore(
				txn.Datastore(),
				c.VersionID(),
				val.Type(),
				fieldDescription.Kind,
				fieldKey,
				fieldDescription.Name,
			)
			if err != nil {
				return err
			}

			delta, err := merkleCRDT.Delta(ctx, crdt.NewDocField(primaryKey.DocID, k, val))
			if err != nil {
				return err
			}

			link, _, err := coreblock.AddDelta(ctx, merkleCRDT, delta)
			if err != nil {
				return err
			}

			links = append(links, coreblock.NewDAGLink(k, link))
		}
	}

	merkleCRDT := crdt.NewDocComposite(
		txn.Datastore(),
		c.Version().VersionID,
		primaryKey.ToDataStoreKey().WithFieldID(core.COMPOSITE_NAMESPACE),
	)

	link, headNode, err := coreblock.AddDelta(ctx, merkleCRDT, merkleCRDT.Delta(), links...)
	if err != nil {
		return err
	}

	// publish an update event when the txn succeeds
	updateEvent := event.Update{
		DocID:        doc.ID().String(),
		Cid:          link.Cid,
		CollectionID: c.Version().CollectionID,
		Block:        headNode,
	}
	txn.OnSuccess(func() {
		c.db.sendUpdate(updateEvent)
	})

	txn.OnSuccess(func() {
		doc.SetHead(link.Cid)
	})

	if c.def.IsBranchable {
		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}
		collectionCRDT := crdt.NewCollection(
			c.Version().VersionID,
			keys.NewHeadstoreColKey(shortID),
		)

		link, headNode, err := coreblock.AddDelta(
			ctx,
			collectionCRDT,
			collectionCRDT.Delta(),
			[]coreblock.DAGLink{{Link: link}}...,
		)
		if err != nil {
			return err
		}

		updateEvent := event.Update{
			Cid:          link.Cid,
			CollectionID: c.Version().CollectionID,
			Block:        headNode,
		}

		txn.OnSuccess(func() {
			c.db.sendUpdate(updateEvent)
		})
	}

	return nil
}

// DeleteDocument will attempt to delete a document by docID and return true if a deletion is successful,
// otherwise will return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given DocID. This includes data, block, and head storage.
func (c *collection) DeleteDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.DeleteDocumentOptions],
) (bool, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteDocumentPerm); err != nil {
		return false, err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return false, err
	}

	err = c.deleteIndexedDocWithID(ctx, docID)
	if err != nil {
		return false, err
	}

	err = c.applyDelete(ctx, primaryKey)
	if err != nil {
		return false, err
	}
	return true, txn.Commit()
}

// ExistsDocument checks if a given document exists with supplied DocID.
func (c *collection) ExistsDocument(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.ExistsDocumentOptions],
) (bool, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeReadDocumentPerm); err != nil {
		return false, err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return false, err
	}

	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return false, err
	}
	return exists && !isDeleted, txn.Commit()
}

// check if a document exists with the given primary key
func (c *collection) exists(
	ctx context.Context,
	primaryKey keys.PrimaryDataStoreKey,
) (exists bool, isDeleted bool, err error) {
	canRead, err := c.checkAccessOfDocWithACP(
		ctx,
		acpTypes.DocumentReadPerm,
		primaryKey.DocID,
	)
	if err != nil {
		return false, false, err
	} else if !canRead {
		return false, false, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	val, err := txn.Datastore().Get(ctx, primaryKey)
	if err != nil && errors.Is(err, corekv.ErrNotFound) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}
	if bytes.Equal(val, []byte{base.DeletedObjectMarker}) {
		return true, true, nil
	}

	return true, false, nil
}

func (c *collection) getPrimaryKeyFromDocID(
	ctx context.Context,
	docID client.DocID,
) (keys.PrimaryDataStoreKey, error) {
	shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
	if err != nil {
		return keys.PrimaryDataStoreKey{}, err
	}

	return keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocID:             docID.String(),
	}, nil
}
