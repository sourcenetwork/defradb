// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/lens"
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db             *DB
	def            client.CollectionDefinition
	indexes        []CollectionIndex
	fetcherFactory func() fetcher.Fetcher
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionVersions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// newCollection returns a pointer to a newly instantiated DB Collection
func (db *DB) newCollection(desc client.CollectionVersion, schema client.SchemaDescription) (*collection, error) {
	col := &collection{
		db:  db,
		def: client.CollectionDefinition{Version: desc, Schema: schema},
	}
	for _, index := range desc.Indexes {
		colIndex, err := NewCollectionIndex(col, index)
		if err != nil {
			return nil, err
		}
		col.indexes = append(col.indexes, colIndex)
	}
	return col, nil
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
		innerFetcher = fetcher.NewDocumentFetcher()
	}

	return lens.NewFetcher(innerFetcher, c.db.LensRegistry())
}

func (db *DB) getCollectionByID(ctx context.Context, id string) (client.Collection, error) {
	txn := datastore.MustGetTxn(ctx)

	col, err := description.GetCollectionByID(ctx, txn, id)
	if err != nil {
		return nil, err
	}

	schema, err := description.GetSchemaVersion(ctx, txn, id)
	if err != nil {
		return nil, err
	}

	collection, err := db.newCollection(col, schema)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// getCollectionByName returns an existing collection within the database.
func (db *DB) getCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	if name == "" {
		return nil, ErrCollectionNameEmpty
	}

	cols, err := db.getCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, corekv.ErrNotFound
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

// getCollections returns all collections and their descriptions matching the given options
// that currently exist within this [Store].
//
// Inactive collections are not returned by default unless a specific schema version ID
// is provided.
func (db *DB) getCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	txn := datastore.MustGetTxn(ctx)

	var cols []client.CollectionVersion
	switch {
	case options.Name.HasValue():
		col, err := description.GetCollectionByName(ctx, txn, options.Name.Value())
		if err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return nil, err
		}
		cols = append(cols, col)

	case options.VersionID.HasValue():
		col, err := description.GetCollectionByID(ctx, txn, options.VersionID.Value())
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)

	case options.CollectionID.HasValue():
		var err error
		cols, err = description.GetCollectionsByCollectionID(ctx, txn, options.CollectionID.Value())
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
		if options.VersionID.HasValue() {
			if col.VersionID != options.VersionID.Value() {
				continue
			}
		}

		// By default, we don't return inactive collections unless a specific version is requested.
		if !options.IncludeInactive.Value() && !col.IsActive && !options.VersionID.HasValue() {
			continue
		}

		schema, err := description.GetSchemaVersion(ctx, txn, col.VersionID)
		if err != nil {
			// If the schema is not found we leave it as empty and carry on. This can happen when
			// a migration is registered before the schema is declared locally.
			if !errors.Is(err, corekv.ErrNotFound) {
				return nil, err
			}
		}

		collection, err := db.newCollection(col, schema)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// getAllActiveDefinitions returns all queryable collection/views and any embedded schema used by them.
func (db *DB) getAllActiveDefinitions(ctx context.Context) ([]client.CollectionDefinition, error) {
	txn := datastore.MustGetTxn(ctx)

	cols, err := description.GetActiveCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	definitions := make([]client.CollectionDefinition, len(cols))
	for i, col := range cols {
		schema, err := description.GetSchemaVersion(ctx, txn, col.VersionID)
		if err != nil {
			return nil, err
		}

		collection, err := db.newCollection(col, schema)
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
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, _, err := ensureContextTxn(ctx, c.db, true)
	if err != nil {
		return nil, err
	}
	return c.getAllDocIDsChan(ctx)
}

func (c *collection) getAllDocIDsChan(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	txn := datastore.MustGetTxn(ctx)

	shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
	if err != nil {
		return nil, err
	}
	prefix := keys.PrimaryDataStoreKey{ // empty path for all keys prefix
		CollectionShortID: shortID,
	}
	iter, err := datastore.DatastoreFrom(txn.Store()).Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix.Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	resCh := make(chan client.DocIDResult)
	go func() {
		defer func() {
			if err := iter.Close(); err != nil {
				log.ErrorContextE(ctx, errFailedtoCloseQueryReqAllIDs, err)
			}
			close(resCh)
			txn.Discard(ctx)
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
				resCh <- client.DocIDResult{
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
				resCh <- client.DocIDResult{
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

// Version returns the client.CollectionVersion.
func (c *collection) Version() client.CollectionVersion {
	return c.Definition().Version
}

// Name returns the collection name.
func (c *collection) Name() string {
	return c.Version().Name
}

// Schema returns the Schema of the collection.
func (c *collection) Schema() client.SchemaDescription {
	return c.Definition().Schema
}

// VersionID returns the VersionID of the collection.
func (c *collection) VersionID() string {
	return c.Version().VersionID
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
	opts ...client.DocCreateOption,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = c.create(ctx, doc, opts)
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
	opts ...client.DocCreateOption,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	for _, doc := range docs {
		err = c.create(ctx, doc, opts)
		if err != nil {
			return err
		}
	}
	return txn.Commit(ctx)
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

func (c *collection) create(
	ctx context.Context,
	doc *client.Document,
	opts []client.DocCreateOption,
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
		txn := datastore.MustGetTxn(ctx)

		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}

		valueKey := keys.DataStoreKey{
			CollectionShortID: shortID,
			DocID:             docID.String(),
			InstanceType:      keys.ValueKey,
		}

		err = datastore.DatastoreFrom(txn.Store()).Set(ctx, valueKey.Bytes(), []byte{base.ObjectMarker})
		if err != nil {
			return err
		}
	}

	ctx = setContextDocEncryption(ctx, opts)

	// write data to DB via MerkleClock/CRDT
	err = c.save(ctx, doc, true)
	if err != nil {
		return err
	}

	err = c.indexNewDoc(ctx, doc)
	if err != nil {
		return err
	}

	return c.registerDocWithACP(ctx, doc.ID().String())
}

func setContextDocEncryption(ctx context.Context, opts []client.DocCreateOption) context.Context {
	createOptions := client.DocCreateOptions{}
	createOptions.Apply(opts)
	if !createOptions.EncryptDoc && len(createOptions.EncryptedFields) == 0 {
		return ctx
	}
	ctx = encryption.SetContextConfigFromParams(ctx, createOptions.EncryptDoc, createOptions.EncryptedFields)
	return ctx
}

// Update an existing document with the new values.
// Any field that needs to be removed or cleared should call doc.Clear(field) before.
// Any field that is nil/empty that hasn't called Clear will be ignored.
func (c *collection) Update(
	ctx context.Context,
	doc *client.Document,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

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

// Save a document into the db.
// Either by creating a new document or by updating an existing one
func (c *collection) Save(
	ctx context.Context,
	doc *client.Document,
	opts ...client.DocCreateOption,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

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
		err = c.create(ctx, doc, opts)
	}
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
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
		if _, exists := c.Schema().GetFieldByName(field); !exists {
			return client.NewErrFieldNotExist(field)
		}
		if strings.HasPrefix(field, "_") {
			return NewErrCanNotEncryptBuiltinField(field)
		}
	}
	return nil
}

// save saves the document state. save MUST not be called outside the `c.create`
// and `c.update` methods as we wrap the acp logic within those methods. Calling
// save elsewhere could cause the omission of acp checks.
func (c *collection) save(
	ctx context.Context,
	doc *client.Document,
	isCreate bool,
) error {
	if err := c.validateEncryptedFields(ctx); err != nil {
		return err
	}

	if !isCreate {
		err := c.updateIndexedDoc(ctx, doc)
		if err != nil {
			return err
		}
	}
	txn := datastore.MustGetTxn(ctx)

	ident := identity.FromContext(ctx)
	if (!ident.HasValue() || !hasPrivateKey(ident.Value())) && c.db.nodeIdentity.HasValue() {
		ctx = identity.WithContext(ctx, c.db.nodeIdentity)
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
			fieldID, err := id.GetShortFieldID(ctx, shortID, k)
			if err != nil {
				return err
			}
			fieldKey := keys.DataStoreKey{
				CollectionShortID: shortID,
				DocID:             primaryKey.DocID,
				FieldID:           strconv.FormatUint(uint64(fieldID), 10),
			}

			fieldDescription, valid := c.Definition().GetFieldByName(k)
			if !valid {
				return client.NewErrFieldNotExist(k)
			}

			// by default the type will have been set to LWW_REGISTER. We need to ensure
			// that it's set to the same as the field description CRDT type.
			val.SetType(fieldDescription.Typ)

			err = c.validateOneToOneLinkDoesntAlreadyExist(
				ctx,
				doc.ID().String(),
				fieldDescription,
				val.Value(),
			)
			if err != nil {
				return err
			}

			merkleCRDT, err := crdt.FieldLevelCRDTWithStore(
				datastore.DatastoreFrom(txn.Store()),
				c.Schema().VersionID,
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

			link, _, err := coreblock.AddDelta(ctx, txn, merkleCRDT, delta)
			if err != nil {
				return err
			}

			links = append(links, coreblock.NewDAGLink(k, link))
		}
	}

	merkleCRDT := crdt.NewDocComposite(
		datastore.DatastoreFrom(txn.Store()),
		c.Schema().VersionID,
		primaryKey.ToDataStoreKey().WithFieldID(core.COMPOSITE_NAMESPACE),
	)

	link, headNode, err := coreblock.AddDelta(ctx, txn, merkleCRDT, merkleCRDT.Delta(), links...)
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
		c.db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
	})

	txn.OnSuccess(func() {
		doc.SetHead(link.Cid)
	})

	if c.def.Version.IsBranchable {
		shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
		if err != nil {
			return err
		}
		collectionCRDT := crdt.NewCollection(
			c.Schema().VersionID,
			keys.NewHeadstoreColKey(shortID),
		)

		link, headNode, err := coreblock.AddDelta(
			ctx,
			txn,
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
			c.db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
		})
	}

	return nil
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

	otherCol, _, err := client.GetDefinitionFromStore(ctx, c.db, c.Definition(), objFieldDescription.Kind)
	if err != nil {
		return err
	}

	otherObjFieldDescription, _ := otherCol.Version.GetFieldByRelation(
		fieldDescription.RelationName,
		c.Name(),
		objFieldDescription.Name,
	)
	if !(otherObjFieldDescription.Kind.HasValue() &&
		otherObjFieldDescription.Kind.Value().IsObject() &&
		!otherObjFieldDescription.Kind.Value().IsArray()) {
		// If the other field is not an object field then this is not a one to one relation and we can continue
		return nil
	}

	filter := fmt.Sprintf(
		`%s: {_ne: "%s"}, %s: {_eq: "%s"}`,
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
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)

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
	return true, txn.Commit(ctx)
}

// Exists checks if a given document exists with supplied DocID.
func (c *collection) Exists(
	ctx context.Context,
	docID client.DocID,
) (bool, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return false, err
	}

	exists, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return false, err
	}
	return exists && !isDeleted, txn.Commit(ctx)
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

	txn := datastore.MustGetTxn(ctx)
	val, err := datastore.DatastoreFrom(txn.Store()).Get(ctx, primaryKey.Bytes())
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
