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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	ipld "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db  *db
	txn datastore.Txn

	colID uint32

	schemaID string

	desc client.CollectionDescription
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionDescriptions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// NewCollection returns a pointer to a newly instanciated DB Collection
func (db *db) newCollection(desc client.CollectionDescription) (*collection, error) {
	if desc.Name == "" {
		return nil, client.NewErrUninitializeProperty("Collection", "Name")
	}

	if len(desc.Schema.Fields) == 0 {
		return nil, client.NewErrUninitializeProperty("Collection", "Fields")
	}

	docKeyField := desc.Schema.Fields[0]
	if docKeyField.Kind != client.FieldKind_DocKey || docKeyField.Name != "_key" {
		return nil, ErrSchemaFirstFieldDocKey
	}

	desc.Schema.FieldIDs = make([]uint32, len(desc.Schema.Fields))
	for i, field := range desc.Schema.Fields {
		if field.Name == "" {
			return nil, client.NewErrUninitializeProperty("Collection.Schema", "Name")
		}
		if field.Kind == client.FieldKind_None {
			return nil, client.NewErrUninitializeProperty("Collection.Schema", "FieldKind")
		}
		if (field.Kind != client.FieldKind_DocKey && !field.IsObject()) &&
			field.Typ == client.NONE_CRDT {
			return nil, client.NewErrUninitializeProperty("Collection.Schema", "CRDT type")
		}
		desc.Schema.FieldIDs[i] = uint32(i)
		desc.Schema.Fields[i].ID = client.FieldID(i)
	}

	// for now, ignore any defined indexes, and overwrite the entire IndexDescription
	// property with the correct default one.
	desc.Indexes = []client.IndexDescription{
		{
			Name:    "primary",
			ID:      uint32(0),
			Primary: true,
			Unique:  true,
		},
	}

	return &collection{
		db:    db,
		desc:  desc,
		colID: desc.ID,
	}, nil
}

// CreateCollection creates a collection and saves it to the database in its system store.
// Note: Collection.ID is an autoincrementing value that is generated by the database.
func (db *db) CreateCollection(
	ctx context.Context,
	desc client.CollectionDescription,
) (client.Collection, error) {
	// check if collection by this name exists
	collectionKey := core.NewCollectionKey(desc.Name)
	exists, err := db.systemstore().Has(ctx, collectionKey.ToDS())
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCollectionAlreadyExists
	}

	colSeq, err := db.getSequence(ctx, core.COLLECTION)
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next(ctx)
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)
	col, err := db.newCollection(desc)
	if err != nil {
		return nil, err
	}

	// Local elements such as secondary indexes should be excluded
	// from the (global) schemaId.
	globalSchemaBuf, err := json.Marshal(struct {
		Name   string
		Schema client.SchemaDescription
	}{col.desc.Name, col.desc.Schema})
	if err != nil {
		return nil, err
	}

	// add a reference to this DB by desc hash
	cid, err := core.NewSHA256CidV1(globalSchemaBuf)
	if err != nil {
		return nil, err
	}
	schemaID := cid.String()
	col.schemaID = schemaID

	// For new schemas the initial version id will match the schema id
	schemaVersionID := schemaID

	col.desc.Schema.VersionID = schemaVersionID
	col.desc.Schema.SchemaID = schemaID

	// buffer must include all the ids, as it is saved and loaded from the store later.
	buf, err := json.Marshal(col.desc)
	if err != nil {
		return nil, err
	}

	collectionSchemaVersionKey := core.NewCollectionSchemaVersionKey(schemaVersionID)
	// Whilst the schemaVersionKey is global, the data persisted at the key's location
	// is local to the node (the global only elements are not useful beyond key generation).
	err = db.systemstore().Put(ctx, collectionSchemaVersionKey.ToDS(), buf)
	if err != nil {
		return nil, err
	}

	collectionSchemaKey := core.NewCollectionSchemaKey(schemaID)
	err = db.systemstore().Put(ctx, collectionSchemaKey.ToDS(), []byte(schemaVersionID))
	if err != nil {
		return nil, err
	}

	err = db.systemstore().Put(ctx, collectionKey.ToDS(), []byte(schemaVersionID))
	if err != nil {
		return nil, err
	}

	log.Debug(
		ctx,
		"Created collection",
		logging.NewKV("Name", col.Name()),
		logging.NewKV("ID", col.SchemaID),
	)
	return col, nil
}

// getCollectionByVersionId returns the [*collection] at the given [schemaVersionId] version.
//
// Will return an error if the given key is empty, or not found.
func (db *db) getCollectionByVersionId(ctx context.Context, schemaVersionId string) (*collection, error) {
	if schemaVersionId == "" {
		return nil, ErrSchemaVersionIdEmpty
	}

	key := core.NewCollectionSchemaVersionKey(schemaVersionId)
	buf, err := db.systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return nil, err
	}

	var desc client.CollectionDescription
	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return nil, err
	}

	return &collection{
		db:       db,
		desc:     desc,
		colID:    desc.ID,
		schemaID: desc.Schema.SchemaID,
	}, nil
}

// GetCollection returns an existing collection within the database.
func (db *db) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	if name == "" {
		return nil, ErrCollectionNameEmpty
	}

	key := core.NewCollectionKey(name)
	buf, err := db.systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return nil, err
	}

	schemaVersionId := string(buf)
	return db.getCollectionByVersionId(ctx, schemaVersionId)
}

// GetCollectionBySchemaID returns an existing collection using the schema hash ID.
func (db *db) GetCollectionBySchemaID(
	ctx context.Context,
	schemaID string,
) (client.Collection, error) {
	if schemaID == "" {
		return nil, ErrSchemaIdEmpty
	}

	key := core.NewCollectionSchemaKey(schemaID)
	buf, err := db.systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return nil, err
	}

	schemaVersionId := string(buf)
	return db.getCollectionByVersionId(ctx, schemaVersionId)
}

// GetAllCollections gets all the currently defined collections.
func (db *db) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	// create collection system prefix query
	prefix := core.NewCollectionSchemaVersionKey("")
	q, err := db.systemstore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	cols := make([]client.Collection, 0)
	for res := range q.Next() {
		if res.Error != nil {
			return nil, err
		}

		schemaVersionId := ds.NewKey(res.Key).BaseNamespace()
		col, err := db.getCollectionByVersionId(ctx, schemaVersionId)
		if err != nil {
			return nil, NewErrFailedToGetCollection(schemaVersionId, err)
		}
		cols = append(cols, col)
	}

	return cols, nil
}

// GetAllDocKeys returns all the document keys that exist in the collection.
//
// @todo: We probably need a lock on the collection for this kind of op since
// it hits every key and will cause Tx conflicts for concurrent Txs
func (c *collection) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return nil, err
	}

	return c.getAllDocKeysChan(ctx, txn)
}

func (c *collection) getAllDocKeysChan(
	ctx context.Context,
	txn datastore.Txn,
) (<-chan client.DocKeysResult, error) {
	prefix := core.PrimaryDataStoreKey{ // empty path for all keys prefix
		CollectionId: fmt.Sprint(c.colID),
	}
	q, err := txn.Datastore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	resCh := make(chan client.DocKeysResult)
	go func() {
		defer func() {
			if err := q.Close(); err != nil {
				log.ErrorE(ctx, "Failed to close AllDocKeys query", err)
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
				resCh <- client.DocKeysResult{
					Err: res.Error,
				}
				return
			}

			// now we have a doc key
			rawDocKey := ds.NewKey(res.Key).BaseNamespace()
			key, err := client.NewDocKeyFromString(rawDocKey)
			if err != nil {
				resCh <- client.DocKeysResult{
					Err: res.Error,
				}
				return
			}
			resCh <- client.DocKeysResult{
				Key: key,
			}
		}
	}()

	return resCh, nil
}

// Description returns the client.CollectionDescription.
func (c *collection) Description() client.CollectionDescription {
	return c.desc
}

// Name returns the collection name.
func (c *collection) Name() string {
	return c.desc.Name
}

// Schema returns the Schema of the collection.
func (c *collection) Schema() client.SchemaDescription {
	return c.desc.Schema
}

// ID returns the ID of the collection.
func (c *collection) ID() uint32 {
	return c.colID
}

// Indexes returns the defined indexes on the Collection.
// @todo: Properly handle index creation/management
func (c *collection) Indexes() []client.IndexDescription {
	return c.desc.Indexes
}

// PrimaryIndex returns the primary index for the given collection.
func (c *collection) PrimaryIndex() client.IndexDescription {
	return c.desc.Indexes[0]
}

// Index returns the index with the given index ID.
func (c *collection) Index(id uint32) (client.IndexDescription, error) {
	for _, index := range c.desc.Indexes {
		if index.ID == id {
			return index, nil
		}
	}

	return client.IndexDescription{}, client.ErrIndexNotFound
}

func (c *collection) SchemaID() string {
	return c.schemaID
}

// WithTxn returns a new instance of the collection, with a transaction
// handle instead of a raw DB handle.
func (c *collection) WithTxn(txn datastore.Txn) client.Collection {
	return &collection{
		db:       c.db,
		txn:      txn,
		desc:     c.desc,
		colID:    c.colID,
		schemaID: c.schemaID,
	}
}

// Create a new document.
// Will verify the DocKey/CID to ensure that the new document is correctly formatted.
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
// Will verify the DocKey/CID to ensure that the new documents are correctly formatted.
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

func (c *collection) create(ctx context.Context, txn datastore.Txn, doc *client.Document) error {
	// DocKey verification
	buf, err := doc.Bytes()
	if err != nil {
		return err
	}
	// @todo:  grab the cid Prefix from the DocKey internal CID if available
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}
	// And then feed it some data
	doccid, err := pref.Sum(buf)
	if err != nil {
		return err
	}

	dockey := client.NewDocKeyV0(doccid)
	key := c.getPrimaryKeyFromDocKey(dockey)
	if key.DocKey != doc.Key().String() {
		return NewErrDocVerification(doc.Key().String(), key.DocKey)
	}

	// check if doc already exists
	exists, err := c.exists(ctx, txn, key)
	if err != nil {
		return err
	}
	if exists {
		return ErrDocumentAlreadyExists
	}

	// write value object marker if we have an empty doc
	if len(doc.Values()) == 0 {
		valueKey := c.getDSKeyFromDockey(dockey)
		err = txn.Datastore().Put(ctx, valueKey.ToDS(), []byte{base.ObjectMarker})
		if err != nil {
			return err
		}
	}

	// write data to DB via MerkleClock/CRDT
	_, err = c.save(ctx, txn, doc)
	if err != nil {
		return err
	}

	return err
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

	dockey := c.getPrimaryKeyFromDocKey(doc.Key())
	exists, err := c.exists(ctx, txn, dockey)
	if err != nil {
		return err
	}
	if !exists {
		return client.ErrDocumentNotFound
	}

	err = c.update(ctx, txn, doc)
	if err != nil {
		return err
	}

	return c.commitImplicitTxn(ctx, txn)
}

// Contract: DB Exists check is already performed, and a doc with the given key exists.
// Note: Should we CompareAndSet the update, IE: Query(read-only) the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (c *collection) update(ctx context.Context, txn datastore.Txn, doc *client.Document) error {
	_, err := c.save(ctx, txn, doc)
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

	// Check if document already exists with key
	dockey := c.getPrimaryKeyFromDocKey(doc.Key())
	exists, err := c.exists(ctx, txn, dockey)
	if err != nil {
		return err
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
) (cid.Cid, error) {
	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instantiate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	primaryKey := c.getPrimaryKeyFromDocKey(doc.Key())
	links := make([]core.DAGLink, 0)
	merge := make(map[string]any)
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

			if c.isFieldNameRelationID(k) {
				return cid.Undef, client.NewErrFieldNotExist(k)
			}

			c, _, err := c.saveDocValue(ctx, txn, fieldKey, val)
			if err != nil {
				return cid.Undef, err
			}
			if val.IsDelete() {
				merge[k] = nil
			} else {
				merge[k] = val.Value()
			}

			// NOTE: We delay the final Clean() call until we know
			// the commit on the transaction is successful. If we didn't
			// wait, and just did it here, then *if* the commit fails down
			// the line, then we have no way to roll back the state
			// side-effect on the document func called here.
			txn.OnSuccess(func() {
				doc.Clean()
			})

			link := core.DAGLink{
				Name: k,
				Cid:  c.Cid(),
			}
			links = append(links, link)
		}
	}
	// Update CompositeDAG
	em, err := cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		return cid.Undef, err
	}
	buf, err := em.Marshal(merge)
	if err != nil {
		return cid.Undef, nil
	}

	headNode, priority, err := c.saveValueToMerkleCRDT(
		ctx,
		txn,
		primaryKey.ToDataStoreKey(),
		client.COMPOSITE,
		buf,
		links,
	)
	if err != nil {
		return cid.Undef, err
	}

	if c.db.events.Updates.HasValue() {
		txn.OnSuccess(
			func() {
				c.db.events.Updates.Value().Publish(
					events.Update{
						DocKey:   doc.Key().String(),
						Cid:      headNode.Cid(),
						SchemaID: c.schemaID,
						Block:    headNode,
						Priority: priority,
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

// Delete will attempt to delete a document by key will return true if a deletion is successful,
// and return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given DocKey. This includes data, block, and head storage.
func (c *collection) Delete(ctx context.Context, key client.DocKey) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	dsKey := c.getPrimaryKeyFromDocKey(key)
	exists, err := c.exists(ctx, txn, dsKey)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, client.ErrDocumentNotFound
	}

	// run delete, commit if successful
	deleted, err := c.delete(ctx, txn, dsKey)
	if err != nil {
		return false, err
	}
	return deleted, c.commitImplicitTxn(ctx, txn)
}

// at the moment, delete only does data storage delete.
// Dag, and head store will soon follow.
func (c *collection) delete(
	ctx context.Context,
	txn datastore.Txn,
	key core.PrimaryDataStoreKey,
) (bool, error) {
	err := txn.Datastore().Delete(ctx, key.ToDS())
	if err != nil {
		return false, err
	}

	// delete value instance
	keyDS := key.ToDataStoreKey()
	keyDS.InstanceType = core.ValueKey
	if _, err = c.deleteWithPrefix(ctx, txn, keyDS); err != nil {
		return false, err
	}

	// delete priority instance
	keyDS.InstanceType = core.PriorityKey
	return c.deleteWithPrefix(ctx, txn, keyDS)
}

// deleteWithPrefix will delete all the keys using a prefix query set as the given key.
func (c *collection) deleteWithPrefix(ctx context.Context, txn datastore.Txn, key core.DataStoreKey) (bool, error) {
	q := query.Query{
		Prefix:   key.ToString(),
		KeysOnly: true,
	}

	if key.InstanceType == core.ValueKey {
		err := txn.Datastore().Delete(ctx, key.ToDS())
		if err != nil {
			return false, err
		}
	}

	res, err := txn.Datastore().Query(ctx, q)

	for e := range res.Next() {
		if e.Error != nil {
			return false, err
		}

		dsKey, err := core.NewDataStoreKey(e.Key)
		if err != nil {
			return false, err
		}

		err = txn.Datastore().Delete(ctx, dsKey.ToDS())
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// Exists checks if a given document exists with supplied DocKey.
func (c *collection) Exists(ctx context.Context, key client.DocKey) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	dsKey := c.getPrimaryKeyFromDocKey(key)
	exists, err := c.exists(ctx, txn, dsKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return false, err
	}
	return exists, c.commitImplicitTxn(ctx, txn)
}

// check if a document exists with the given key
func (c *collection) exists(
	ctx context.Context,
	txn datastore.Txn,
	key core.PrimaryDataStoreKey,
) (bool, error) {
	return txn.Datastore().Has(ctx, key.ToDS())
}

func (c *collection) saveDocValue(
	ctx context.Context,
	txn datastore.Txn,
	key core.DataStoreKey,
	val client.Value,
) (ipld.Node, uint64, error) {
	switch val.Type() {
	case client.LWW_REGISTER:
		wval, ok := val.(client.WriteableValue)
		if !ok {
			return nil, 0, client.ErrValueTypeMismatch
		}
		var bytes []byte
		var err error
		if val.IsDelete() { // empty byte array
			bytes = []byte{}
		} else {
			bytes, err = wval.Bytes()
			if err != nil {
				return nil, 0, err
			}
		}
		return c.saveValueToMerkleCRDT(ctx, txn, key, client.LWW_REGISTER, bytes)
	default:
		return nil, 0, ErrUnknownCRDT
	}
}

func (c *collection) saveValueToMerkleCRDT(
	ctx context.Context,
	txn datastore.Txn,
	key core.DataStoreKey,
	ctype client.CType,
	args ...any) (ipld.Node, uint64, error) {
	switch ctype {
	case client.LWW_REGISTER:
		datatype, err := c.db.crdtFactory.InstanceWithStores(
			txn,
			core.NewCollectionSchemaVersionKey(c.Schema().VersionID),
			c.db.events.Updates,
			ctype,
			key,
		)
		if err != nil {
			return nil, 0, err
		}

		var bytes []byte
		var ok bool
		// parse args
		if len(args) != 1 {
			return nil, 0, ErrUnknownCRDTArgument
		}
		bytes, ok = args[0].([]byte)
		if !ok {
			return nil, 0, ErrUnknownCRDTArgument
		}
		lwwreg := datatype.(*crdt.MerkleLWWRegister)
		return lwwreg.Set(ctx, bytes)
	case client.COMPOSITE:
		key = key.WithFieldId(core.COMPOSITE_NAMESPACE)
		datatype, err := c.db.crdtFactory.InstanceWithStores(
			txn,
			core.NewCollectionSchemaVersionKey(c.Schema().VersionID),
			c.db.events.Updates,
			ctype,
			key,
		)
		if err != nil {
			return nil, 0, err
		}
		var bytes []byte
		var links []core.DAGLink
		var ok bool
		// parse args
		if len(args) != 2 {
			return nil, 0, ErrUnknownCRDTArgument
		}
		bytes, ok = args[0].([]byte)
		if !ok {
			return nil, 0, ErrUnknownCRDTArgument
		}
		links, ok = args[1].([]core.DAGLink)
		if !ok {
			return nil, 0, ErrUnknownCRDTArgument
		}
		comp := datatype.(*crdt.MerkleCompositeDAG)
		return comp.Set(ctx, bytes, links)
	}
	return nil, 0, ErrUnknownCRDT
}

// getTxn gets or creates a new transaction from the underlying db.
// If the collection already has a txn, return the existing one.
// Otherwise, create a new implicit transaction.
func (c *collection) getTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	if c.txn != nil {
		return c.txn, nil
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
	if c.txn == nil {
		txn.Discard(ctx)
	}
}

func (c *collection) commitImplicitTxn(ctx context.Context, txn datastore.Txn) error {
	if c.txn == nil {
		return txn.Commit(ctx)
	}
	return nil
}

func (c *collection) getPrimaryKey(docKey string) core.PrimaryDataStoreKey {
	return core.PrimaryDataStoreKey{
		CollectionId: fmt.Sprint(c.colID),
		DocKey:       docKey,
	}
}

func (c *collection) getPrimaryKeyFromDocKey(docKey client.DocKey) core.PrimaryDataStoreKey {
	return core.PrimaryDataStoreKey{
		CollectionId: fmt.Sprint(c.colID),
		DocKey:       docKey.String(),
	}
}

func (c *collection) getDSKeyFromDockey(docKey client.DocKey) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: fmt.Sprint(c.colID),
		DocKey:       docKey.String(),
		InstanceType: core.ValueKey,
	}
}

func (c *collection) tryGetFieldKey(key core.PrimaryDataStoreKey, fieldName string) (core.DataStoreKey, bool) {
	fieldId, hasField := c.tryGetSchemaFieldID(fieldName)
	if !hasField {
		return core.DataStoreKey{}, false
	}

	return core.DataStoreKey{
		CollectionId: key.CollectionId,
		DocKey:       key.DocKey,
		FieldId:      strconv.FormatUint(uint64(fieldId), 10),
	}, true
}

// tryGetSchemaFieldID returns the FieldID of the given fieldName.
// Will return false if the field is not found.
func (c *collection) tryGetSchemaFieldID(fieldName string) (uint32, bool) {
	for _, field := range c.desc.Schema.Fields {
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

// isFieldNameRelationID returns true if the given field is the id field backing a relationship.
func (c *collection) isFieldNameRelationID(fieldName string) bool {
	fieldDescription, valid := c.desc.GetField(fieldName)
	if !valid {
		return false
	}

	return c.isFieldDescriptionRelationID(&fieldDescription)
}

// isFieldDescriptionRelationID returns true if the given field is the id field backing a relationship.
func (c *collection) isFieldDescriptionRelationID(fieldDescription *client.FieldDescription) bool {
	if fieldDescription.RelationType == client.Relation_Type_INTERNAL_ID {
		relationDescription, valid := c.desc.GetField(strings.TrimSuffix(fieldDescription.Name, "_id"))
		if !valid {
			return false
		}
		if relationDescription.IsPrimaryRelation() {
			return true
		}
	}
	return false
}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
// func makeCollectionDataKey(collectionID uint32) core.Key {
// 	return collectionNs.ChildString(name)
// }
