// Copyright 2022 Democratized Data Foundation
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
	"fmt"
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/lens"
	merklecrdt "github.com/sourcenetwork/defradb/internal/merkle/crdt"
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db             *db
	def            client.CollectionDefinition
	indexes        []CollectionIndex
	fetcherFactory func() fetcher.Fetcher
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionDescriptions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// newCollection returns a pointer to a newly instantiated DB Collection
func (db *db) newCollection(desc client.CollectionDescription, schema client.SchemaDescription) *collection {
	return &collection{
		db:  db,
		def: client.CollectionDefinition{Description: desc, Schema: schema},
	}
}

// newFetcher returns a new fetcher instance for this collection.
// If a fetcherFactory is set, it will be used to create the fetcher.
// It's a very simple factory, but it allows us to inject a mock fetcher
// for testing.
func (c *collection) newFetcher() fetcher.Fetcher {
	var innerFetcher fetcher.Fetcher
	if c.fetcherFactory != nil {
		innerFetcher = c.fetcherFactory()
	} else {
		innerFetcher = new(fetcher.DocumentFetcher)
	}

	return lens.NewFetcher(innerFetcher, c.db.LensRegistry())
}

func (db *db) getCollectionByID(ctx context.Context, id uint32) (client.Collection, error) {
	txn := mustGetContextTxn(ctx)

	col, err := description.GetCollectionByID(ctx, txn, id)
	if err != nil {
		return nil, err
	}

	schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
	if err != nil {
		return nil, err
	}

	collection := db.newCollection(col, schema)

	err = collection.loadIndexes(ctx)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// getCollectionByName returns an existing collection within the database.
func (db *db) getCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	if name == "" {
		return nil, ErrCollectionNameEmpty
	}

	cols, err := db.getCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

// getCollections returns all collections and their descriptions matching the given options
// that currently exist within this [Store].
//
// Inactive collections are not returned by default unless a specific schema version ID
// is provided.
func (db *db) getCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	txn := mustGetContextTxn(ctx)

	var cols []client.CollectionDescription
	switch {
	case options.Name.HasValue():
		col, err := description.GetCollectionByName(ctx, txn, options.Name.Value())
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)

	case options.SchemaVersionID.HasValue():
		var err error
		cols, err = description.GetCollectionsBySchemaVersionID(ctx, txn, options.SchemaVersionID.Value())
		if err != nil {
			return nil, err
		}

	case options.SchemaRoot.HasValue():
		var err error
		cols, err = description.GetCollectionsBySchemaRoot(ctx, txn, options.SchemaRoot.Value())
		if err != nil {
			return nil, err
		}

	default:
		if options.IncludeInactive.HasValue() && options.IncludeInactive.Value() {
			var err error
			cols, err = description.GetCollections(ctx, txn)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			cols, err = description.GetActiveCollections(ctx, txn)
			if err != nil {
				return nil, err
			}
		}
	}

	collections := []client.Collection{}
	for _, col := range cols {
		if options.SchemaVersionID.HasValue() {
			if col.SchemaVersionID != options.SchemaVersionID.Value() {
				continue
			}
		}
		// By default, we don't return inactive collections unless a specific version is requested.
		if !options.IncludeInactive.Value() && !col.Name.HasValue() && !options.SchemaVersionID.HasValue() {
			continue
		}

		schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
		if err != nil {
			// If the schema is not found we leave it as empty and carry on. This can happen when
			// a migration is registered before the schema is declared locally.
			if !errors.Is(err, ds.ErrNotFound) {
				return nil, err
			}
		}

		if options.SchemaRoot.HasValue() {
			if schema.Root != options.SchemaRoot.Value() {
				continue
			}
		}

		collection := db.newCollection(col, schema)
		collections = append(collections, collection)

		err = collection.loadIndexes(ctx)
		if err != nil {
			return nil, err
		}
	}

	return collections, nil
}

// getAllActiveDefinitions returns all queryable collection/views and any embedded schema used by them.
func (db *db) getAllActiveDefinitions(ctx context.Context) ([]client.CollectionDefinition, error) {
	txn := mustGetContextTxn(ctx)

	cols, err := description.GetActiveCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	definitions := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
		if err != nil {
			return nil, err
		}

		collection := db.newCollection(col, schema)

		err = collection.loadIndexes(ctx)
		if err != nil {
			return nil, err
		}

		definitions[i] = collection.Definition()
	}

	schemas, err := description.GetCollectionlessSchemas(ctx, txn)
	if err != nil {
		return nil, err
	}

	for _, schema := range schemas {
		definitions = append(
			definitions,
			client.CollectionDefinition{
				Schema: schema,
			},
		)
	}

	return definitions, nil
}

// GetAllDocIDs returns all the document IDs that exist in the collection.
//
// @todo: We probably need a lock on the collection for this kind of op since
// it hits every key and will cause Tx conflicts for concurrent Txs
func (c *collection) GetAllDocIDs(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	ctx, _, err := ensureContextTxn(ctx, c.db, true)
	if err != nil {
		return nil, err
	}
	return c.getAllDocIDsChan(ctx)
}

func (c *collection) getAllDocIDsChan(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	txn := mustGetContextTxn(ctx)
	prefix := core.PrimaryDataStoreKey{ // empty path for all keys prefix
		CollectionRootID: c.Description().RootID,
	}
	q, err := txn.Datastore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	resCh := make(chan client.DocIDResult)
	go func() {
		defer func() {
			if err := q.Close(); err != nil {
				log.ErrorContextE(ctx, errFailedtoCloseQueryReqAllIDs, err)
			}
			close(resCh)
			txn.Discard(ctx)
		}()
		for res := range q.Next() {
			// check for Done on context first
			select {
			case <-ctx.Done():
				// we've been cancelled! ;)
				return
			default:
				// noop, just continue on the with the for loop
			}
			if res.Error != nil {
				resCh <- client.DocIDResult{
					Err: res.Error,
				}
				return
			}

			rawDocID := ds.NewKey(res.Key).BaseNamespace()
			docID, err := client.NewDocIDFromString(rawDocID)
			if err != nil {
				resCh <- client.DocIDResult{
					Err: err,
				}
				return
			}

			canRead, err := c.checkAccessOfDocWithACP(
				ctx,
				acp.ReadPermission,
				docID.String(),
			)

			if err != nil {
				resCh <- client.DocIDResult{
					Err: err,
				}
				return
			}

			if canRead {
				resCh <- client.DocIDResult{
					ID: docID,
				}
			}
		}
	}()

	return resCh, nil
}

// Description returns the client.CollectionDescription.
func (c *collection) Description() client.CollectionDescription {
	return c.Definition().Description
}

// Name returns the collection name.
func (c *collection) Name() immutable.Option[string] {
	return c.Description().Name
}

// Schema returns the Schema of the collection.
func (c *collection) Schema() client.SchemaDescription {
	return c.Definition().Schema
}

// ID returns the ID of the collection.
func (c *collection) ID() uint32 {
	return c.Description().ID
}

func (c *collection) SchemaRoot() string {
	return c.Schema().Root
}

func (c *collection) Definition() client.CollectionDefinition {
	return c.def
}

// Create a new document.
// Will verify the DocID/CID to ensure that the new document is correctly formatted.
func (c *collection) Create(
	ctx context.Context,
	doc *client.Document,
) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = c.create(ctx, doc)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// CreateMany creates a collection of documents at once.
// Will verify the DocID/CID to ensure that the new documents are correctly formatted.
func (c *collection) CreateMany(
	ctx context.Context,
	docs []*client.Document,
) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	for _, doc := range docs {
		err = c.create(ctx, doc)
		if err != nil {
			return err
		}
	}
	return txn.Commit(ctx)
}

func (c *collection) getDocIDAndPrimaryKeyFromDoc(
	doc *client.Document,
) (client.DocID, core.PrimaryDataStoreKey, error) {
	docID, err := doc.GenerateDocID()
	if err != nil {
		return client.DocID{}, core.PrimaryDataStoreKey{}, err
	}

	primaryKey := c.getPrimaryKeyFromDocID(docID)
	if primaryKey.DocID != doc.ID().String() {
		return client.DocID{}, core.PrimaryDataStoreKey{},
			NewErrDocVerification(doc.ID().String(), primaryKey.DocID)
	}
	return docID, primaryKey, nil
}

func (c *collection) create(
	ctx context.Context,
	doc *client.Document,
) error {
	docID, primaryKey, err := c.getDocIDAndPrimaryKeyFromDoc(doc)
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
		txn := mustGetContextTxn(ctx)
		valueKey := c.getDataStoreKeyFromDocID(docID)
		err = txn.Datastore().Put(ctx, valueKey.ToDS(), []byte{base.ObjectMarker})
		if err != nil {
			return err
		}
	}

	// write data to DB via MerkleClock/CRDT
	_, err = c.save(ctx, doc, true)
	if err != nil {
		return err
	}

	err = c.indexNewDoc(ctx, doc)
	if err != nil {
		return err
	}

	return c.registerDocWithACP(ctx, doc.ID().String())
}

// Update an existing document with the new values.
// Any field that needs to be removed or cleared should call doc.Clear(field) before.
// Any field that is nil/empty that hasn't called Clear will be ignored.
func (c *collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	primaryKey := c.getPrimaryKeyFromDocID(doc.ID())
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

	return txn.Commit(ctx)
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
		acp.WritePermission,
		doc.ID().String(),
	)
	if err != nil {
		return err
	}
	if !canUpdate {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}

	_, err = c.save(ctx, doc, false)
	if err != nil {
		return err
	}
	return nil
}

// Save a document into the db.
// Either by creating a new document or by updating an existing one
func (c *collection) Save(
	ctx context.Context,
	doc *client.Document,
) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// Check if document already exists with primary DS key.
	primaryKey := c.getPrimaryKeyFromDocID(doc.ID())
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
		err = c.create(ctx, doc)
	}
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// save saves the document state. save MUST not be called outside the `c.create`
// and `c.update` methods as we wrap the acp logic within those methods. Calling
// save elsewhere could cause the omission of acp checks.
func (c *collection) save(
	ctx context.Context,
	doc *client.Document,
	isCreate bool,
) (cid.Cid, error) {
	if !isCreate {
		err := c.updateIndexedDoc(ctx, doc)
		if err != nil {
			return cid.Undef, err
		}
	}
	txn := mustGetContextTxn(ctx)

	// NOTE: We delay the final Clean() call until we know
	// the commit on the transaction is successful. If we didn't
	// wait, and just did it here, then *if* the commit fails down
	// the line, then we have no way to roll back the state
	// side-effect on the document func called here.
	txn.OnSuccess(func() {
		doc.Clean()
	})

	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instantiate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	primaryKey := c.getPrimaryKeyFromDocID(doc.ID())
	links := make([]coreblock.DAGLink, 0)
	for k, v := range doc.Fields() {
		val, err := doc.GetValueWithField(v)
		if err != nil {
			return cid.Undef, err
		}

		if val.IsDirty() {
			fieldKey, fieldExists := c.tryGetFieldKey(primaryKey, k)

			if !fieldExists {
				return cid.Undef, client.NewErrFieldNotExist(k)
			}

			fieldDescription, valid := c.Definition().GetFieldByName(k)
			if !valid {
				return cid.Undef, client.NewErrFieldNotExist(k)
			}

			// by default the type will have been set to LWW_REGISTER. We need to ensure
			// that it's set to the same as the field description CRDT type.
			val.SetType(fieldDescription.Typ)

			relationFieldDescription, isSecondaryRelationID := c.isSecondaryIDField(fieldDescription)
			if isSecondaryRelationID {
				primaryId := val.Value().(string)

				err = c.patchPrimaryDoc(
					ctx,
					c.Name().Value(),
					relationFieldDescription,
					primaryKey.DocID,
					primaryId,
				)
				if err != nil {
					return cid.Undef, err
				}

				// If this field was a secondary relation ID the related document will have been
				// updated instead and we should discard this value
				continue
			}

			err = c.validateOneToOneLinkDoesntAlreadyExist(
				ctx,
				doc.ID().String(),
				fieldDescription,
				val.Value(),
			)
			if err != nil {
				return cid.Undef, err
			}

			merkleCRDT, err := merklecrdt.InstanceWithStore(
				txn,
				core.NewCollectionSchemaVersionKey(c.Schema().VersionID, c.ID()),
				val.Type(),
				fieldDescription.Kind,
				fieldKey,
				fieldDescription.Name,
			)
			if err != nil {
				return cid.Undef, err
			}

			link, _, err := merkleCRDT.Save(ctx, val)
			if err != nil {
				return cid.Undef, err
			}

			links = append(links, coreblock.NewDAGLink(k, link))
		}
	}

	link, headNode, err := c.saveCompositeToMerkleCRDT(
		ctx,
		primaryKey.ToDataStoreKey(),
		links,
		client.Active,
	)
	if err != nil {
		return cid.Undef, err
	}

	// publish an update event when the txn succeeds
	updateEvent := events.UpdateEvent{
		DocID:      doc.ID().String(),
		Cid:        link.Cid,
		SchemaRoot: c.Schema().Root,
		Block:      headNode,
		IsCreate:   isCreate,
	}
	txn.OnSuccess(func() {
		c.db.sysEventBus.Publish(events.NewMessage(events.UpdateEventName, updateEvent))
	})

	txn.OnSuccess(func() {
		doc.SetHead(link.Cid)
	})

	return link.Cid, nil
}

func (c *collection) validateOneToOneLinkDoesntAlreadyExist(
	ctx context.Context,
	docID string,
	fieldDescription client.FieldDefinition,
	value any,
) error {
	if fieldDescription.Kind != client.FieldKind_DocID {
		return nil
	}

	if value == nil {
		return nil
	}

	objFieldDescription, ok := c.Definition().GetFieldByName(
		strings.TrimSuffix(fieldDescription.Name, request.RelatedObjectID),
	)
	if !ok {
		return client.NewErrFieldNotExist(strings.TrimSuffix(fieldDescription.Name, request.RelatedObjectID))
	}
	if !(objFieldDescription.Kind.IsObject() && !objFieldDescription.Kind.IsArray()) {
		return nil
	}

	otherCol, err := c.db.getCollectionByName(ctx, objFieldDescription.Kind.Underlying())
	if err != nil {
		return err
	}
	otherObjFieldDescription, _ := otherCol.Description().GetFieldByRelation(
		fieldDescription.RelationName,
		c.Name().Value(),
		objFieldDescription.Name,
	)
	if !(otherObjFieldDescription.Kind.HasValue() &&
		otherObjFieldDescription.Kind.Value().IsObject() &&
		!otherObjFieldDescription.Kind.Value().IsArray()) {
		// If the other field is not an object field then this is not a one to one relation and we can continue
		return nil
	}

	filter := fmt.Sprintf(
		`{_and: [{%s: {_ne: "%s"}}, {%s: {_eq: "%s"}}]}`,
		request.DocIDFieldName,
		docID,
		fieldDescription.Name,
		value,
	)
	selectionPlan, err := c.makeSelectionPlan(ctx, filter)
	if err != nil {
		return err
	}

	err = selectionPlan.Init()
	if err != nil {
		closeErr := selectionPlan.Close()
		if closeErr != nil {
			return errors.Wrap(err.Error(), closeErr)
		}
		return err
	}

	if err = selectionPlan.Start(); err != nil {
		closeErr := selectionPlan.Close()
		if closeErr != nil {
			return errors.Wrap(err.Error(), closeErr)
		}
		return err
	}

	alreadyLinked, err := selectionPlan.Next()
	if err != nil {
		closeErr := selectionPlan.Close()
		if closeErr != nil {
			return errors.Wrap(err.Error(), closeErr)
		}
		return err
	}

	if alreadyLinked {
		existingDocument := selectionPlan.Value()
		err := selectionPlan.Close()
		if err != nil {
			return err
		}
		return NewErrOneOneAlreadyLinked(docID, existingDocument.GetID(), objFieldDescription.RelationName)
	}

	err = selectionPlan.Close()
	if err != nil {
		return err
	}

	return nil
}

// Delete will attempt to delete a document by docID and return true if a deletion is successful,
// otherwise will return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given DocID. This includes data, block, and head storage.
func (c *collection) Delete(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)

	primaryKey := c.getPrimaryKeyFromDocID(docID)

	err = c.applyDelete(ctx, primaryKey)
	if err != nil {
		return false, err
	}
	return true, txn.Commit(ctx)
}

// Exists checks if a given document exists with supplied DocID.
func (c *collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)

	primaryKey := c.getPrimaryKeyFromDocID(docID)
	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return false, err
	}
	return exists && !isDeleted, txn.Commit(ctx)
}

// check if a document exists with the given primary key
func (c *collection) exists(
	ctx context.Context,
	primaryKey core.PrimaryDataStoreKey,
) (exists bool, isDeleted bool, err error) {
	canRead, err := c.checkAccessOfDocWithACP(
		ctx,
		acp.ReadPermission,
		primaryKey.DocID,
	)
	if err != nil {
		return false, false, err
	} else if !canRead {
		return false, false, nil
	}

	txn := mustGetContextTxn(ctx)
	val, err := txn.Datastore().Get(ctx, primaryKey.ToDS())
	if err != nil && errors.Is(err, ds.ErrNotFound) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}
	if bytes.Equal(val, []byte{base.DeletedObjectMarker}) {
		return true, true, nil
	}

	return true, false, nil
}

// saveCompositeToMerkleCRDT saves the composite to the merkle CRDT.
// It returns the CID of the block and the encoded block.
// saveCompositeToMerkleCRDT MUST not be called outside the `c.save`
// and `c.applyDelete` methods as we wrap the acp logic around those methods.
// Calling it elsewhere could cause the omission of acp checks.
func (c *collection) saveCompositeToMerkleCRDT(
	ctx context.Context,
	dsKey core.DataStoreKey,
	links []coreblock.DAGLink,
	status client.DocumentStatus,
) (cidlink.Link, []byte, error) {
	txn := mustGetContextTxn(ctx)
	dsKey = dsKey.WithFieldId(core.COMPOSITE_NAMESPACE)
	merkleCRDT := merklecrdt.NewMerkleCompositeDAG(
		txn,
		core.NewCollectionSchemaVersionKey(c.Schema().VersionID, c.ID()),
		dsKey,
		"",
	)

	if status.IsDeleted() {
		return merkleCRDT.Delete(ctx, links)
	}

	return merkleCRDT.Save(ctx, links)
}

func (c *collection) getPrimaryKeyFromDocID(docID client.DocID) core.PrimaryDataStoreKey {
	return core.PrimaryDataStoreKey{
		CollectionRootID: c.Description().RootID,
		DocID:            docID.String(),
	}
}

func (c *collection) getDataStoreKeyFromDocID(docID client.DocID) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionRootID: c.Description().RootID,
		DocID:            docID.String(),
		InstanceType:     core.ValueKey,
	}
}

func (c *collection) tryGetFieldKey(primaryKey core.PrimaryDataStoreKey, fieldName string) (core.DataStoreKey, bool) {
	fieldId, hasField := c.tryGetFieldID(fieldName)
	if !hasField {
		return core.DataStoreKey{}, false
	}

	return core.DataStoreKey{
		CollectionRootID: c.Description().RootID,
		DocID:            primaryKey.DocID,
		FieldId:          strconv.FormatUint(uint64(fieldId), 10),
	}, true
}

// tryGetFieldID returns the FieldID of the given fieldName.
// Will return false if the field is not found.
func (c *collection) tryGetFieldID(fieldName string) (uint32, bool) {
	for _, field := range c.Definition().GetFields() {
		if field.Name == fieldName {
			if field.Kind.IsObject() || field.Kind.IsObjectArray() {
				// We do not wish to match navigational properties, only
				// fields directly on the collection.
				return uint32(0), false
			}
			return uint32(field.ID), true
		}
	}

	return uint32(0), false
}
