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
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/utils"

	"errors"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	mh "github.com/multiformats/go-multihash"
)

var (
	ErrDocumentAlreadyExists = errors.New("A document with the given key already exists")
	ErrUnknownCRDTArgument   = errors.New("Invalid CRDT arguments")
	ErrUnknownCRDT           = errors.New("")
)

var _ client.Collection = (*collection)(nil)

// collection stores data records at Documents, which are gathered
// together under a collection name. This is analogous to SQL Tables.
type collection struct {
	db  *db
	txn datastore.Txn

	colID uint32

	schemaID string

	desc base.CollectionDescription
}

// @todo: Move the base Descriptions to an internal API within the db/ package.
// @body: Currently, the New/Create Collection APIs accept CollectionDescriptions
// as params. We want these Descriptions objects to be low level descriptions, and
// to be auto generated based on a more controllable and user friendly
// CollectionOptions object.

// NewCollection returns a pointer to a newly instanciated DB Collection
func (db *db) newCollection(desc base.CollectionDescription) (*collection, error) {
	if desc.Name == "" {
		return nil, errors.New("Collection requires name to not be empty")
	}

	if len(desc.Schema.Fields) == 0 {
		return nil, errors.New("Collection schema has no fields")
	}

	docKeyField := desc.Schema.Fields[0]
	if docKeyField.Kind != base.FieldKind_DocKey || docKeyField.Name != "_key" {
		return nil, errors.New("Collection schema first field must be a DocKey")
	}

	desc.Schema.FieldIDs = make([]uint32, len(desc.Schema.Fields))
	for i, field := range desc.Schema.Fields {
		if field.Name == "" {
			return nil, errors.New("Collection schema field missing Name")
		}
		if field.Kind == base.FieldKind_None {
			return nil, errors.New("Collection schema field missing FieldKind")
		}
		if (field.Kind != base.FieldKind_DocKey && !field.IsObject()) && field.Typ == client.NONE_CRDT {
			return nil, errors.New("Collection schema field missing CRDT type")
		}
		desc.Schema.FieldIDs[i] = uint32(i)
		desc.Schema.Fields[i].ID = base.FieldID(i)
	}

	// for now, ignore any defined indexes, and overwrite the entire IndexDescription
	// property with the correct default one.
	desc.Indexes = []base.IndexDescription{
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
func (db *db) CreateCollection(ctx context.Context, desc base.CollectionDescription) (client.Collection, error) {
	// check if collection by this name exists
	cKey := core.NewCollectionKey(desc.Name)
	exists, err := db.systemstore().Has(ctx, cKey.ToDS())
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Collection already exists")
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

	buf, err := json.Marshal(col.desc)
	if err != nil {
		return nil, err
	}

	key := core.NewCollectionKey(col.desc.Name)

	// write the collection metadata to the system store
	err = db.systemstore().Put(ctx, key.ToDS(), buf)
	if err != nil {
		return nil, err
	}

	buf, err = json.Marshal(struct {
		Name   string
		Schema base.SchemaDescription
	}{col.desc.Name, col.desc.Schema})
	if err != nil {
		return nil, err
	}

	// add a reference to this DB by desc hash
	cid, err := utils.NewCidV1(buf)
	if err != nil {
		return nil, err
	}
	col.schemaID = cid.String()

	csKey := core.NewCollectionSchemaKey(cid.String())
	err = db.systemstore().Put(ctx, csKey.ToDS(), []byte(desc.Name))
	log.Debug(ctx, "Created collection", logging.NewKV("Name", col.Name()), logging.NewKV("Id", col.SchemaID))
	return col, err
}

// GetCollection returns an existing collection within the database
func (db *db) GetCollectionByName(ctx context.Context, name string) (client.Collection, error) {
	if name == "" {
		return nil, errors.New("Collection name can't be empty")
	}

	key := core.NewCollectionKey(name)
	buf, err := db.systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return nil, err
	}

	var desc base.CollectionDescription
	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return nil, err
	}

	buf, err = json.Marshal(struct {
		Name   string
		Schema base.SchemaDescription
	}{desc.Name, desc.Schema})
	if err != nil {
		return nil, err
	}

	// add a reference to this DB by desc hash
	cid, err := utils.NewCidV1(buf)
	if err != nil {
		return nil, err
	}

	sid := cid.String()
	log.Debug(ctx, "Retrieved collection", logging.NewKV("Name", desc.Name), logging.NewKV("Id", sid))

	return &collection{
		db:       db,
		desc:     desc,
		colID:    desc.ID,
		schemaID: sid,
	}, nil
}

// GetCollectionBySchemaID returns an existing collection within the database using the
// schema hash ID
func (db *db) GetCollectionBySchemaID(ctx context.Context, schemaID string) (client.Collection, error) {
	if schemaID == "" {
		return nil, fmt.Errorf("Schema ID can't be empty")
	}

	key := core.NewCollectionSchemaKey(schemaID)
	buf, err := db.systemstore().Get(ctx, key.ToDS())
	if err != nil {
		return nil, err
	}

	name := string(buf)
	return db.GetCollectionByName(ctx, name)
}

// GetAllCollections gets all the currently defined collections in the
// database
func (db *db) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	// create collection system prefix query
	prefix := core.NewCollectionKey("")
	q, err := db.systemstore().Query(ctx, query.Query{
		Prefix:   prefix.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create collection prefix query: %w", err)
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

		colName := ds.NewKey(res.Key).BaseNamespace()
		col, err := db.GetCollectionByName(ctx, colName)
		if err != nil {
			return nil, fmt.Errorf("Failed to get collection (%s): %w", colName, err)
		}
		cols = append(cols, col)
	}

	return cols, nil
}

// GetAllDocKeys returns all the document keys that exist in the collection
// @todo: We probably need a lock on the collection for this kind of op since
// it hits every key and will cause Tx conflicts for concurrent Txs
func (c *collection) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	return c.getAllDocKeysChan(ctx, txn)
}

func (c *collection) getAllDocKeysChan(ctx context.Context, txn datastore.Txn) (<-chan client.DocKeysResult, error) {
	prefix := c.getPrimaryIndexDocKey(core.DataStoreKey{}) // empty path for all keys prefix
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

			// looking for /<colID>/1/<Dockey>:v

			// not it - prob a child key
			if strings.Count(res.Key, "/") != 3 {
				continue
			}

			// not it - missing value suffix
			if !strings.HasSuffix(res.Key, ":v") {
				continue
			}

			// now we have a doc key
			rawDocKey := ds.NewKey(res.Key).Type()
			key, err := key.NewFromString(rawDocKey)
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

// Description returns the base.CollectionDescription
func (c *collection) Description() base.CollectionDescription {
	return c.desc
}

// Name returns the collection name
func (c *collection) Name() string {
	return c.desc.Name
}

// Schema returns the Schema of the collection
func (c *collection) Schema() base.SchemaDescription {
	return c.desc.Schema
}

// ID returns the ID of the collection
func (c *collection) ID() uint32 {
	return c.colID
}

// Indexes returns the defined indexes on the Collection
// @todo: Properly handle index creation/management
func (c *collection) Indexes() []base.IndexDescription {
	return c.desc.Indexes
}

// PrimaryIndex returns the primary index for the given collection
func (c *collection) PrimaryIndex() base.IndexDescription {
	return c.desc.Indexes[0]
}

// Index returns the index with the given index ID
func (c *collection) Index(id uint32) (base.IndexDescription, error) {
	for _, index := range c.desc.Indexes {
		if index.ID == id {
			return index, nil
		}
	}

	return base.IndexDescription{}, errors.New("No index found for given ID")
}

// CreateIndex creates a new index on the collection. Custom indexes
// are always "Secondary indexes". Primary indexes are automatically created
// on Collection creation, and cannot be changed.
func (c *collection) CreateIndex(idesc base.IndexDescription) error {
	panic("not implemented")
}

func (c *collection) SchemaID() string {
	return c.schemaID
}

// WithTxn returns a new instance of the collection, with a transaction
// handle instead of a raw DB handle
func (c *collection) WithTxn(txn datastore.Txn) client.Collection {
	return &collection{
		db:       c.db,
		txn:      txn,
		desc:     c.desc,
		colID:    c.colID,
		schemaID: c.schemaID,
	}
}

// Create a new document
// Will verify the DocKey/CID to ensure that the new document is correctly formatted.
func (c *collection) Create(ctx context.Context, doc *document.Document) error {
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
func (c *collection) CreateMany(ctx context.Context, docs []*document.Document) error {
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

func (c *collection) create(ctx context.Context, txn datastore.Txn, doc *document.Document) error {
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

	dockey := key.NewDocKeyV0(doccid)
	key := core.DataStoreKeyFromDocKey(dockey)
	if key.DocKey != doc.Key().String() {
		return fmt.Errorf("Expected %s, got %s : %w", doc.Key(), key.DocKey, ErrDocVerification)
	}

	// check if doc already exists
	exists, err := c.exists(ctx, txn, key)
	if err != nil {
		return err
	}
	if exists {
		return ErrDocumentAlreadyExists
	}

	// write object marker
	err = writeObjectMarker(ctx, txn.Datastore(), c.getPrimaryIndexDocKey(key))
	if err != nil {
		return err
	}
	// write data to DB via MerkleClock/CRDT
	_, err = c.save(ctx, txn, doc)
	return err
}

// Update an existing document with the new values
// Any field that needs to be removed or cleared
// should call doc.Clear(field) before. Any field that
// is nil/empty that hasn't called Clear will be ignored
func (c *collection) Update(ctx context.Context, doc *document.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	dockey := core.DataStoreKeyFromDocKey(doc.Key())
	exists, err := c.exists(ctx, txn, dockey)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDocumentNotFound
	}

	err = c.update(ctx, txn, doc)
	if err != nil {
		return err
	}

	return c.commitImplicitTxn(ctx, txn)
}

// Contract: DB Exists check is already performed, and a doc with the given key exists
// Note: Should we CompareAndSet the update, IE: Query the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (c *collection) update(ctx context.Context, txn datastore.Txn, doc *document.Document) error {
	_, err := c.save(ctx, txn, doc)
	if err != nil {
		return err
	}
	return nil
}

// Save a document into the db
// Either by creating a new document or by updating an existing one
func (c *collection) Save(ctx context.Context, doc *document.Document) error {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(ctx, txn)

	// Check if document already exists with key
	dockey := core.DataStoreKeyFromDocKey(doc.Key())
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

func (c *collection) save(ctx context.Context, txn datastore.Txn, doc *document.Document) (cid.Cid, error) {
	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instantiate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	dockey := core.DataStoreKeyFromDocKey(doc.Key())
	links := make([]core.DAGLink, 0)
	merge := make(map[string]interface{})
	for k, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		if val.IsDirty() {
			fieldKey := c.getFieldKey(dockey, k)
			c, err := c.saveDocValue(ctx, txn, c.getPrimaryIndexDocKey(fieldKey), val)
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
				Cid:  c,
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

	headCID, err := c.saveValueToMerkleCRDT(ctx, txn, c.getPrimaryIndexDocKey(dockey), client.COMPOSITE, buf, links)
	if err != nil {
		return cid.Undef, err
	}

	txn.OnSuccess(func() {
		doc.SetHead(headCID)
	})

	return headCID, nil
}

// Delete will attempt to delete a document by key
// will return true if a deletion is successful, and
// return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return
// false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given
// DocKey. This includes data, block, and head storage.
func (c *collection) Delete(ctx context.Context, key key.DocKey) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	dsKey := core.DataStoreKeyFromDocKey(key)
	exists, err := c.exists(ctx, txn, dsKey)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrDocumentNotFound
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
func (c *collection) delete(ctx context.Context, txn datastore.Txn, key core.DataStoreKey) (bool, error) {
	q := query.Query{
		Prefix:   c.getPrimaryIndexDocKey(key).ToString(),
		KeysOnly: true,
	}
	res, err := txn.Datastore().Query(ctx, q)

	for e := range res.Next() {
		if e.Error != nil {
			return false, err
		}

		err = txn.Datastore().Delete(ctx, c.getPrimaryIndexDocKey(core.NewDataStoreKey(e.Key)).ToDS())
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// Exists checks if a given document exists with supplied DocKey
func (c *collection) Exists(ctx context.Context, key key.DocKey) (bool, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	dsKey := core.DataStoreKeyFromDocKey(key)
	exists, err := c.exists(ctx, txn, dsKey)
	if err != nil && err != ds.ErrNotFound {
		return false, err
	}
	return exists, c.commitImplicitTxn(ctx, txn)
}

// check if a document exists with the given key
func (c *collection) exists(ctx context.Context, txn datastore.Txn, key core.DataStoreKey) (bool, error) {
	return txn.Datastore().Has(ctx, c.getPrimaryIndexDocKey(key.WithValueFlag()).ToDS())
}

func (c *collection) saveDocValue(ctx context.Context, txn datastore.Txn, key core.DataStoreKey, val document.Value) (cid.Cid, error) {
	switch val.Type() {
	case client.LWW_REGISTER:
		wval, ok := val.(document.WriteableValue)
		if !ok {
			return cid.Cid{}, document.ErrValueTypeMismatch
		}
		var bytes []byte
		var err error
		if val.IsDelete() { // empty byte array
			bytes = []byte{}
		} else {
			bytes, err = wval.Bytes()
			if err != nil {
				return cid.Cid{}, err
			}
		}
		return c.saveValueToMerkleCRDT(ctx, txn, key, client.LWW_REGISTER, bytes)
	default:
		return cid.Cid{}, ErrUnknownCRDT
	}
}

func (c *collection) saveValueToMerkleCRDT(
	ctx context.Context,
	txn datastore.Txn,
	key core.DataStoreKey,
	ctype client.CType,
	args ...interface{}) (cid.Cid, error) {
	switch ctype {
	case client.LWW_REGISTER:
		datatype, err := c.db.crdtFactory.InstanceWithStores(txn, c.schemaID, c.db.broadcaster, ctype, key)
		if err != nil {
			return cid.Cid{}, err
		}

		var bytes []byte
		var ok bool
		// parse args
		if len(args) != 1 {
			return cid.Cid{}, ErrUnknownCRDTArgument
		}
		bytes, ok = args[0].([]byte)
		if !ok {
			return cid.Cid{}, ErrUnknownCRDTArgument
		}
		lwwreg := datatype.(*crdt.MerkleLWWRegister)
		return lwwreg.Set(ctx, bytes)
	case client.COMPOSITE:
		key = key.WithFieldId(core.COMPOSITE_NAMESPACE)
		datatype, err := c.db.crdtFactory.InstanceWithStores(txn, c.SchemaID(), c.db.broadcaster, ctype, key)
		if err != nil {
			return cid.Cid{}, err
		}
		var bytes []byte
		var links []core.DAGLink
		var ok bool
		// parse args
		if len(args) != 2 {
			return cid.Cid{}, ErrUnknownCRDTArgument
		}
		bytes, ok = args[0].([]byte)
		if !ok {
			return cid.Cid{}, ErrUnknownCRDTArgument
		}
		links, ok = args[1].([]core.DAGLink)
		if !ok {
			return cid.Cid{}, ErrUnknownCRDTArgument
		}
		comp := datatype.(*crdt.MerkleCompositeDAG)
		return comp.Set(ctx, bytes, links)
	}
	return cid.Cid{}, ErrUnknownCRDT
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

// discardImplicitTxn is a proxy function used by the collection to execute the Discard() transaction
// function only if its an implicit transaction.
// Implicit transactions are transactions that are created *during* an operation execution as a side effect.
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

func (c *collection) getPrimaryIndexDocKey(key core.DataStoreKey) core.DataStoreKey {
	return core.DataStoreKey{
		CollectionId: fmt.Sprint(c.colID),
		IndexId:      fmt.Sprint(c.PrimaryIndex().ID),
	}.WithInstanceInfo(key)
}

func (c *collection) getFieldKey(key core.DataStoreKey, fieldName string) core.DataStoreKey {
	return key.WithFieldId(fmt.Sprint(c.getSchemaFieldID(fieldName)))
}

// getSchemaFieldID returns the FieldID of the given fieldName.
// It assumes a schema exists for the collection, and that the
// field exists in the schema.
func (c *collection) getSchemaFieldID(fieldName string) uint32 {
	for _, field := range c.desc.Schema.Fields {
		if field.Name == fieldName {
			return uint32(field.ID)
		}
	}
	return uint32(0)
}

func writeObjectMarker(ctx context.Context, store ds.Write, key core.DataStoreKey) error {
	return store.Put(ctx, key.WithValueFlag().ToDS(), []byte{base.ObjectMarker})
}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
// func makeCollectionDataKey(collectionID uint32) core.Key {
// 	return collectionNs.ChildString(name)
// }
