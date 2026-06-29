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

	"github.com/ipfs/go-cid"

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
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encryption"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/immutable"
)

// docIDResult wraps the result of an attempt at a DocID retrieval operation.
type docIDResult struct {
	ID  client.DocID
	Err error
}

func (c *collection) getAllDocIDsChan(
	ctx context.Context,
) (<-chan docIDResult, error) {
	systemstore := c.db.Multistore().Systemstore()
	collectionShortID, err := id.GetUncachedCollectionShortID(ctx, c.Version().CollectionID, systemstore)
	if err != nil {
		return nil, err
	}
	prefix := keys.PrimaryDataStoreKey{CollectionShortID: collectionShortID}
	iter, err := c.db.Multistore().Datastore().Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrGetAllDocIDs(err)
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

			key, err := keys.NewPrimaryDataStoreKey(string(iter.Key()))
			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}

			docIDString, found, err := id.GetDocIDFromStore(ctx, systemstore, key.DocShortID)
			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}
			if !found {
				continue
			}
			docID, err := client.NewDocIDFromString(docIDString)
			if err != nil {
				closeIterator()
				resCh <- docIDResult{
					Err: err,
				}
				return
			}

			canRead, err := c.checkAccessOfDoc(
				ctx,
				acpTypes.DocumentReadPerm,
				docIDString,
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
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)
	ctx = setContextSigning(ctx, c.db.signingDisabled, opt.EnableSigning)

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
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)
	ctx = setContextSigning(ctx, c.db.signingDisabled, opt.EnableSigning)

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

func (c *collection) add(
	ctx context.Context,
	doc *client.Document,
	opt *options.AddDocumentOptions,
) error {
	err := c.setEmbedding(ctx, doc, true)
	if err != nil {
		return err
	}

	ctx = setContextDocEncryption(ctx, opt)

	err = c.save(ctx, doc, true)
	if err != nil {
		return err
	}

	err = c.addDocToIndex(ctx, doc)
	if err != nil {
		return err
	}

	return c.registerDoc(ctx, doc.ID().String())
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

func setContextSigning(
	ctx context.Context,
	signingDisabled bool,
	enableSigning immutable.Option[bool],
) context.Context {
	enabled := !signingDisabled
	if enableSigning.HasValue() {
		enabled = enableSigning.Value()
	}
	return coreblock.ContextWithSigning(ctx, enabled)
}

// UpdateDocument updates an existing document with the new values.
// Any field that needs to be removed or cleared should call doc.Clear(field) before.
// Any field that is nil/empty that hasn't called Clear will be ignored.
func (c *collection) UpdateDocument(
	ctx context.Context,
	doc *client.Document,
	opts ...options.Enumerable[options.UpdateDocumentOptions],
) error {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)
	ctx = setContextSigning(ctx, c.db.signingDisabled, opt.EnableSigning)

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
		return NewErrDocumentDeleted(doc.ID().String())
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
	canUpdate, err := c.checkAccessOfDoc(
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
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeUpdateDocumentPerm); err != nil {
		return err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)
	ctx = setContextSigning(ctx, c.db.signingDisabled, opt.EnableSigning)

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

func (c *collection) contextForSigning(ctx context.Context) context.Context {
	signingCtx := ctx
	ident := iIdentity.FromContext(signingCtx)
	if (!ident.HasValue() || !hasPrivateKey(ident.Value())) && c.db.nodeIdentity.HasValue() {
		signingCtx = iIdentity.WithContext(signingCtx, c.db.nodeIdentity)
	}

	if enabled, ok := coreblock.SigningConfigFromContext(signingCtx); ok && enabled {
		signingCtx = coreblock.ContextWithEnabledSigning(signingCtx)
	}
	return signingCtx
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

	signingCtx := c.contextForSigning(ctx)

	// NOTE: We delay the final Clean() call until we know
	// the commit on the transaction is successful. If we didn't
	// wait, and just did it here, then *if* the commit fails down
	// the line, then we have no way to roll back the state
	// side-effect on the document func called here.
	txn.OnSuccess(func() {
		doc.Clean()
	})

	collectionShortID, err := id.GetCollectionShortID(ctx, c.Version().CollectionID)
	if err != nil {
		return err
	}

	var primaryKey keys.PrimaryDataStoreKey
	if isAdd {
		docShortID, err := id.NextDocShortID(ctx)
		if err != nil {
			return err
		}
		primaryKey = keys.PrimaryDataStoreKey{
			CollectionShortID: collectionShortID,
			DocShortID:        docShortID,
		}
	} else {
		primaryKey, err = c.getPrimaryKeyFromDocID(ctx, doc.ID())
		if err != nil {
			return err
		}
	}

	encryptionDocID := keys.EncodeDocRef(collectionShortID, primaryKey.DocShortID)

	links := make([]coreblock.DAGLink, 0)
	encryptionCIDs := make([]cid.Cid, 0)
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

			fieldID, err := id.GetShortFieldID(ctx, collectionShortID, fieldDescription.FieldID)
			if err != nil {
				return err
			}
			fieldKey := keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        primaryKey.DocShortID,
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
			delta, err := merkleCRDT.Delta(ctx, crdt.NewDocField(k, val))
			if err != nil {
				return err
			}

			link, rawBlock, err := coreblock.AddDeltaWithOptions(
				signingCtx,
				merkleCRDT,
				delta,
				coreblock.AddDeltaOptions{EncryptionDocKey: encryptionDocID},
			)
			if err != nil {
				return err
			}

			links = append(links, coreblock.NewDAGLink(k, link))
			encryptionCIDs, err = appendEncryptionCID(encryptionCIDs, rawBlock)
			if err != nil {
				return err
			}
		}
	}

	merkleCRDT := crdt.NewDocComposite(
		txn.Datastore(),
		c.Version().VersionID,
		primaryKey.ToDataStoreKey().WithFieldID(core.COMPOSITE_NAMESPACE),
	)
	link, headNode, err := coreblock.AddDeltaWithOptions(
		signingCtx,
		merkleCRDT,
		merkleCRDT.Delta(),
		coreblock.AddDeltaOptions{EncryptionDocKey: encryptionDocID},
		links...,
	)
	if err != nil {
		return err
	}
	encryptionCIDs, err = appendEncryptionCID(encryptionCIDs, headNode)
	if err != nil {
		return err
	}

	updateDocID := doc.ID().String()
	if isAdd {
		docID := client.NewDocIDV0(link.Cid)
		docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, docID.String())
		if err != nil {
			return err
		}
		if found {
			existingKey := keys.PrimaryDataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docShortID,
			}
			exists, isDeleted, err := c.exists(ctx, existingKey)
			if err != nil {
				return err
			}
			if isDeleted {
				return NewErrDocumentDeleted(docID.String())
			}
			if exists || docShortID != primaryKey.DocShortID {
				return NewErrDocumentAlreadyExists(docID.String())
			}
		}
		if err := id.SetDocIDMapping(ctx, collectionShortID, primaryKey.DocShortID, docID.String()); err != nil {
			return err
		}
		client.ApplySavedDocumentID(doc, docID)
		updateDocID = docID.String()
	}
	if err := id.SetBlockDocIDMapping(ctx, link.Cid, updateDocID); err != nil {
		return err
	}
	for _, link := range links {
		if err := id.SetBlockDocIDMapping(ctx, link.Cid, updateDocID); err != nil {
			return err
		}
	}
	for _, encCID := range encryptionCIDs {
		if err := id.SetBlockDocIDMapping(ctx, encCID, updateDocID); err != nil {
			return err
		}
	}

	// publish an update event when the txn succeeds
	updateEvent := event.Update{
		DocID:        updateDocID,
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
		collectionShortID, err := id.GetCollectionShortID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}
		collectionCRDT := crdt.NewCollection(
			c.Version().VersionID,
			keys.NewHeadstoreColKey(collectionShortID),
		)

		link, headNode, err := coreblock.AddDelta(
			signingCtx,
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

	_, stillExists, err := c.db.collectionRepository.TryGet(ctx, description.CollectionIndex{
		Kind:  description.CollectionID,
		Value: c.def.CollectionID,
	})
	if err != nil {
		return err
	}
	if !stillExists {
		return client.ErrCollectionNotFound
	}

	return nil
}

func appendEncryptionCID(cids []cid.Cid, rawBlock []byte) ([]cid.Cid, error) {
	block, err := coreblock.GetFromBytes(rawBlock)
	if err != nil {
		return nil, err
	}
	if block.Encryption == nil {
		return cids, nil
	}
	return append(cids, block.Encryption.Cid), nil
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
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteDocumentPerm); err != nil {
		return false, err
	}

	ctx = iIdentity.WithContext(ctx, opt.Identity)
	ctx = setContextSigning(ctx, c.db.signingDisabled, opt.EnableSigning)

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}

	defer txn.Discard()

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
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
	if primaryKey.DocShortID == 0 {
		return false, false, nil
	}

	docID, err := c.getDocIDFromPrimaryKey(ctx, primaryKey)
	if err != nil {
		return false, false, err
	}

	canRead, err := c.checkAccessOfDoc(
		ctx,
		acpTypes.DocumentReadPerm,
		docID,
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
		return false, false, NewErrGetDocStatus(err, docID)
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
	return c.getPrimaryKeyFromDocIDString(ctx, docID.String())
}

func (c *collection) getPrimaryKeyFromDocIDString(
	ctx context.Context,
	docID string,
) (keys.PrimaryDataStoreKey, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, c.Version().CollectionID)
	if err != nil {
		return keys.PrimaryDataStoreKey{}, NewErrGetCollectionShortIDForDoc(err, c.Version().CollectionID)
	}

	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, docID)
	if err != nil {
		return keys.PrimaryDataStoreKey{}, err
	}
	if found {
		return keys.PrimaryDataStoreKey{
			CollectionShortID: collectionShortID,
			DocShortID:        docShortID,
		}, nil
	}

	return keys.PrimaryDataStoreKey{
		CollectionShortID: collectionShortID,
	}, nil
}

func (c *collection) getDocIDFromPrimaryKey(
	ctx context.Context,
	primaryKey keys.PrimaryDataStoreKey,
) (string, error) {
	docID, found, err := id.GetDocID(ctx, primaryKey.DocShortID)
	if err != nil {
		return "", err
	}
	if found {
		return docID, nil
	}
	return "", nil
}
