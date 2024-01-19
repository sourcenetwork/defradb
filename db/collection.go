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
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/description"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/lens"
	merklecrdt "github.com/sourcenetwork/defradb/merkle/crdt"
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db *db

	// txn represents any externally provided [datastore.Txn] for which any
	// operation on this [collection] instance should be scoped to.
	//
	// If this has no value, operations requiring a transaction should use an
	// implicit internally managed transaction, which only lives for duration
	// of the operation in question.
	txn immutable.Option[datastore.Txn]

	def client.CollectionDefinition

	indexes        []CollectionIndex
	fetcherFactory func() fetcher.Fetcher
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionDescriptions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// NewCollection returns a pointer to a newly instanciated DB Collection
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

// createCollection creates a collection and saves it to the database in its system store.
// Note: Collection.ID is an autoincrementing value that is generated by the database.
func (db *db) createCollection(
	ctx context.Context,
	txn datastore.Txn,
	def client.CollectionDefinition,
) (client.Collection, error) {
	schema := def.Schema
	desc := def.Description

	if desc.Name.HasValue() {
		exists, err := description.HasCollectionByName(ctx, txn, desc.Name.Value())
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCollectionAlreadyExists
		}
	}

	colSeq, err := db.getSequence(ctx, txn, core.COLLECTION)
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next(ctx, txn)
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)

	schema, err = description.CreateSchemaVersion(ctx, txn, schema)
	if err != nil {
		return nil, err
	}
	desc.SchemaVersionID = schema.VersionID

	desc, err = description.SaveCollection(ctx, txn, desc)
	if err != nil {
		return nil, err
	}

	col := db.newCollection(desc, schema)
	for _, index := range desc.Indexes {
		if _, err := col.createIndex(ctx, txn, index); err != nil {
			return nil, err
		}
	}

	return db.getCollectionByID(ctx, txn, desc.ID)
}

// updateSchema updates the persisted schema description matching the name of the given
// description, to the values in the given description.
//
// It will validate the given description using [validateUpdateSchema] before updating it.
//
// The schema (including the schema version ID) will only be updated if any changes have actually
// been made, if the given description matches the current persisted description then no changes will be
// applied.
func (db *db) updateSchema(
	ctx context.Context,
	txn datastore.Txn,
	existingSchemaByName map[string]client.SchemaDescription,
	proposedDescriptionsByName map[string]client.SchemaDescription,
	schema client.SchemaDescription,
	setAsDefaultVersion bool,
) error {
	hasChanged, err := db.validateUpdateSchema(
		ctx,
		txn,
		existingSchemaByName,
		proposedDescriptionsByName,
		schema,
	)
	if err != nil {
		return err
	}

	if !hasChanged {
		return nil
	}

	for _, field := range schema.Fields {
		if field.RelationType.IsSet(client.Relation_Type_ONE) {
			idFieldName := field.Name + "_id"
			if _, ok := schema.GetField(idFieldName); !ok {
				schema.Fields = append(schema.Fields, client.FieldDescription{
					Name:         idFieldName,
					Kind:         client.FieldKind_DocID,
					RelationType: client.Relation_Type_INTERNAL_ID,
					RelationName: field.RelationName,
				})
			}
		}
	}

	for i, field := range schema.Fields {
		if field.Typ == client.NONE_CRDT {
			// If no CRDT Type has been provided, default to LWW_REGISTER.
			field.Typ = client.LWW_REGISTER
			schema.Fields[i] = field
		}
	}

	previousVersionID := schema.VersionID
	schema, err = description.CreateSchemaVersion(ctx, txn, schema)
	if err != nil {
		return err
	}

	if setAsDefaultVersion {
		cols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, previousVersionID)
		if err != nil {
			return err
		}

		for _, col := range cols {
			if !col.Name.HasValue() {
				// Nameless collections cannot be made default as they cannot be queried without a name.
				// Note: The `setAsDefaultVersion` block will need a re-write when collections become immutable
				// and the schema version stuff gets tracked by [CollectionDescription.Sources] instead.
				continue
			}

			col.SchemaVersionID = schema.VersionID

			col, err = description.SaveCollection(ctx, txn, col)
			if err != nil {
				return err
			}

			err = db.setDefaultSchemaVersionExplicit(ctx, txn, col.Name.Value(), schema.VersionID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateUpdateSchema validates that the given schema description is a valid update.
//
// Will return true if the given description differs from the current persisted state of the
// schema. Will return an error if it fails validation.
func (db *db) validateUpdateSchema(
	ctx context.Context,
	txn datastore.Txn,
	existingDescriptionsByName map[string]client.SchemaDescription,
	proposedDescriptionsByName map[string]client.SchemaDescription,
	proposedDesc client.SchemaDescription,
) (bool, error) {
	if proposedDesc.Name == "" {
		return false, ErrSchemaNameEmpty
	}

	existingDesc, collectionExists := existingDescriptionsByName[proposedDesc.Name]
	if !collectionExists {
		return false, NewErrAddCollectionWithPatch(proposedDesc.Name)
	}

	if proposedDesc.Root != existingDesc.Root {
		return false, NewErrSchemaRootDoesntMatch(
			proposedDesc.Name,
			existingDesc.Root,
			proposedDesc.Root,
		)
	}

	if proposedDesc.Name != existingDesc.Name {
		// There is actually little reason to not support this atm besides controlling the surface area
		// of the new feature.  Changing this should not break anything, but it should be tested first.
		return false, NewErrCannotModifySchemaName(existingDesc.Name, proposedDesc.Name)
	}

	if proposedDesc.VersionID != "" && proposedDesc.VersionID != existingDesc.VersionID {
		// If users specify this it will be overwritten, an error is prefered to quietly ignoring it.
		return false, ErrCannotSetVersionID
	}

	hasChangedFields, err := validateUpdateSchemaFields(proposedDescriptionsByName, existingDesc, proposedDesc)
	if err != nil {
		return hasChangedFields, err
	}

	return hasChangedFields, err
}

func validateUpdateSchemaFields(
	descriptionsByName map[string]client.SchemaDescription,
	existingDesc client.SchemaDescription,
	proposedDesc client.SchemaDescription,
) (bool, error) {
	hasChanged := false
	existingFieldsByID := map[client.FieldID]client.FieldDescription{}
	existingFieldIndexesByName := map[string]int{}
	for i, field := range existingDesc.Fields {
		existingFieldIndexesByName[field.Name] = i
		existingFieldsByID[field.ID] = field
	}

	newFieldNames := map[string]struct{}{}
	newFieldIds := map[client.FieldID]struct{}{}
	for proposedIndex, proposedField := range proposedDesc.Fields {
		var existingField client.FieldDescription
		var fieldAlreadyExists bool
		if proposedField.ID != client.FieldID(0) ||
			proposedField.Name == request.DocIDFieldName {
			existingField, fieldAlreadyExists = existingFieldsByID[proposedField.ID]
		}

		if proposedField.ID != client.FieldID(0) && !fieldAlreadyExists {
			return false, NewErrCannotSetFieldID(proposedField.Name, proposedField.ID)
		}

		// If the field is new, then the collection has changed
		hasChanged = hasChanged || !fieldAlreadyExists

		if !fieldAlreadyExists && (proposedField.Kind == client.FieldKind_FOREIGN_OBJECT ||
			proposedField.Kind == client.FieldKind_FOREIGN_OBJECT_ARRAY) {
			if proposedField.Schema == "" {
				return false, NewErrRelationalFieldMissingSchema(proposedField.Name, proposedField.Kind)
			}

			relatedDesc, relatedDescFound := descriptionsByName[proposedField.Schema]

			if !relatedDescFound {
				return false, NewErrSchemaNotFound(proposedField.Name, proposedField.Schema)
			}

			if proposedField.Kind == client.FieldKind_FOREIGN_OBJECT {
				if !proposedField.RelationType.IsSet(client.Relation_Type_ONE) ||
					!(proposedField.RelationType.IsSet(client.Relation_Type_ONEONE) ||
						proposedField.RelationType.IsSet(client.Relation_Type_ONEMANY)) {
					return false, NewErrRelationalFieldInvalidRelationType(
						proposedField.Name,
						fmt.Sprintf(
							"%v and %v or %v, with optionally %v",
							client.Relation_Type_ONE,
							client.Relation_Type_ONEONE,
							client.Relation_Type_ONEMANY,
							client.Relation_Type_Primary,
						),
						proposedField.RelationType,
					)
				}
			}

			if proposedField.Kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
				if !proposedField.RelationType.IsSet(client.Relation_Type_MANY) ||
					!proposedField.RelationType.IsSet(client.Relation_Type_ONEMANY) {
					return false, NewErrRelationalFieldInvalidRelationType(
						proposedField.Name,
						client.Relation_Type_MANY|client.Relation_Type_ONEMANY,
						proposedField.RelationType,
					)
				}
			}

			if proposedField.RelationName == "" {
				return false, NewErrRelationalFieldMissingRelationName(proposedField.Name)
			}

			if proposedField.RelationType.IsSet(client.Relation_Type_Primary) {
				if proposedField.Kind == client.FieldKind_FOREIGN_OBJECT_ARRAY {
					return false, NewErrPrimarySideOnMany(proposedField.Name)
				}
			}

			if proposedField.Kind == client.FieldKind_FOREIGN_OBJECT {
				idFieldName := proposedField.Name + request.RelatedObjectID
				idField, idFieldFound := proposedDesc.GetField(idFieldName)
				if idFieldFound {
					if idField.Kind != client.FieldKind_DocID {
						return false, NewErrRelationalFieldIDInvalidType(idField.Name, client.FieldKind_DocID, idField.Kind)
					}

					if idField.RelationType != client.Relation_Type_INTERNAL_ID {
						return false, NewErrRelationalFieldInvalidRelationType(
							idField.Name,
							client.Relation_Type_INTERNAL_ID,
							idField.RelationType,
						)
					}

					if idField.RelationName == "" {
						return false, NewErrRelationalFieldMissingRelationName(idField.Name)
					}
				}
			}

			var relatedFieldFound bool
			var relatedField client.FieldDescription
			for _, field := range relatedDesc.Fields {
				if field.RelationName == proposedField.RelationName &&
					!field.RelationType.IsSet(client.Relation_Type_INTERNAL_ID) &&
					!(relatedDesc.Name == proposedDesc.Name && field.Name == proposedField.Name) {
					relatedFieldFound = true
					relatedField = field
					break
				}
			}

			if !relatedFieldFound {
				return false, client.NewErrRelationOneSided(proposedField.Name, proposedField.Schema)
			}

			if !(proposedField.RelationType.IsSet(client.Relation_Type_Primary) ||
				relatedField.RelationType.IsSet(client.Relation_Type_Primary)) {
				return false, NewErrPrimarySideNotDefined(proposedField.RelationName)
			}

			if proposedField.RelationType.IsSet(client.Relation_Type_Primary) &&
				relatedField.RelationType.IsSet(client.Relation_Type_Primary) {
				return false, NewErrBothSidesPrimary(proposedField.RelationName)
			}

			if proposedField.RelationType.IsSet(client.Relation_Type_ONEONE) &&
				relatedField.Kind != client.FieldKind_FOREIGN_OBJECT {
				return false, NewErrRelatedFieldKindMismatch(
					proposedField.RelationName,
					client.FieldKind_FOREIGN_OBJECT,
					relatedField.Kind,
				)
			}

			if proposedField.RelationType.IsSet(client.Relation_Type_ONEMANY) &&
				proposedField.Kind == client.FieldKind_FOREIGN_OBJECT &&
				relatedField.Kind != client.FieldKind_FOREIGN_OBJECT_ARRAY {
				return false, NewErrRelatedFieldKindMismatch(
					proposedField.RelationName,
					client.FieldKind_FOREIGN_OBJECT_ARRAY,
					relatedField.Kind,
				)
			}

			if proposedField.RelationType.IsSet(client.Relation_Type_ONEONE) &&
				!relatedField.RelationType.IsSet(client.Relation_Type_ONEONE) {
				return false, NewErrRelatedFieldRelationTypeMismatch(
					proposedField.RelationName,
					client.Relation_Type_ONEONE,
					relatedField.RelationType,
				)
			}
		}

		if _, isDuplicate := newFieldNames[proposedField.Name]; isDuplicate {
			return false, NewErrDuplicateField(proposedField.Name)
		}

		if fieldAlreadyExists && proposedField != existingField {
			return false, NewErrCannotMutateField(proposedField.ID, proposedField.Name)
		}

		if existingIndex := existingFieldIndexesByName[proposedField.Name]; fieldAlreadyExists &&
			proposedIndex != existingIndex {
			return false, NewErrCannotMoveField(proposedField.Name, proposedIndex, existingIndex)
		}

		if !proposedField.Typ.IsSupportedFieldCType() {
			return false, client.NewErrInvalidCRDTType(proposedField.Name, proposedField.Typ.String())
		}

		if !proposedField.Typ.IsCompatibleWith(proposedField.Kind) {
			return false, client.NewErrCRDTKindMismatch(proposedField.Typ.String(), proposedField.Kind.String())
		}

		newFieldNames[proposedField.Name] = struct{}{}
		newFieldIds[proposedField.ID] = struct{}{}
	}

	for _, field := range existingDesc.Fields {
		if _, stillExists := newFieldIds[field.ID]; !stillExists {
			return false, NewErrCannotDeleteField(field.Name, field.ID)
		}
	}
	return hasChanged, nil
}

func (db *db) setDefaultSchemaVersion(
	ctx context.Context,
	txn datastore.Txn,
	schemaVersionID string,
) error {
	if schemaVersionID == "" {
		return ErrSchemaVersionIDEmpty
	}

	schema, err := description.GetSchemaVersion(ctx, txn, schemaVersionID)
	if err != nil {
		return err
	}

	colDescs, err := description.GetCollectionsBySchemaRoot(ctx, txn, schema.Root)
	if err != nil {
		return err
	}

	for _, col := range colDescs {
		col.SchemaVersionID = schemaVersionID
		col, err = description.SaveCollection(ctx, txn, col)
		if err != nil {
			return err
		}
	}

	return db.loadSchema(ctx, txn)
}

func (db *db) setDefaultSchemaVersionExplicit(
	ctx context.Context,
	txn datastore.Txn,
	collectionName string,
	schemaVersionID string,
) error {
	if schemaVersionID == "" {
		return ErrSchemaVersionIDEmpty
	}

	col, err := description.GetCollectionByName(ctx, txn, collectionName)
	if err != nil {
		return err
	}

	col.SchemaVersionID = schemaVersionID

	_, err = description.SaveCollection(ctx, txn, col)
	return err
}

// getCollectionsByVersionId returns the [*collection]s at the given [schemaVersionId] version.
//
// Will return an error if the given key is empty, or if none are found.
func (db *db) getCollectionsByVersionID(
	ctx context.Context,
	txn datastore.Txn,
	schemaVersionId string,
) ([]*collection, error) {
	cols, err := description.GetCollectionsBySchemaVersionID(ctx, txn, schemaVersionId)
	if err != nil {
		return nil, err
	}

	collections := make([]*collection, len(cols))
	for i, col := range cols {
		schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
		if err != nil {
			return nil, err
		}

		collections[i] = db.newCollection(col, schema)

		err = collections[i].loadIndexes(ctx, txn)
		if err != nil {
			return nil, err
		}
	}

	return collections, nil
}

func (db *db) getCollectionByID(ctx context.Context, txn datastore.Txn, id uint32) (client.Collection, error) {
	col, err := description.GetCollectionByID(ctx, txn, id)
	if err != nil {
		return nil, err
	}

	schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
	if err != nil {
		return nil, err
	}

	collection := db.newCollection(col, schema)
	err = collection.loadIndexes(ctx, txn)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// getCollectionByName returns an existing collection within the database.
func (db *db) getCollectionByName(ctx context.Context, txn datastore.Txn, name string) (client.Collection, error) {
	if name == "" {
		return nil, ErrCollectionNameEmpty
	}

	col, err := description.GetCollectionByName(ctx, txn, name)
	if err != nil {
		return nil, err
	}

	schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
	if err != nil {
		return nil, err
	}

	collection := db.newCollection(col, schema)
	err = collection.loadIndexes(ctx, txn)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

// getCollectionsBySchemaRoot returns all existing collections using the schema root.
func (db *db) getCollectionsBySchemaRoot(
	ctx context.Context,
	txn datastore.Txn,
	schemaRoot string,
) ([]client.Collection, error) {
	if schemaRoot == "" {
		return nil, ErrSchemaRootEmpty
	}

	cols, err := description.GetCollectionsBySchemaRoot(ctx, txn, schemaRoot)
	if err != nil {
		return nil, err
	}

	collections := make([]client.Collection, len(cols))
	for i, col := range cols {
		schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
		if err != nil {
			return nil, err
		}

		collection := db.newCollection(col, schema)
		collections[i] = collection

		err = collection.loadIndexes(ctx, txn)
		if err != nil {
			return nil, err
		}
	}

	return collections, nil
}

// getAllCollections gets all the currently defined collections.
func (db *db) getAllCollections(ctx context.Context, txn datastore.Txn) ([]client.Collection, error) {
	cols, err := description.GetCollections(ctx, txn)
	if err != nil {
		return nil, err
	}

	collections := make([]client.Collection, len(cols))
	for i, col := range cols {
		schema, err := description.GetSchemaVersion(ctx, txn, col.SchemaVersionID)
		if err != nil {
			return nil, err
		}

		collection := db.newCollection(col, schema)
		collections[i] = collection

		err = collection.loadIndexes(ctx, txn)
		if err != nil {
			return nil, err
		}
	}

	return collections, nil
}

// getAllActiveDefinitions returns all queryable collection/views and any embedded schema used by them.
func (db *db) getAllActiveDefinitions(ctx context.Context, txn datastore.Txn) ([]client.CollectionDefinition, error) {
	cols, err := description.GetCollections(ctx, txn)
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

		err = collection.loadIndexes(ctx, txn)
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
func (c *collection) GetAllDocIDs(ctx context.Context) (<-chan client.DocIDResult, error) {
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return nil, err
	}

	return c.getAllDocIDsChan(ctx, txn)
}

func (c *collection) getAllDocIDsChan(
	ctx context.Context,
	txn datastore.Txn,
) (<-chan client.DocIDResult, error) {
	prefix := core.PrimaryDataStoreKey{ // empty path for all keys prefix
		CollectionId: fmt.Sprint(c.ID()),
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
				log.ErrorE(ctx, errFailedtoCloseQueryReqAllIDs, err)
			}
			close(resCh)
			c.discardImplicitTxn(ctx, txn)
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
					Err: res.Error,
				}
				return
			}
			resCh <- client.DocIDResult{
				ID: docID,
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

// WithTxn returns a new instance of the collection, with a transaction
// handle instead of a raw DB handle.
func (c *collection) WithTxn(txn datastore.Txn) client.Collection {
	return &collection{
		db:             c.db,
		txn:            immutable.Some(txn),
		def:            c.def,
		indexes:        c.indexes,
		fetcherFactory: c.fetcherFactory,
	}
}

// Create a new document.
// Will verify the DocID/CID to ensure that the new document is correctly formatted.
func (c *collection) Create(ctx context.Context, doc *client.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	err = c.create(ctx, txn, doc)
	if err != nil {
		return err
	}
	return c.commitImplicitTxn(ctx, txn)
}

// CreateMany creates a collection of documents at once.
// Will verify the DocID/CID to ensure that the new documents are correctly formatted.
func (c *collection) CreateMany(ctx context.Context, docs []*client.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	for _, doc := range docs {
		err = c.create(ctx, txn, doc)
		if err != nil {
			return err
		}
	}
	return c.commitImplicitTxn(ctx, txn)
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

func (c *collection) create(ctx context.Context, txn datastore.Txn, doc *client.Document) error {
	docID, primaryKey, err := c.getDocIDAndPrimaryKeyFromDoc(doc)
	if err != nil {
		return err
	}

	// check if doc already exists
	exists, isDeleted, err := c.exists(ctx, txn, primaryKey)
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
		valueKey := c.getDataStoreKeyFromDocID(docID)
		err = txn.Datastore().Put(ctx, valueKey.ToDS(), []byte{base.ObjectMarker})
		if err != nil {
			return err
		}
	}

	// write data to DB via MerkleClock/CRDT
	_, err = c.save(ctx, txn, doc, true)
	if err != nil {
		return err
	}

	return c.indexNewDoc(ctx, txn, doc)
}

// Update an existing document with the new values.
// Any field that needs to be removed or cleared should call doc.Clear(field) before.
// Any field that is nil/empty that hasn't called Clear will be ignored.
func (c *collection) Update(ctx context.Context, doc *client.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	primaryKey := c.getPrimaryKeyFromDocID(doc.ID())
	exists, isDeleted, err := c.exists(ctx, txn, primaryKey)
	if err != nil {
		return err
	}
	if !exists {
		return client.ErrDocumentNotFound
	}
	if isDeleted {
		return NewErrDocumentDeleted(primaryKey.DocID)
	}

	err = c.update(ctx, txn, doc)
	if err != nil {
		return err
	}

	return c.commitImplicitTxn(ctx, txn)
}

// Contract: DB Exists check is already performed, and a doc with the given ID exists.
// Note: Should we CompareAndSet the update, IE: Query(read-only) the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (c *collection) update(ctx context.Context, txn datastore.Txn, doc *client.Document) error {
	_, err := c.save(ctx, txn, doc, false)
	if err != nil {
		return err
	}
	return nil
}

// Save a document into the db.
// Either by creating a new document or by updating an existing one
func (c *collection) Save(ctx context.Context, doc *client.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	// Check if document already exists with primary DS key.
	primaryKey := c.getPrimaryKeyFromDocID(doc.ID())
	exists, isDeleted, err := c.exists(ctx, txn, primaryKey)
	if err != nil {
		return err
	}

	if isDeleted {
		return NewErrDocumentDeleted(doc.ID().String())
	}

	if exists {
		err = c.update(ctx, txn, doc)
	} else {
		err = c.create(ctx, txn, doc)
	}
	if err != nil {
		return err
	}

	return c.commitImplicitTxn(ctx, txn)
}

func (c *collection) save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
	isCreate bool,
) (cid.Cid, error) {
	if !isCreate {
		err := c.updateIndexedDoc(ctx, txn, doc)
		if err != nil {
			return cid.Undef, err
		}
	}
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
	links := make([]core.DAGLink, 0)
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

			fieldDescription, valid := c.Schema().GetField(k)
			if !valid {
				return cid.Undef, client.NewErrFieldNotExist(k)
			}

			// by default the type will have been set to LWW_REGISTER. We need to ensure
			// that it's set to the same as the field description CRDT type.
			val.SetType(fieldDescription.Typ)

			relationFieldDescription, isSecondaryRelationID := c.isSecondaryIDField(fieldDescription)
			if isSecondaryRelationID {
				primaryId := val.Value().(string)

				err = c.patchPrimaryDoc(ctx, txn, c.Name().Value(), relationFieldDescription, primaryKey.DocID, primaryId)
				if err != nil {
					return cid.Undef, err
				}

				// If this field was a secondary relation ID the related document will have been
				// updated instead and we should discard this value
				continue
			}

			err = c.validateOneToOneLinkDoesntAlreadyExist(ctx, txn, doc.ID().String(), fieldDescription, val.Value())
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

			node, _, err := merkleCRDT.Save(ctx, val)
			if err != nil {
				return cid.Undef, err
			}

			link := core.DAGLink{
				Name: k,
				Cid:  node.Cid(),
			}
			links = append(links, link)
		}
	}

	headNode, priority, err := c.saveCompositeToMerkleCRDT(
		ctx,
		txn,
		primaryKey.ToDataStoreKey(),
		links,
		client.Active,
	)
	if err != nil {
		return cid.Undef, err
	}

	if c.db.events.Updates.HasValue() {
		txn.OnSuccess(
			func() {
				c.db.events.Updates.Value().Publish(
					events.Update{
						DocID:      doc.ID().String(),
						Cid:        headNode.Cid(),
						SchemaRoot: c.Schema().Root,
						Block:      headNode,
						Priority:   priority,
					},
				)
			},
		)
	}

	txn.OnSuccess(func() {
		doc.SetHead(headNode.Cid())
	})

	return headNode.Cid(), nil
}

func (c *collection) validateOneToOneLinkDoesntAlreadyExist(
	ctx context.Context,
	txn datastore.Txn,
	docID string,
	fieldDescription client.FieldDescription,
	value any,
) error {
	if !fieldDescription.RelationType.IsSet(client.Relation_Type_INTERNAL_ID) {
		return nil
	}

	if value == nil {
		return nil
	}

	objFieldDescription, ok := c.Schema().GetField(strings.TrimSuffix(fieldDescription.Name, request.RelatedObjectID))
	if !ok {
		return client.NewErrFieldNotExist(strings.TrimSuffix(fieldDescription.Name, request.RelatedObjectID))
	}
	if !objFieldDescription.RelationType.IsSet(client.Relation_Type_ONEONE) {
		return nil
	}

	filter := fmt.Sprintf(
		`{_and: [{%s: {_ne: "%s"}}, {%s: {_eq: "%s"}}]}`,
		request.DocIDFieldName,
		docID,
		fieldDescription.Name,
		value,
	)
	selectionPlan, err := c.makeSelectionPlan(ctx, txn, filter)
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
func (c *collection) Delete(ctx context.Context, docID client.DocID) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	primaryKey := c.getPrimaryKeyFromDocID(docID)
	exists, isDeleted, err := c.exists(ctx, txn, primaryKey)
	if err != nil {
		return false, err
	}
	if !exists || isDeleted {
		return false, client.ErrDocumentNotFound
	}
	if isDeleted {
		return false, NewErrDocumentDeleted(primaryKey.DocID)
	}

	err = c.applyDelete(ctx, txn, primaryKey)
	if err != nil {
		return false, err
	}
	return true, c.commitImplicitTxn(ctx, txn)
}

// Exists checks if a given document exists with supplied DocID.
func (c *collection) Exists(ctx context.Context, docID client.DocID) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	primaryKey := c.getPrimaryKeyFromDocID(docID)
	exists, isDeleted, err := c.exists(ctx, txn, primaryKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return false, err
	}
	return exists && !isDeleted, c.commitImplicitTxn(ctx, txn)
}

// check if a document exists with the given primary key
func (c *collection) exists(
	ctx context.Context,
	txn datastore.Txn,
	primaryKey core.PrimaryDataStoreKey,
) (exists bool, isDeleted bool, err error) {
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

func (c *collection) saveCompositeToMerkleCRDT(
	ctx context.Context,
	txn datastore.Txn,
	dsKey core.DataStoreKey,
	links []core.DAGLink,
	status client.DocumentStatus,
) (ipld.Node, uint64, error) {
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

// getTxn gets or creates a new transaction from the underlying db.
// If the collection already has a txn, return the existing one.
// Otherwise, create a new implicit transaction.
func (c *collection) getTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	if c.txn.HasValue() {
		return c.txn.Value(), nil
	}
	return c.db.NewTxn(ctx, readonly)
}

// discardImplicitTxn is a proxy function used by the collection to execute the Discard()
// transaction function only if its an implicit transaction.
//
// Implicit transactions are transactions that are created *during* an operation execution as a side effect.
//
// Explicit transactions are provided to the collection object via the "WithTxn(...)" function.
func (c *collection) discardImplicitTxn(ctx context.Context, txn datastore.Txn) {
	if !c.txn.HasValue() {
		txn.Discard(ctx)
	}
}

func (c *collection) commitImplicitTxn(ctx context.Context, txn datastore.Txn) error {
	if !c.txn.HasValue() {
		return txn.Commit(ctx)
	}
	return nil
}

func (c *collection) getPrimaryKeyFromDocID(docID client.DocID) core.PrimaryDataStoreKey {
	return core.PrimaryDataStoreKey{
		CollectionId: fmt.Sprint(c.ID()),
		DocID:        docID.String(),
	}
}

func (c *collection) getDataStoreKeyFromDocID(docID client.DocID) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionID: fmt.Sprint(c.ID()),
		DocID:        docID.String(),
		InstanceType: core.ValueKey,
	}
}

func (c *collection) tryGetFieldKey(primaryKey core.PrimaryDataStoreKey, fieldName string) (core.DataStoreKey, bool) {
	fieldId, hasField := c.tryGetSchemaFieldID(fieldName)
	if !hasField {
		return core.DataStoreKey{}, false
	}

	return core.DataStoreKey{
		CollectionID: primaryKey.CollectionId,
		DocID:        primaryKey.DocID,
		FieldId:      strconv.FormatUint(uint64(fieldId), 10),
	}, true
}

// tryGetSchemaFieldID returns the FieldID of the given fieldName.
// Will return false if the field is not found.
func (c *collection) tryGetSchemaFieldID(fieldName string) (uint32, bool) {
	for _, field := range c.Schema().Fields {
		if field.Name == fieldName {
			if field.IsObject() || field.IsObjectArray() {
				// We do not wish to match navigational properties, only
				// fields directly on the collection.
				return uint32(0), false
			}
			return uint32(field.ID), true
		}
	}

	return uint32(0), false
}
