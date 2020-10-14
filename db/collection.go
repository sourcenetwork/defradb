package db

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	mh "github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/crdt"
)

var (
	collectionSeqKey = "collection"
	collectionNs     = ds.NewKey("/collection")
)

type Collection struct {
	db  *DB
	txn *Txn

	colID    uint32
	colIDKey core.Key

	desc base.CollectionDescription
}

// NewCollection returns a pointer to a newly instanciated DB Collection
func (db *DB) newCollection(desc base.CollectionDescription) (*Collection, error) {
	if desc.Name == "" {
		return nil, errors.New("Collection requires name to not be empty")
	}

	if desc.ID == 0 {
		return nil, errors.New("Collection ID must be greater than 0")
	}

	if !desc.Schema.IsEmpty() {
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
			if field.Typ == core.NONE_CRDT {
				return nil, errors.New("Collection schema field missing CRDT type")
			}
			desc.Schema.FieldIDs = append(desc.Schema.FieldIDs, uint32(i))
			desc.Schema.Fields[i].ID = uint32(i)
		}
	}

	return &Collection{
		db:       db,
		desc:     desc,
		colID:    desc.ID,
		colIDKey: core.NewKey(fmt.Sprint(desc.ID)),
	}, nil
}

// CreateCollection creates a collection and saves it to the database in its system store.
// Note: Collection.ID is an autoincrementing value that is generated by the database.
func (db *DB) CreateCollection(desc base.CollectionDescription) (*Collection, error) {
	colSeq, err := db.getSequence(collectionSeqKey)
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next()
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)
	col, err := db.newCollection(desc)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(desc)
	if err != nil {
		return nil, err
	}
	key := makeCollectionSystemKey(desc.Name)

	//write the collection metadata to the system store
	err = db.systemstore.Put(key, buf)
	return col, err
}

// GetCollection returns an existing collection within the database
func (db *DB) GetCollection(name string) (*Collection, error) {
	if name == "" {
		return nil, errors.New("Collection name can't be empty")
	}

	key := makeCollectionSystemKey(name)
	buf, err := db.systemstore.Get(key)
	if err != nil {
		return nil, err
	}
	var col *Collection
	err = json.Unmarshal(buf, col)
	return col, err
}

func (c *Collection) ValidDescription() bool {
	return false
}

func (c *Collection) WithTxn(txn *Txn) *Collection {
	return &Collection{
		txn:      txn,
		desc:     c.desc,
		colID:    c.colID,
		colIDKey: c.colIDKey,
	}
}

// Create a new document
// Will verify the DocKey/CID to ensure that the new document is correctly formatted.
func (c *Collection) Create(doc *document.Document) error {
	txn, err := c.getTxn(false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(txn)

	err = c.create(txn, doc)
	if err != nil {
		return err
	}
	return c.commitImplicitTxn(txn)
}

// CreateMany creates a collection of documents at once.
// Will verify the DocKey/CID to ensure that the new documents are correctly formatted.
func (c *Collection) CreateMany(docs []*document.Document) error {
	txn, err := c.getTxn(false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(txn)

	for _, doc := range docs {
		err = c.create(txn, doc)
		if err != nil {
			return err
		}
	}
	return c.commitImplicitTxn(txn)
}

func (c *Collection) create(txn *Txn, doc *document.Document) error {
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
	// fmt.Println(c)
	dockey := key.NewDocKeyV0(doccid)
	if !dockey.Key.Equal(doc.Key().Key) {
		return errors.Wrap(ErrDocVerification, fmt.Sprintf("Expected %s, got %s", doc.Key().UUID(), dockey.UUID()))
	}

	// check if DocKey exists in DB
	// write object marker
	err = writeObjectMarker(txn.datastore, c.getDocKey(doc.Key().Instance("v")))
	if err != nil {
		return err
	}
	// write data to DB via MerkleClock/CRDT
	return c.save(txn, doc)
}

// Update an existing document with the new values
// Any field that needs to be removed or cleared
// should call doc.Clear(field) before. Any field that
// is nil/empty that hasn't called Clear will be ignored
func (c *Collection) Update(doc *document.Document) error {
	txn, err := c.getTxn(false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(txn)

	exists, err := c.exists(txn, doc.Key())
	if err != nil {
		return err
	}
	if !exists {
		return ErrDocumentNotFound
	}

	err = c.update(txn, doc)
	if err != nil {
		return err
	}

	return c.commitImplicitTxn(txn)
}

// Contract: DB Exists check is already perfomed, and a doc with the given key exists
// Note: Should we CompareAndSet the update, IE: Query the state, and update if changed
// or, just update everything regardless.
// Should probably be smart about the update due to the MerkleCRDT overhead, shouldn't
// add to the bloat.
func (c *Collection) update(txn *Txn, doc *document.Document) error {
	err := c.save(txn, doc)
	if err != nil {
		return err
	}
	return nil
}

// Save a document into the db
// Either by creating a new document or by updating an existing one
func (c *Collection) Save(doc *document.Document) error {
	txn, err := c.getTxn(false)
	if err != nil {
		return err
	}
	defer c.discardImplicitTxn(txn)

	// Check if document already exists with key
	exists, err := c.exists(txn, doc.Key())
	if err != nil {
		return err
	}

	if exists {
		err = c.update(txn, doc)
	} else {
		err = c.create(txn, doc)
	}
	if err != nil {
		return err
	}
	return c.commitImplicitTxn(txn)
}

func (c *Collection) save(txn *Txn, doc *document.Document) error {
	// New batch transaction/store (optional/todo)
	// Ensute/Set doc object marker
	// Loop through doc values
	//	=> 		instanciate MerkleCRDT objects
	//	=> 		Set/Publish new CRDT values
	for k, v := range doc.Fields() {
		val, _ := doc.GetValueWithField(v)
		if val.IsDirty() {
			err := c.saveValueToMerkleCRDT(txn, c.getDocKey(doc.Key().ChildString(k)), val)
			if err != nil {
				return err
			}
			if val.IsDelete() {
				doc.SetAs(v.Name(), nil, v.Type())
			}
			// set value as clean
			val.Clean()
		}
	}

	return nil
}

// Delete will attempt to delete a document by key
// will return true if a deltion is successful, and
// return false, along with an error, if it cannot.
// If the document doesn't exist, then it will return
// false, and a ErrDocumentNotFound error.
// This operation will all state relating to the given
// DocKey. This includes data, block, and head storage.
func (c *Collection) Delete(key key.DocKey) (bool, error) {
	// create txn
	txn, err := c.getTxn(false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(txn)

	exists, err := c.exists(txn, key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrDocumentNotFound
	}

	// run delete, commit if successful
	deleted, err := c.delete(txn, key)
	if err != nil {
		return false, err
	}
	return deleted, c.commitImplicitTxn(txn)
}

// at the moment, delete only does data storage delete.
// Dag, and head store will soon follow.
func (c *Collection) delete(txn *Txn, key key.DocKey) (bool, error) {
	q := query.Query{
		Prefix:   c.getDocKey(key.Key).String(),
		KeysOnly: true,
	}
	res, err := txn.datastore.Query(q)

	for e := range res.Next() {
		if e.Error != nil {
			return false, err
		}

		err = txn.datastore.Delete(c.getDocKey(ds.NewKey(e.Key)))
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// Exists checks if a given document exists with supplied DocKey
func (c *Collection) Exists(key key.DocKey) (bool, error) {
	// create txn
	txn, err := c.getTxn(false)
	if err != nil {
		return false, err
	}
	defer c.discardImplicitTxn(txn)

	exists, err := c.exists(txn, key)
	if err != nil && err != ds.ErrNotFound {
		return false, err
	}
	return exists, c.commitImplicitTxn(txn)
}

// check if a document exists with the given key
func (c *Collection) exists(txn *Txn, key key.DocKey) (bool, error) {
	return txn.datastore.Has(c.getDocKey(key.Key.Instance("v")))
}

func (c *Collection) saveValueToMerkleCRDT(txn *Txn, key ds.Key, val document.Value) error {
	switch val.Type() {
	case core.LWW_REGISTER:
		wval, ok := val.(document.WriteableValue)
		if !ok {
			return document.ErrValueTypeMismatch
		}
		datatype, err := c.db.crdtFactory.InstanceWithStores(txn, core.LWW_REGISTER, key)
		if err != nil {
			return err
		}
		lwwreg := datatype.(*crdt.MerkleLWWRegister)
		var bytes []byte
		if val.IsDelete() { // empty byte array
			bytes = []byte{}
		} else {
			bytes, err = wval.Bytes()
			if err != nil {
				return err
			}
		}
		return lwwreg.Set(bytes)

	case core.OBJECT:
		// db.writeObjectMarker(db.datastore, subdoc.Instance("v"))
		c.db.log.Debug("Sub objects not yet supported")
		break
	}
	return nil
}

// getTxn gets or creates a new transaction from the underlying db.
// If the collection already has a txn, return the existing one.
// Otherwise, create a new implicit transaction.
func (c *Collection) getTxn(readonly bool) (*Txn, error) {
	if c.txn != nil {
		return c.txn, nil
	}
	return c.db.newTxn(readonly)
}

// discardImplicitTxn is a proxy function used by the collection to execute the Discard() transaction
// function only if its an implicit transaction.
// Implicit transactions are transactions that are created *during* an operation execution as a side effect.
// Explicit transactions are provided to the collection object via the "WithTxn(...)" function.
func (c *Collection) discardImplicitTxn(txn *Txn) {
	if c.txn == nil {
		txn.Discard()
	}
}

func (c *Collection) commitImplicitTxn(txn *Txn) error {
	if c.txn == nil {
		return txn.Commit()
	}
	return nil
}

func (c *Collection) getDocKey(key ds.Key) ds.Key {
	return c.colIDKey.Child(key)
}

func writeObjectMarker(store ds.Write, key ds.Key) error {
	if key.Name() != "v" {
		key = key.Instance("v")
	}
	return store.Put(key, []byte{objectMarker})
}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
func makeCollectionSystemKey(name string) ds.Key {
	return collectionNs.ChildString(name)
}

// makeCollectionKey returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
// func makeCollectionDataKey(collectionID uint32) ds.Key {
// 	return collectionNs.ChildString(name)
// }
